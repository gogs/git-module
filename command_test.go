// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"context"
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	_, err := NewCommand(ctx, "version").Run()
	assert.Equal(t, ErrExecTimeout, err)
}
