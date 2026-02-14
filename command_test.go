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

func TestCommand_String(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name   string
		args   []string
		expStr string
	}{
		{
			name:   "no args",
			args:   nil,
			expStr: "git",
		},
		{
			name:   "has one arg",
			args:   []string{"version"},
			expStr: "git version",
		},
		{
			name:   "has more args",
			args:   []string{"config", "--global", "http.proxy", "http://localhost:8080"},
			expStr: "git config --global http.proxy http://localhost:8080",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := NewCommand(ctx, test.args...)
			assert.Equal(t, test.expStr, cmd.String())
		})
	}
}

func TestCommand_AddArgs(t *testing.T) {
	ctx := context.Background()
	cmd := NewCommand(ctx)
	assert.Equal(t, []string(nil), cmd.args)

	cmd.AddArgs("push")
	cmd.AddArgs("origin", "master")
	assert.Equal(t, []string{"push", "origin", "master"}, cmd.args)
}

func TestCommand_AddEnvs(t *testing.T) {
	ctx := context.Background()
	cmd := NewCommand(ctx)
	assert.Equal(t, []string(nil), cmd.envs)

	cmd.AddEnvs("GIT_DIR=/tmp")
	cmd.AddEnvs("HOME=/Users/unknwon", "GIT_EDITOR=code")
	assert.Equal(t, []string{"GIT_DIR=/tmp", "HOME=/Users/unknwon", "GIT_EDITOR=code"}, cmd.envs)
}

func TestCommand_RunWithContextTimeout(t *testing.T) {
	t.Run("context already expired before start", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		time.Sleep(time.Millisecond) // ensure deadline has passed
		_, err := NewCommand(ctx, "version").Run()
		assert.Equal(t, ErrExecTimeout, err)
	})

	t.Run("context deadline fires mid-execution", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Use a blocking reader so the command starts successfully and blocks
		// reading stdin until the context deadline fires.
		err := NewCommand(ctx, "hash-object", "--stdin").RunInDirWithOptions("", RunInDirOptions{
			Stdin:  blockingReader{cancel: ctx.Done()},
			Stdout: io.Discard,
			Stderr: io.Discard,
		})
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

func TestCommand_RunWithContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel in the background after a short delay so the command is already
	// running when cancellation arrives. Close done to unblock the reader.
	done := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
		close(done)
	}()

	err := NewCommand(ctx, "hash-object", "--stdin").RunInDirWithOptions("", RunInDirOptions{
		Stdin:  blockingReader{cancel: done},
		Stdout: io.Discard,
		Stderr: io.Discard,
	})
	assert.ErrorIs(t, err, context.Canceled)
	// Must NOT be ErrExecTimeout — cancellation is distinct from deadline.
	assert.NotEqual(t, ErrExecTimeout, err)
}

func TestCommand_DefaultTimeoutApplied(t *testing.T) {
	// A plain context.Background() has no deadline. The command should still
	// succeed because DefaultTimeout (1 min) is applied automatically and
	// "git version" completes well within that.
	ctx := context.Background()
	stdout, err := NewCommand(ctx, "version").Run()
	assert.NoError(t, err)
	assert.Contains(t, string(stdout), "git version")
}
