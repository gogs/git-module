// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"fmt"
	"strings"
)

const BranchPrefix = "refs/heads/"

// IsReferenceExist returns true if given reference exists in the repository.
func IsReferenceExist(repoPath, name string) bool {
	_, err := NewCommand("show-ref", "--verify", name).RunInDir(repoPath)
	return err == nil
}

// IsBranchExist returns true if given branch exists in the repository.
func IsBranchExist(repoPath, name string) bool {
	return IsReferenceExist(repoPath, BranchPrefix+name)
}

func (r *Repository) IsBranchExist(name string) bool {
	return IsBranchExist(r.path, name)
}

// Branch represents a Git branch.
type Branch struct {
	Name string
	Path string
}

// GetHEADBranch returns corresponding branch of HEAD.
func (r *Repository) GetHEADBranch() (*Branch, error) {
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
