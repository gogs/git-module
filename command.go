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

// CommandOptions contains options for running a command.
type CommandOptions struct {
	Args []string
	Envs []string
}

// DefaultTimeout is the default timeout duration for all commands. It is
// applied when the context does not already have a deadline.
const DefaultTimeout = time.Minute

// gitCmd builds a *run.Command for "git" with the given arguments, environment
// variables and working directory. If the context does not already have a
// deadline, DefaultTimeout will be applied automatically.
func gitCmd(ctx context.Context, dir string, args []string, envs []string) *run.Command {
	if ctx == nil {
		ctx = context.Background()
	}

	// Apply default timeout if the context doesn't already have a deadline.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultTimeout)
		_ = cancel
	}

	// run.Cmd joins all parts into a single string and then shell-parses it.
	// We must quote each argument so that special characters (spaces, quotes,
	// angle brackets, etc.) are preserved correctly.
	parts := make([]string, 0, 1+len(args))
	parts = append(parts, "git")
	for _, arg := range args {
		parts = append(parts, run.Arg(arg))
	}

	cmd := run.Cmd(ctx, parts...)
	if dir != "" {
		cmd = cmd.Dir(dir)
	}
	if len(envs) > 0 {
		cmd = cmd.Environ(append(os.Environ(), envs...))
	}
	return cmd
}

// gitRun executes a git command in the given directory and returns stdout as
// bytes. Stderr is included in the error message on failure. If the command's
// context does not have a deadline, DefaultTimeout will be applied
// automatically. It returns an ErrExecTimeout if the execution was timed out.
func gitRun(ctx context.Context, dir string, args []string, envs []string) ([]byte, error) {
	cmd := gitCmd(ctx, dir, args, envs)

	logBuf := new(bytes.Buffer)
	if logOutput != nil {
		logBuf.Grow(512)
	}

	defer func() {
		logf(dir, args, logBuf.Bytes())
	}()

	// Use Stream to a buffer to preserve raw bytes (including NUL bytes from
	// commands like "ls-tree -z"). The String/Lines methods process output
	// line-by-line which corrupts binary-ish output.
	stdout := new(bytes.Buffer)
	err := cmd.StdOut().Run().Stream(stdout)
	if err != nil {
		return nil, mapContextError(err, ctx)
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

// gitPipeline executes a git command in the given directory, streaming stdout
// to the given writer. If stderr writer is provided and the command fails,
// stderr content extracted from the error is written to it. stdin is optional.
func gitPipeline(ctx context.Context, dir string, args []string, envs []string, stdout, stderr io.Writer, stdin io.Reader) error {
	cmd := gitCmd(ctx, dir, args, envs)
	if stdin != nil {
		cmd = cmd.Input(stdin)
	}

	buf := new(bytes.Buffer)
	w := stdout
	if logOutput != nil {
		buf.Grow(512)
		w = &limitDualWriter{
			W: buf,
			N: int64(buf.Cap()),
			w: stdout,
		}
	}

	defer func() {
		logf(dir, args, buf.Bytes())
	}()

	streamErr := cmd.StdOut().Run().Stream(w)
	if streamErr != nil {
		if stderr != nil {
			_, _ = fmt.Fprint(stderr, extractStderr(streamErr))
		}
		return mapContextError(streamErr, ctx)
	}
	return nil
}

// committerEnvs returns environment variables for setting the Git committer.
func committerEnvs(committer *Signature) []string {
	return []string{
		"GIT_COMMITTER_NAME=" + committer.Name,
		"GIT_COMMITTER_EMAIL=" + committer.Email,
	}
}

// logf logs a git command execution with optional output.
func logf(dir string, args []string, output []byte) {
	cmdStr := "git"
	if len(args) > 0 {
		cmdStr = "git " + strings.Join(args, " ")
	}
	if len(dir) == 0 {
		log("%s\n%s", cmdStr, output)
	} else {
		log("%s: %s\n%s", dir, cmdStr, output)
	}
}

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
	// Also check if the error itself wraps a context error.
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
