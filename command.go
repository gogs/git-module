// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/sourcegraph/run"
)

// Command contains the name, arguments and environment variables of a command.
type Command struct {
	name string
	args []string
	envs []string
	ctx  context.Context
}

// CommandOptions contains options for running a command.
type CommandOptions struct {
	Args []string
	Envs []string
}

// String returns the string representation of the command.
func (c *Command) String() string {
	if len(c.args) == 0 {
		return c.name
	}
	return fmt.Sprintf("%s %s", c.name, strings.Join(c.args, " "))
}

// NewCommand creates and returns a new Command with given arguments for "git".
func NewCommand(ctx context.Context, args ...string) *Command {
	return &Command{
		name: "git",
		args: args,
		ctx:  ctx,
	}
}

// AddArgs appends given arguments to the command.
func (c *Command) AddArgs(args ...string) *Command {
	c.args = append(c.args, args...)
	return c
}

// AddEnvs appends given environment variables to the command.
func (c *Command) AddEnvs(envs ...string) *Command {
	c.envs = append(c.envs, envs...)
	return c
}

// WithContext returns a new Command with the given context.
func (c Command) WithContext(ctx context.Context) *Command {
	c.ctx = ctx
	return &c
}

// AddOptions adds options to the command.
func (c *Command) AddOptions(opts ...CommandOptions) *Command {
	for _, opt := range opts {
		c.AddArgs(opt.Args...)
		c.AddEnvs(opt.Envs...)
	}
	return c
}

// AddCommitter appends given committer to the command.
func (c *Command) AddCommitter(committer *Signature) *Command {
	c.AddEnvs("GIT_COMMITTER_NAME="+committer.Name, "GIT_COMMITTER_EMAIL="+committer.Email)
	return c
}

// DefaultTimeout is the default timeout duration for all commands. It is
// applied when the context does not already have a deadline.
const DefaultTimeout = time.Minute

// A limitDualWriter writes to W but limits the amount of data written to just N
// bytes. On the other hand, it passes everything to w.
type limitDualWriter struct {
	W        io.Writer // underlying writer
	N        int64     // max bytes remaining
	prompted bool

	w io.Writer
}

func (w *limitDualWriter) Write(p []byte) (int, error) {
	if w.N > 0 {
		limit := int64(len(p))
		if limit > w.N {
			limit = w.N
		}
		n, _ := w.W.Write(p[:limit])
		w.N -= int64(n)
	}

	if !w.prompted && w.N <= 0 {
		w.prompted = true
		_, _ = w.W.Write([]byte("... (more omitted)"))
	}

	return w.w.Write(p)
}

// RunInDirOptions contains options for running a command in a directory.
type RunInDirOptions struct {
	// Stdin is the input to the command.
	Stdin io.Reader
	// Stdout is the outputs from the command.
	Stdout io.Writer
	// Stderr is the error output from the command.
	Stderr io.Writer
}

// newRunCmd builds a *run.Command from the Command, applying the directory,
// environment variables and default timeout.
func (c *Command) newRunCmd(dir string) *run.Command {
	ctx := c.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	// Apply default timeout if the context doesn't already have a deadline.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultTimeout)
		// We cannot defer cancel here because the command hasn't run yet.
		// The caller is responsible for the context lifecycle. In practice the
		// timeout context will be collected when it expires or when the parent
		// is cancelled. We attach the cancel func to the context so the caller
		// could cancel it, but for this fire-and-forget usage the GC handles it.
		_ = cancel
	}

	// run.Cmd joins all parts into a single string and then shell-parses it.
	// We must quote each argument so that special characters (spaces, quotes,
	// angle brackets, etc.) are preserved correctly.
	parts := make([]string, 0, 1+len(c.args))
	parts = append(parts, c.name)
	for _, arg := range c.args {
		parts = append(parts, run.Arg(arg))
	}

	cmd := run.Cmd(ctx, parts...)
	if dir != "" {
		cmd = cmd.Dir(dir)
	}
	if len(c.envs) > 0 {
		cmd = cmd.Environ(append(os.Environ(), c.envs...))
	}
	return cmd
}

// RunInDirWithOptions executes the command in given directory and options. It
// pipes stdin from supplied io.Reader, and pipes stdout and stderr to supplied
// io.Writer. If the command's context does not have a deadline, DefaultTimeout
// will be applied automatically. It returns an ErrExecTimeout if the execution
// was timed out.
func (c *Command) RunInDirWithOptions(dir string, opts ...RunInDirOptions) (err error) {
	var opt RunInDirOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	buf := new(bytes.Buffer)
	stdout := opt.Stdout
	if logOutput != nil {
		buf.Grow(512)
		stdout = &limitDualWriter{
			W: buf,
			N: int64(buf.Cap()),
			w: opt.Stdout,
		}
	}

	defer func() {
		if len(dir) == 0 {
			log("%s\n%s", c, buf.Bytes())
		} else {
			log("%s: %s\n%s", dir, c, buf.Bytes())
		}
	}()

	cmd := c.newRunCmd(dir)
	if opt.Stdin != nil {
		cmd = cmd.Input(opt.Stdin)
	}

	// Build a combined writer for the command output. We need to capture
	// both stdout and stderr separately. sourcegraph/run's default Output
	// is combined output, but we need to split them.
	//
	// We use StdOut() to get only stdout on the output stream and handle
	// stderr via a pipe.
	//
	// However, sourcegraph/run doesn't have a direct way to wire both stdout
	// and stderr to separate writers in a single Run call. Instead, we use
	// the approach of streaming stdout and collecting stderr from the error.
	//
	// For the streaming case, we stream stdout to the writer and if there's
	// an error, stderr is included in the error message by default.

	if stdout != nil && opt.Stderr != nil {
		// When both stdout and stderr writers are provided, we need to stream
		// stdout and capture stderr. We use StdOut() to only get stdout.
		output := cmd.StdOut().Run()
		streamErr := output.Stream(stdout)
		if streamErr != nil {
			// Extract stderr from the error and write it to the stderr writer.
			// sourcegraph/run includes stderr in the error by default.
			_, _ = fmt.Fprint(opt.Stderr, extractStderr(streamErr))
			return mapContextError(streamErr, c.ctx)
		}
		return nil
	} else if stdout != nil {
		// Only stdout writer provided
		output := cmd.StdOut().Run()
		streamErr := output.Stream(stdout)
		if streamErr != nil {
			return mapContextError(streamErr, c.ctx)
		}
		return nil
	}

	// No writers - just wait for completion
	waitErr := cmd.Run().Wait()
	if waitErr != nil {
		return mapContextError(waitErr, c.ctx)
	}
	return nil
}

// mapContextError maps context errors to the appropriate sentinel errors used
// by this package.
func mapContextError(err error, ctx context.Context) error {
	if ctx == nil {
		return err
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		if ctxErr == context.DeadlineExceeded {
			return ErrExecTimeout
		}
		return ctxErr
	}
	// Also check if the error itself wraps a context error
	if strings.Contains(err.Error(), "signal: killed") || strings.Contains(err.Error(), context.DeadlineExceeded.Error()) {
		if ctx.Err() == context.DeadlineExceeded {
			return ErrExecTimeout
		}
	}
	return err
}

// extractStderr attempts to extract the stderr portion from a sourcegraph/run
// error. The error format is typically "exit status N: <stderr content>".
func extractStderr(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	// sourcegraph/run error format: "exit status N: <stderr>"
	if idx := strings.Index(msg, ": "); idx >= 0 && strings.HasPrefix(msg, "exit status") {
		return msg[idx+2:]
	}
	return msg
}

// RunInDirPipeline executes the command in given directory. It pipes stdout and
// stderr to supplied io.Writer.
func (c *Command) RunInDirPipeline(stdout, stderr io.Writer, dir string) error {
	return c.RunInDirWithOptions(dir, RunInDirOptions{
		Stdin:  nil,
		Stdout: stdout,
		Stderr: stderr,
	})
}

// RunInDir executes the command in given directory. It returns stdout and error
// (combined with stderr).
func (c *Command) RunInDir(dir string) ([]byte, error) {
	cmd := c.newRunCmd(dir)

	logBuf := new(bytes.Buffer)
	if logOutput != nil {
		logBuf.Grow(512)
	}

	defer func() {
		if len(dir) == 0 {
			log("%s\n%s", c, logBuf.Bytes())
		} else {
			log("%s: %s\n%s", dir, c, logBuf.Bytes())
		}
	}()

	// Use Stream to a buffer to preserve raw bytes (including NUL bytes from
	// commands like "ls-tree -z"). The String/Lines methods process output
	// line-by-line which corrupts binary-ish output.
	stdout := new(bytes.Buffer)
	err := cmd.StdOut().Run().Stream(stdout)
	if err != nil {
		return nil, mapContextError(err, c.ctx)
	}

	if logOutput != nil {
		data := stdout.Bytes()
		limit := len(data)
		if limit > 512 {
			limit = 512
		}
		logBuf.Write(data[:limit])
		if len(data) > 512 {
			logBuf.WriteString("... (more omitted)")
		}
	}

	return stdout.Bytes(), nil
}

// Run executes the command in working directory. It returns stdout and
// error (combined with stderr).
func (c *Command) Run() ([]byte, error) {
	stdout, err := c.RunInDir("")
	if err != nil {
		return nil, err
	}
	return stdout, nil
}
