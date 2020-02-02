// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

const BranchPrefix = "refs/heads/"

// VerifyRefOptions contains optional arguments for verifying a reference.
type VerifyRefOptions struct {
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// VerifyRef returns true if given reference exists in the repository.
func (r *Repository) VerifyRef(ref string, opts ...VerifyRefOptions) bool {
	var opt VerifyRefOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	_, err := NewCommand("show-ref", "--verify", ref).RunInDirWithTimeout(opt.Timeout, r.path)
	return err == nil
}

// HasBranch returns true if given branch exists in the repository.
func (r *Repository) HasBranch(name string) bool {
	return r.VerifyRef(name)
}

// Branch contains information of a Git branch.
type Branch struct {
	Name string
	Path string
}

// SymbolicRef returns the current branch of HEAD.
func (r *Repository) SymbolicRef() (*Branch, error) {
	stdout, err := NewCommand("symbolic-ref", "HEAD").RunInDir(r.path)
	if err != nil {
		return nil, err
	}
	ref := strings.TrimSpace(string(stdout))

	if !strings.HasPrefix(ref, BranchPrefix) {
		return nil, fmt.Errorf("invalid HEAD branch: %v", stdout)
	}

	return &Branch{
		Name: ref[len(BranchPrefix):],
		Path: ref,
	}, nil
}

// SetDefaultBranch sets default branch of repository.
func (r *Repository) SetDefaultBranch(name string) error {
	_, err := NewCommand("symbolic-ref", "HEAD", BranchPrefix+name).RunInDir(r.path)
	return err
}

// GetBranches returns all branches of the repository.
func (r *Repository) GetBranches() ([]string, error) {
	stdout, err := NewCommand("show-ref", "--heads").RunInDir(r.path)
	if err != nil {
		return nil, err
	}

	infos := strings.Split(string(stdout), "\n")
	branches := make([]string, len(infos)-1)
	for i, info := range infos[:len(infos)-1] {
		fields := strings.Fields(info)
		if len(fields) != 2 {
			continue // NOTE: I should believe git will not give me wrong string.
		}
		branches[i] = strings.TrimPrefix(fields[1], BranchPrefix)
	}
	return branches, nil
}

// Option(s) for delete branch
type DeleteBranchOptions struct {
	Force bool
}

// DeleteBranch deletes a branch from given repository path.
func DeleteBranch(repoPath, name string, opts DeleteBranchOptions) error {
	cmd := NewCommand("branch")

	if opts.Force {
		cmd.AddArgs("-D")
	} else {
		cmd.AddArgs("-d")
	}

	cmd.AddArgs(name)
	_, err := cmd.RunInDir(repoPath)

	return err
}

// DeleteBranch deletes a branch from repository.
func (r *Repository) DeleteBranch(name string, opts DeleteBranchOptions) error {
	return DeleteBranch(r.path, name, opts)
}

// Reference contains information of a Git reference.
type Reference struct {
	ID   string
	Name string
}

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
			ID:   string(fields[0]),
			Name: string(fields[1]),
		})
	}
	return refs, nil
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
