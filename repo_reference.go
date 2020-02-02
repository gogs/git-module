// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"errors"
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
func (r *Repository) HasReference(ref string, opts ...ShowRefVerifyOptions) bool {
	_, err := r.ShowRefVerify(ref, opts...)
	return err == nil
}

// SymbolicRefOptions contains optional arguments for get and set symbolic ref.
type SymbolicRefOptions struct {
	// The name of the symbolic ref. When not set, default ref "HEAD" is used.
	Name string
	// The name of the reference. When set, it will be used to update the symbolic ref.
	Ref string
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// SymbolicRef returns the reference name pointed by the symbolic ref. It returns an empty string
// and nil error when doing set operation.
func (r *Repository) SymbolicRef(opts ...SymbolicRefOptions) (string, error) {
	var opt SymbolicRefOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("symbolic-ref")
	if opt.Name == "" {
		opt.Name = "HEAD"
	}
	cmd.AddArgs(opt.Name)
	if opt.Ref != "" {
		cmd.AddArgs(opt.Ref)
	}

	stdout, err := cmd.RunInDirWithTimeout(opt.Timeout, r.path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(stdout)), nil
}

// ShowRefOptions contains optional arguments for listing references.
// Docs: https://git-scm.com/docs/git-show-ref
type ShowRefOptions struct {
	// Indicates whether to only show heads.
	Heads bool
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// ShowRef returns a list of references in the repository.
func (r *Repository) ShowRef(opts ...ShowRefOptions) ([]string, error) {
	var opt ShowRefOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("show-ref")
	if opt.Heads {
		cmd.AddArgs("--heads")
	}

	stdout, err := cmd.RunInDirWithTimeout(opt.Timeout, r.path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(stdout), "\n")
	refs := make([]string, 0, len(lines))
	for i := range lines {
		fields := strings.Fields(lines[i])
		if len(fields) != 2 {
			continue
		}
		refs = append(refs, fields[1])
	}
	return refs, nil
}

// DeleteBranchOptions contains optional arguments for deleting a branch.
// // Docs: https://git-scm.com/docs/git-branch
type DeleteBranchOptions struct {
	// Indicates whether to force delete the branch.
	Force bool
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// DeleteBranch deletes the branch from the repository.
func (r *Repository) DeleteBranch(name string, opts ...DeleteBranchOptions) error {
	var opt DeleteBranchOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("branch")
	if opt.Force {
		cmd.AddArgs("-D")
	} else {
		cmd.AddArgs("-d")
	}
	_, err := cmd.AddArgs(name).RunInDirWithTimeout(opt.Timeout, r.path)
	return err
}
