// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"context"
	"errors"
	"strings"
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
	// The additional options to be passed to the underlying git.
	CommandOptions
}

var ErrReferenceNotExist = errors.New("reference does not exist")

// ShowRefVerify returns the commit ID of given reference (e.g.
// "refs/heads/master") if it exists in the repository.
func (r *Repository) ShowRefVerify(ctx context.Context, ref string, opts ...ShowRefVerifyOptions) (string, error) {
	var opt ShowRefVerifyOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	args := []string{"show-ref", "--verify"}
	args = append(args, opt.Args...)
	args = append(args, "--end-of-options", ref)

	stdout, err := gitRun(ctx, r.path, args, opt.Envs)
	if err != nil {
		if strings.Contains(err.Error(), "not a valid ref") {
			return "", ErrReferenceNotExist
		}
		return "", err
	}
	return strings.Split(string(stdout), " ")[0], nil
}

// BranchCommitID returns the commit ID of given branch if it exists in the
// repository. The branch must be given in short name e.g. "master".
func (r *Repository) BranchCommitID(ctx context.Context, branch string, opts ...ShowRefVerifyOptions) (string, error) {
	return r.ShowRefVerify(ctx, RefsHeads+branch, opts...)
}

// TagCommitID returns the commit ID of given tag if it exists in the
// repository. The tag must be given in short name e.g. "v1.0.0".
func (r *Repository) TagCommitID(ctx context.Context, tag string, opts ...ShowRefVerifyOptions) (string, error) {
	return r.ShowRefVerify(ctx, RefsTags+tag, opts...)
}

// HasReference returns true if given reference exists in the repository. The
// reference must be given in full refspec, e.g. "refs/heads/master".
func (r *Repository) HasReference(ctx context.Context, ref string, opts ...ShowRefVerifyOptions) bool {
	_, err := r.ShowRefVerify(ctx, ref, opts...)
	return err == nil
}

// HasBranch returns true if given branch exists in the repository. The branch
// must be given in short name e.g. "master".
func (r *Repository) HasBranch(ctx context.Context, branch string, opts ...ShowRefVerifyOptions) bool {
	return r.HasReference(ctx, RefsHeads+branch, opts...)
}

// HasTag returns true if given tag exists in the repository. The tag must be
// given in short name e.g. "v1.0.0".
func (r *Repository) HasTag(ctx context.Context, tag string, opts ...ShowRefVerifyOptions) bool {
	return r.HasReference(ctx, RefsTags+tag, opts...)
}

// SymbolicRefOptions contains optional arguments for get and set symbolic ref.
type SymbolicRefOptions struct {
	// The name of the symbolic ref. When not set, default ref "HEAD" is used.
	Name string
	// The name of the reference, e.g. "refs/heads/master". When set, it will be
	// used to update the symbolic ref.
	Ref string
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// SymbolicRef returns the reference name (e.g. "refs/heads/master") pointed by
// the symbolic ref. It returns an empty string and nil error when doing set
// operation.
func (r *Repository) SymbolicRef(ctx context.Context, opts ...SymbolicRefOptions) (string, error) {
	var opt SymbolicRefOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	args := []string{"symbolic-ref"}
	args = append(args, opt.Args...)
	if opt.Name == "" {
		opt.Name = "HEAD"
	}
	args = append(args, "--end-of-options", opt.Name)
	if opt.Ref != "" {
		args = append(args, opt.Ref)
	}

	stdout, err := gitRun(ctx, r.path, args, opt.Envs)
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
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// ShowRef returns a list of references in the repository.
func (r *Repository) ShowRef(ctx context.Context, opts ...ShowRefOptions) ([]*Reference, error) {
	var opt ShowRefOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	args := []string{"show-ref"}
	args = append(args, opt.Args...)
	if opt.Heads {
		args = append(args, "--heads")
	}
	if opt.Tags {
		args = append(args, "--tags")
	}
	args = append(args, "--")
	if len(opt.Patterns) > 0 {
		args = append(args, opt.Patterns...)
	}

	stdout, err := gitRun(ctx, r.path, args, opt.Envs)
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
func (r *Repository) Branches(ctx context.Context) ([]string, error) {
	heads, err := r.ShowRef(ctx, ShowRefOptions{Heads: true})
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
// Docs: https://git-scm.com/docs/git-branch
type DeleteBranchOptions struct {
	// Indicates whether to force delete the branch.
	Force bool
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// DeleteBranch deletes the branch from the repository.
func (r *Repository) DeleteBranch(ctx context.Context, name string, opts ...DeleteBranchOptions) error {
	var opt DeleteBranchOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	args := []string{"branch"}
	args = append(args, opt.Args...)
	if opt.Force {
		args = append(args, "-D")
	} else {
		args = append(args, "-d")
	}
	args = append(args, "--end-of-options", name)

	_, err := gitRun(ctx, r.path, args, opt.Envs)
	return err
}
