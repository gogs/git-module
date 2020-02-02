// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	RefsHeads = "refs/heads/"
	RefsTags  = "refs/tags/"
)

// Reference contains information of a Git reference.
type Reference struct {
	ID      string
	Refspec string
}

// ShowRefVerifyOptions contains optional arguments for verifying a reference.
// Docs: https://git-scm.com/docs/git-show-ref#Documentation/git-show-ref.txt---verify
type ShowRefVerifyOptions struct {
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

var ErrReferenceNotExist = errors.New("reference does not exist")

// ShowRefVerify returns the commit ID of given reference if it exists in the repository.
func (r *Repository) ShowRefVerify(ref string, opts ...ShowRefVerifyOptions) (string, error) {
	var opt ShowRefVerifyOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	stdout, err := NewCommand("show-ref", "--verify", ref).RunInDirWithTimeout(opt.Timeout, r.path)
	if err != nil {
		if strings.Contains(err.Error(), "not a valid ref") {
			return "", ErrReferenceNotExist
		}
		return "", err
	}
	return strings.Split(string(stdout), " ")[0], nil
}

// HasBranch returns true if given branch exists in the repository.
func (r *Repository) HasReference(ref string) bool {
	_, err := r.ShowRefVerify(ref)
	return err == nil
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

	if !strings.HasPrefix(ref, RefsHeads) {
		return nil, fmt.Errorf("invalid HEAD branch: %v", stdout)
	}

	return &Branch{
		Name: ref[len(RefsHeads):],
		Path: ref,
	}, nil
}

// SetDefaultBranch sets default branch of repository.
func (r *Repository) SetDefaultBranch(name string) error {
	_, err := NewCommand("symbolic-ref", "HEAD", RefsHeads+name).RunInDir(r.path)
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
		branches[i] = strings.TrimPrefix(fields[1], RefsHeads)
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
