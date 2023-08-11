// Copyright 2023 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"time"
)

// UpdateServerInfoOptions contains optional arguments for updating auxiliary
// info file on the server side.
//
// Docs: https://git-scm.com/docs/git-update-server-info
type UpdateServerInfoOptions struct {
	// Indicates whether to overwrite the existing server info.
	Force bool
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// UpdateServerInfo updates the auxiliary info file on the server side for the
// repository in given path.
func UpdateServerInfo(path string, opts ...UpdateServerInfoOptions) error {
	var opt UpdateServerInfoOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	cmd := NewCommand("update-server-info").AddOptions(opt.CommandOptions)
	if opt.Force {
		cmd.AddArgs("--force")
	}
	_, err := cmd.RunInDirWithTimeout(opt.Timeout, path)
	return err
}

// ReceivePackOptions contains optional arguments for receiving the info pushed
// to the repository.
//
// Docs: https://git-scm.com/docs/git-receive-pack
type ReceivePackOptions struct {
	// Indicates whether to suppress the log output.
	Quiet bool
	// Indicates whether to generate the "info/refs" used by the "git http-backend".
	HTTPBackendInfoRefs bool
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// ReceivePack receives what is pushed into the repository in given path.
func ReceivePack(path string, opts ...ReceivePackOptions) ([]byte, error) {
	var opt ReceivePackOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	cmd := NewCommand("receive-pack").AddOptions(opt.CommandOptions)
	if opt.Quiet {
		cmd.AddArgs("--quiet")
	}
	if opt.HTTPBackendInfoRefs {
		cmd.AddArgs("--http-backend-info-refs")
	}
	cmd.AddArgs(".")
	return cmd.RunInDirWithTimeout(opt.Timeout, path)
}

// UploadPackOptions contains optional arguments for sending the packfile to the
// client.
//
// Docs: https://git-scm.com/docs/git-upload-pack
type UploadPackOptions struct {
	// Indicates whether to quit after a single request/response exchange.
	StatelessRPC bool
	// Indicates whether to not try "<directory>/.git/" if "<directory>" is not a
	// Git directory.
	Strict bool
	// Indicates whether to generate the "info/refs" used by the "git http-backend".
	HTTPBackendInfoRefs bool
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// UploadPack sends the packfile to the client for the repository in given path.
func UploadPack(path string, opts ...UploadPackOptions) ([]byte, error) {
	var opt UploadPackOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	cmd := NewCommand("upload-pack").AddOptions(opt.CommandOptions)
	if opt.StatelessRPC {
		cmd.AddArgs("--stateless-rpc")
	}
	if opt.Strict {
		cmd.AddArgs("--strict")
	}
	if opt.Timeout > 0 {
		cmd.AddArgs("--timeout", opt.Timeout.String())
	}
	if opt.HTTPBackendInfoRefs {
		cmd.AddArgs("--http-backend-info-refs")
	}
	cmd.AddArgs(".")
	return cmd.RunInDirWithTimeout(opt.Timeout, path)
}
