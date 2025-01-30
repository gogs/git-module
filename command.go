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
	"os/exec"
	"strings"
	"time"
)

// Command contains the name, arguments and environment variables of a command.
type Command struct {
	name    string
	args    []string
	envs    []string
	timeout time.Duration
	ctx     context.Context
}

// CommandOptions contains options for running a command.
// If timeout is zero, DefaultTimeout will be used.
// If timeout is less than zero, no timeout will be set.
// If context is nil, context.Background() will be used.
type CommandOptions struct {
	Args    []string
	Envs    []string
	Timeout time.Duration
	Context context.Context
}

// String returns the string representation of the command.
func (c *Command) String() string {
	if len(c.args) == 0 {
		return c.name
	}
	return fmt.Sprintf("%s %s", c.name, strings.Join(c.args, " "))
}

// NewCommand creates and returns a new Command with given arguments for "git".
func NewCommand(args ...string) *Command {
	return NewCommandWithContext(context.Background(), args...)
}

// NewCommandWithContext creates and returns a new Command with given arguments
// and context for "git".
func NewCommandWithContext(ctx context.Context, args ...string) *Command {
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
func (c *Command) WithContext(ctx context.Context) *Command {
	c.ctx = ctx
	return c
}

// WithTimeout returns a new Command with given timeout.
func (c *Command) WithTimeout(timeout time.Duration) *Command {
	c.timeout = timeout
	return c
}

// SetTimeout sets the timeout for the command.
func (c *Command) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// AddOptions adds options to the command.
// Note: only the last option will take effect if there are duplicated options.
func (c *Command) AddOptions(opts ...CommandOptions) *Command {
	for _, opt := range opts {
		c.timeout = opt.Timeout
		c.ctx = opt.Context
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

// DefaultTimeout is the default timeout duration for all commands.
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
	// Timeout is the duration to wait before timing out.
	//
	// Deprecated: Use CommandOptions.Timeout or *Command.WithTimeout instead.
	Timeout time.Duration
}

// RunInDirWithOptions executes the command in given directory and options. It
// pipes stdin from supplied io.Reader, and pipes stdout and stderr to supplied
// io.Writer. DefaultTimeout will be used if the timeout duration is less than
// time.Nanosecond (i.e. less than or equal to 0). It returns an ErrExecTimeout
// if the execution was timed out.
func (c *Command) RunInDirWithOptions(dir string, opts ...RunInDirOptions) (err error) {
	var opt RunInDirOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	timeout := c.timeout
	// TODO: remove this in newer version
	if opt.Timeout > 0 {
		timeout = opt.Timeout
	}

	if timeout == 0 {
		timeout = DefaultTimeout
	}

	buf := new(bytes.Buffer)
	w := opt.Stdout
	if logOutput != nil {
		buf.Grow(512)
		w = &limitDualWriter{
			W: buf,
			N: int64(buf.Cap()),
			w: opt.Stdout,
		}
	}

	defer func() {
		if len(dir) == 0 {
			log("[timeout: %v] %s\n%s", timeout, c, buf.Bytes())
		} else {
			log("[timeout: %v] %s: %s\n%s", timeout, dir, c, buf.Bytes())
		}
	}()

	ctx := context.Background()
	if c.ctx != nil {
		ctx = c.ctx
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer func() {
			cancel()
			if err == context.DeadlineExceeded {
				err = ErrExecTimeout
			}
		}()
	}

	cmd := exec.CommandContext(ctx, c.name, c.args...)
	if len(c.envs) > 0 {
		cmd.Env = append(os.Environ(), c.envs...)
	}
	cmd.Dir = dir
	cmd.Stdin = opt.Stdin
	cmd.Stdout = w
	cmd.Stderr = opt.Stderr
	if err = cmd.Start(); err != nil {
		return err
	}

	result := make(chan error)
	go func() {
		result <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		<-result
		if cmd.Process != nil && cmd.ProcessState != nil && !cmd.ProcessState.Exited() {
			if err := cmd.Process.Kill(); err != nil {
				return fmt.Errorf("kill process: %v", err)
			}
		}

		return ErrExecTimeout
	case err = <-result:
		return err
	}

}

// RunInDirPipeline executes the command in given directory and default timeout
// duration. It pipes stdout and stderr to supplied io.Writer.
func (c *Command) RunInDirPipeline(stdout, stderr io.Writer, dir string) error {
	return c.RunInDirWithOptions(dir, RunInDirOptions{
		Stdin:  nil,
		Stdout: stdout,
		Stderr: stderr,
	})
}

// RunInDirPipelineWithTimeout executes the command in given directory and
// timeout duration. It pipes stdout and stderr to supplied io.Writer.
// DefaultTimeout will be used if the timeout duration is less than
// time.Nanosecond (i.e. less than or equal to 0). It returns an ErrExecTimeout
// if the execution was timed out.
//
// Deprecated: Use RunInDirPipeline and CommandOptions instead.
// TODO: remove this in the next major version
func (c *Command) RunInDirPipelineWithTimeout(timeout time.Duration, stdout, stderr io.Writer, dir string) (err error) {
	if timeout != 0 {
		c = c.WithTimeout(timeout)
	}
	return c.RunInDirPipeline(stdout, stderr, dir)
}

// RunInDirWithTimeout executes the command in given directory and timeout
// duration. It returns stdout in []byte and error (combined with stderr).
//
// Deprecated: Use RunInDir and CommandOptions instead.
// TODO: remove this in the next major version
func (c *Command) RunInDirWithTimeout(timeout time.Duration, dir string) ([]byte, error) {
	if timeout != 0 {
		c = c.WithTimeout(timeout)
	}
	return c.RunInDir(dir)
}

// RunInDir executes the command in given directory and default timeout
// duration. It returns stdout and error (combined with stderr).
func (c *Command) RunInDir(dir string) ([]byte, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	if err := c.RunInDirPipeline(stdout, stderr, dir); err != nil {
		return nil, concatenateError(err, stderr.String())
	}
	return stdout.Bytes(), nil
}

// RunWithTimeout executes the command in working directory and given timeout
// duration. It returns stdout in string and error (combined with stderr).
//
// Deprecated: Use RunInDir and CommandOptions instead.
// TODO: remove this in the next major version
func (c *Command) RunWithTimeout(timeout time.Duration) ([]byte, error) {
	if timeout != 0 {
		c = c.WithTimeout(timeout)
	}
	return c.Run()
}

// Run executes the command in working directory and default timeout duration.
// It returns stdout in string and error (combined with stderr).
func (c *Command) Run() ([]byte, error) {
	stdout, err := c.RunInDir("")
	if err != nil {
		return nil, err
	}
	return stdout, nil
}
