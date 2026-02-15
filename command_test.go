// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGitRun_ContextTimeout(t *testing.T) {
	t.Run("context already expired before start", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		time.Sleep(time.Millisecond) // ensure deadline has passed
		_, err := gitRun(ctx, "", []string{"version"}, nil)
		assert.Equal(t, ErrExecTimeout, err)
	})

	t.Run("context deadline fires mid-execution", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Use gitCmd directly with a blocking stdin so the command starts
		// successfully and blocks reading until the context deadline fires.
		cmd, timeoutCancel := gitCmd(ctx, "", []string{"hash-object", "--stdin"}, nil)
		defer timeoutCancel()

		err := cmd.Input(blockingReader{cancel: ctx.Done()}).StdOut().Run().Stream(io.Discard)
		err = mapContextError(err, ctx)
		assert.Equal(t, ErrExecTimeout, err)
	})
}

// blockingReader is an io.Reader that blocks until its cancel channel is
// closed, simulating a stdin that never provides data. When cancelled it
// returns io.EOF so that exec's stdin copy goroutine can exit cleanly,
// allowing cmd.Wait() to return.
type blockingReader struct {
	cancel <-chan struct{}
}

func (r blockingReader) Read(p []byte) (int, error) {
	<-r.cancel
	return 0, io.EOF
}

func TestGitCmd_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel in the background after a short delay so the command is already
	// running when cancellation arrives. Close done to unblock the reader.
	done := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
		close(done)
	}()

	cmd, timeoutCancel := gitCmd(ctx, "", []string{"hash-object", "--stdin"}, nil)
	defer timeoutCancel()

	err := cmd.Input(blockingReader{cancel: done}).StdOut().Run().Stream(io.Discard)
	err = mapContextError(err, ctx)
	assert.ErrorIs(t, err, context.Canceled)
	// Must NOT be ErrExecTimeout — cancellation is distinct from deadline.
	assert.NotEqual(t, ErrExecTimeout, err)
}

func TestGitRun_DefaultTimeoutApplied(t *testing.T) {
	// A plain context.Background() has no deadline. The command should still
	// succeed because DefaultTimeout (1 min) is applied automatically and
	// "git version" completes well within that.
	ctx := context.Background()
	stdout, err := gitRun(ctx, "", []string{"version"}, nil)
	assert.NoError(t, err)
	assert.Contains(t, string(stdout), "git version")
}

func TestExtractStderr(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "nil error",
			err:  nil,
			want: "",
		},
		{
			name: "exit status with stderr",
			err:  &exitStatusError{msg: "exit status 1: fatal: not a git repository"},
			want: "fatal: not a git repository",
		},
		{
			name: "other error",
			err:  io.EOF,
			want: "EOF",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, extractStderr(test.err))
		})
	}
}

// exitStatusError is a simple error type for testing extractStderr.
type exitStatusError struct {
	msg string
}

func (e *exitStatusError) Error() string { return e.msg }
