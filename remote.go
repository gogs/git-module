// Copyright 2019 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import "strings"

// RemoveRemote removes a remote from given repository path if exists.
func RemoveRemote(repoPath, remote string) error {
	_, err := NewCommand("remote", "rm", remote).RunInDir(repoPath)
	if err != nil && !strings.Contains(err.Error(), "fatal: No such remote") {
		return err
	}
	return nil
}

// AddRemoteOptions contains options to add a remote address.
type AddRemoteOptions struct {
	Mirror bool
}

// AddRemote adds a new remote
func AddRemote(repoPath, remote, addr string, opts AddRemoteOptions) error {
	cmd := NewCommand("remote", "add", remote)
	if opts.Mirror {
		cmd.AddArguments("--mirror")
	}
	_, err := cmd.AddArguments(addr).RunInDir(repoPath)
	return err
}
