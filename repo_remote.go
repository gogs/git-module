// Copyright 2019 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"strings"
	"time"
)

const RefsRemotes = "refs/remotes/"

// LsRemoteOptions contains arguments for listing references in a remote repository.
// Docs: https://git-scm.com/docs/git-ls-remote
type LsRemoteOptions struct {
	// Indicates whether to only show heads.
	Heads bool
	// Indicates whether to only show tags.
	Tags bool
	// Indicates whether to not show peeled tags or pseudorefs.
	Refs bool
	// The URL of the remote repository.
	URL string
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// LsRemote returns a list references in the remote repository.
func LsRemote(opts ...LsRemoteOptions) ([]*Reference, error) {
	var opt LsRemoteOptions
	if len(opts) > 1 {
		opt = opts[0]
	}

	cmd := NewCommand("ls-remote", "--quiet")
	if opt.Heads {
		cmd.AddArgs("--heads")
	}
	if opt.Tags {
		cmd.AddArgs("--tags")
	}
	if opt.Refs {
		cmd.AddArgs("--refs")
	}
	if opt.URL != "" {
		cmd.AddArgs(opt.URL)
	}

	stdout, err := cmd.RunWithTimeout(opt.Timeout)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(stdout, []byte("\n"))
	refs := make([]*Reference, 0, len(lines))
	for i := range lines {
		fields := bytes.Fields(lines[i])
		if len(fields) < 2 {
			continue
		}

		refs = append(refs, &Reference{
			ID:      string(fields[0]),
			Refspec: string(fields[1]),
		})
	}
	return refs, nil
}

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
		cmd.AddArgs("--mirror")
	}
	_, err := cmd.AddArgs(addr).RunInDir(repoPath)
	return err
}

// AddRemote adds a new remote to repository.
func (r *Repository) AddRemote(name, url string, fetch bool) error {
	cmd := NewCommand("remote", "add")
	if fetch {
		cmd.AddArgs("-f")
	}
	cmd.AddArgs(name, url)

	_, err := cmd.RunInDir(r.path)
	return err
}

// RemoveRemote removes a remote from repository.
func (r *Repository) RemoveRemote(name string) error {
	_, err := NewCommand("remote", "remove", name).RunInDir(r.path)
	return err
}
