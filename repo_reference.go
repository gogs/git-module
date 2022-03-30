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

// RefShortName returns short name of heads or tags. Other references will
// return original string.
func RefShortName(ref string) string {
	if strings.HasPrefix(ref, RefsHeads) {
		return ref[len(RefsHeads):]
	} else if strings.HasPrefix(ref, RefsTags) {
		return ref[len(RefsTags):]
	}

	return ref
}

// Reference contains information of a Git reference.
type Reference struct {
	ID      string
	Refspec string
}

// ShowRefVerifyOptions contains optional arguments for verifying a reference.
//
// Docs: https://git-scm.com/docs/git-show-ref#Documentation/git-show-ref.txt---verify
type ShowRefVerifyOptions struct {
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	Timeout time.Duration
}

var ErrReferenceNotExist = errors.New("reference does not exist")

// ShowRefVerify returns the commit ID of given reference if it exists in the
// repository in given path.
func ShowRefVerify(repoPath, ref string, opts ...ShowRefVerifyOptions) (string, error) {
	var opt ShowRefVerifyOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	stdout, err := NewCommand("show-ref", "--verify", ref).RunInDirWithTimeout(opt.Timeout, repoPath)
	if err != nil {
		if strings.Contains(err.Error(), "not a valid ref") {
			return "", ErrReferenceNotExist
		}
		return "", err
	}
	return strings.Split(string(stdout), " ")[0], nil
}

// Deprecated: Use ShowRefVerify instead.
func RepoShowRefVerify(repoPath, ref string, opts ...ShowRefVerifyOptions) (string, error) {
	return ShowRefVerify(repoPath, ref, opts...)
}

// ShowRefVerify returns the commit ID of given reference (e.g.
// "refs/heads/master") if it exists in the repository.
func (r *Repository) ShowRefVerify(ref string, opts ...ShowRefVerifyOptions) (string, error) {
	return ShowRefVerify(r.path, ref, opts...)
}

// BranchCommitID returns the commit ID of given branch if it exists in the
// repository. The branch must be given in short name e.g. "master".
func (r *Repository) BranchCommitID(branch string, opts ...ShowRefVerifyOptions) (string, error) {
	return r.ShowRefVerify(RefsHeads+branch, opts...)
}

// TagCommitID returns the commit ID of given tag if it exists in the
// repository. The tag must be given in short name e.g. "v1.0.0".
func (r *Repository) TagCommitID(tag string, opts ...ShowRefVerifyOptions) (string, error) {
	return r.ShowRefVerify(RefsTags+tag, opts...)
}

// RepoHasReference returns true if given reference exists in the repository in
// given path. The reference must be given in full refspec, e.g.
// "refs/heads/master".
func RepoHasReference(repoPath, ref string, opts ...ShowRefVerifyOptions) bool {
	_, err := ShowRefVerify(repoPath, ref, opts...)
	return err == nil
}

// RepoHasBranch returns true if given branch exists in the repository in given
// path. The branch must be given in short name e.g. "master".
func RepoHasBranch(repoPath, branch string, opts ...ShowRefVerifyOptions) bool {
	return RepoHasReference(repoPath, RefsHeads+branch, opts...)
}

// HasTag returns true if given tag exists in the repository in given path. The
// tag must be given in short name e.g. "v1.0.0".
func HasTag(repoPath, tag string, opts ...ShowRefVerifyOptions) bool {
	return RepoHasReference(repoPath, RefsTags+tag, opts...)
}

// Deprecated: Use HasTag instead.
func RepoHasTag(repoPath, tag string, opts ...ShowRefVerifyOptions) bool {
	return HasTag(repoPath, tag, opts...)
}

// HasReference returns true if given reference exists in the repository. The
// reference must be given in full refspec, e.g. "refs/heads/master".
func (r *Repository) HasReference(ref string, opts ...ShowRefVerifyOptions) bool {
	return RepoHasReference(r.path, ref, opts...)
}

// HasBranch returns true if given branch exists in the repository. The branch
// must be given in short name e.g. "master".
func (r *Repository) HasBranch(branch string, opts ...ShowRefVerifyOptions) bool {
	return RepoHasBranch(r.path, branch, opts...)
}

// HasTag returns true if given tag exists in the repository. The tag must be
// given in short name e.g. "v1.0.0".
func (r *Repository) HasTag(tag string, opts ...ShowRefVerifyOptions) bool {
	return HasTag(r.path, tag, opts...)
}

// SymbolicRefOptions contains optional arguments for get and set symbolic ref.
type SymbolicRefOptions struct {
	// The name of the symbolic ref. When not set, default ref "HEAD" is used.
	Name string
	// The name of the reference, e.g. "refs/heads/master". When set, it will be
	// used to update the symbolic ref.
	Ref string
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// SymbolicRef returns the reference name (e.g. "refs/heads/master") pointed by
// the symbolic ref. It returns an empty string and nil error when doing set
// operation.
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
//
// Docs: https://git-scm.com/docs/git-show-ref
type ShowRefOptions struct {
	// Indicates whether to include heads.
	Heads bool
	// Indicates whether to include tags.
	Tags bool
	// The list of patterns to filter results.
	Patterns []string
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// ShowRef returns a list of references in the repository.
func (r *Repository) ShowRef(opts ...ShowRefOptions) ([]*Reference, error) {
	var opt ShowRefOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("show-ref").AddOptions(opt.CommandOptions)
	if opt.Heads {
		cmd.AddArgs("--heads")
	}
	if opt.Tags {
		cmd.AddArgs("--tags")
	}
	cmd.AddArgs("--")
	if len(opt.Patterns) > 0 {
		cmd.AddArgs(opt.Patterns...)
	}

	stdout, err := cmd.RunInDirWithTimeout(opt.Timeout, r.path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(stdout), "\n")
	refs := make([]*Reference, 0, len(lines))
	for i := range lines {
		fields := strings.Fields(lines[i])
		if len(fields) != 2 {
			continue
		}
		refs = append(refs, &Reference{
			ID:      fields[0],
			Refspec: fields[1],
		})
	}
	return refs, nil
}

// Branches returns a list of all branches in the repository.
func (r *Repository) Branches() ([]string, error) {
	heads, err := r.ShowRef(ShowRefOptions{Heads: true})
	if err != nil {
		return nil, err
	}

	branches := make([]string, len(heads))
	for i := range heads {
		branches[i] = strings.TrimPrefix(heads[i].Refspec, RefsHeads)
	}
	return branches, nil
}

// DeleteBranchOptions contains optional arguments for deleting a branch.
//
// // Docs: https://git-scm.com/docs/git-branch
type DeleteBranchOptions struct {
	// Indicates whether to force delete the branch.
	Force bool
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// DeleteBranch deletes the branch from the repository in given path.
func DeleteBranch(repoPath, name string, opts ...DeleteBranchOptions) error {
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
	_, err := cmd.AddArgs(name).RunInDirWithTimeout(opt.Timeout, repoPath)
	return err
}

// Deprecated: Use DeleteBranch instead.
func RepoDeleteBranch(repoPath, name string, opts ...DeleteBranchOptions) error {
	return DeleteBranch(repoPath, name, opts...)
}

// DeleteBranch deletes the branch from the repository.
func (r *Repository) DeleteBranch(name string, opts ...DeleteBranchOptions) error {
	return DeleteBranch(r.path, name, opts...)
}
