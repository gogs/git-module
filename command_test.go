// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCommand_String(t *testing.T) {
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
			cmd := NewCommand(test.args...)
			assert.Equal(t, test.expStr, cmd.String())
		})
	}
}

func TestCommand_AddArgs(t *testing.T) {
	cmd := NewCommand()
	assert.Equal(t, []string(nil), cmd.args)

	cmd.AddArgs("push")
	cmd.AddArgs("origin", "master")
	assert.Equal(t, []string{"push", "origin", "master"}, cmd.args)
}

func TestCommand_AddEnvs(t *testing.T) {
	cmd := NewCommand()
	assert.Equal(t, []string(nil), cmd.envs)

	cmd.AddEnvs("GIT_DIR=/tmp")
	cmd.AddEnvs("HOME=/Users/unknwon", "GIT_EDITOR=code")
	assert.Equal(t, []string{"GIT_DIR=/tmp", "HOME=/Users/unknwon", "GIT_EDITOR=code"}, cmd.envs)
}

func TestCommand_RunWithTimeout(t *testing.T) {
	_, err := NewCommand("version").WithTimeout(time.Nanosecond).Run()
	assert.Equal(t, ErrExecTimeout, err)
}

func TestArgsForVersion(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		version string
		want    []string
		wantErr bool
	}{
		{
			name:    "kept when Git accepts it for every command",
			args:    []string{"checkout", "--end-of-options", "main"},
			version: "2.43.1",
			want:    []string{"checkout", "--end-of-options", "main"},
		},
		{
			name:    "kept on a newer Windows build",
			args:    []string{"reset", "--hard", "--end-of-options", "HEAD"},
			version: "2.55.0.windows.1",
			want:    []string{"reset", "--hard", "--end-of-options", "HEAD"},
		},
		{
			name:    "dropped before 2.43.1",
			args:    []string{"reset", "--hard", "--end-of-options", "HEAD"},
			version: "2.43.0",
			want:    []string{"reset", "--hard", "HEAD"},
		},
		{
			name:    "dropped on Git without it at all",
			args:    []string{"rev-list", "--count", "--end-of-options", "main", "--", "-file"},
			version: "2.1.4",
			want:    []string{"rev-list", "--count", "main", "--", "-file"},
		},
		{
			name:    "option-like positional argument refused before 2.43.1",
			args:    []string{"checkout", "--end-of-options", "--orphan=x"},
			version: "2.43.0",
			wantErr: true,
		},
		{
			name:    "option-like positional argument kept since 2.43.1",
			args:    []string{"checkout", "--end-of-options", "--orphan=x"},
			version: "2.43.1",
			want:    []string{"checkout", "--end-of-options", "--orphan=x"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := argsForVersion(test.args, slices.Index(test.args, endOfOptions), test.version)
			if test.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}
