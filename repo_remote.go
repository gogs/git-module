// Copyright 2019 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"strings"
	"time"
)

// LsRemoteOptions contains arguments for listing references in a remote repository.
// Docs: https://git-scm.com/docs/git-ls-remote
type LsRemoteOptions struct {
	// Indicates whether include heads.
	Heads bool
	// Indicates whether include tags.
	Tags bool
	// Indicates whether to not show peeled tags or pseudorefs.
	Refs bool
	// The list of patterns to filter results.
	Patterns []string
	// The timeout duration before giving up for each shell command execution.
	// The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// LsRemote returns a list references in the remote repository.
func LsRemote(url string, opts ...LsRemoteOptions) ([]*Reference, error) {
	var opt LsRemoteOptions
	if len(opts) > 0 {
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
	cmd.AddArgs(url)
	if len(opt.Patterns) > 0 {
		cmd.AddArgs(opt.Patterns...)
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

// IsURLAccessible returns true if given remote URL is accessible via Git
// within given timeout.
func IsURLAccessible(timeout time.Duration, url string) bool {
	_, err := LsRemote(url, LsRemoteOptions{
		Patterns: []string{"HEAD"},
		Timeout:  timeout,
	})
	return err == nil
}

// AddRemoteOptions contains options to add a remote address.
// Docs: https://git-scm.com/docs/git-remote#Documentation/git-remote.txt-emaddem
type AddRemoteOptions struct {
	// Indicates whether to execute git fetch after the remote information is set up.
	Fetch bool
	// Indicates whether to add remote as mirror with --mirror=fetch.
	MirrorFetch bool
	// The timeout duration before giving up for each shell command execution.
	// The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// AddRemote adds a new remote to the repository in given path.
func RepoAddRemote(repoPath, name, url string, opts ...AddRemoteOptions) error {
	var opt AddRemoteOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("remote", "add")
	if opt.Fetch {
		cmd.AddArgs("-f")
	}
	if opt.MirrorFetch {
		cmd.AddArgs("--mirror=fetch")
	}

	_, err := cmd.AddArgs(name, url).RunInDirWithTimeout(opt.Timeout, repoPath)
	return err
}

// AddRemote adds a new remote to the repository.
func (r *Repository) AddRemote(name, url string, opts ...AddRemoteOptions) error {
	return RepoAddRemote(r.path, name, url, opts...)
}

// RemoveRemoteOptions contains arguments for removing a remote from the repository.
// Docs: https://git-scm.com/docs/git-remote#Documentation/git-remote.txt-emremoveem
type RemoveRemoteOptions struct {
	// The timeout duration before giving up for each shell command execution.
	// The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// RepoRemoveRemote removes a remote from the repository in given path.
func RepoRemoveRemote(repoPath, name string, opts ...RemoveRemoteOptions) error {
	var opt RemoveRemoteOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	_, err := NewCommand("remote", "remove", name).RunInDirWithTimeout(opt.Timeout, repoPath)
	if err != nil {
		if strings.Contains(err.Error(), "fatal: No such remote") {
			return ErrRemoteNotExist
		}
		return err
	}
	return nil
}

// RemoveRemote removes a remote from the repository.
func (r *Repository) RemoveRemote(name string, opts ...RemoveRemoteOptions) error {
	return RepoRemoveRemote(r.path, name, opts...)
}
