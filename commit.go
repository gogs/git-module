// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

// Commit contains information of a Git commit.
type Commit struct {
	id        *SHA1
	parents   []*SHA1
	author    *Signature
	committer *Signature
	message   string

	Tree

	submodules     Submodules
	submodulesOnce sync.Once
	submodulesErr  error
}

// ID returns the SHA-1 hash of the commit.
func (c *Commit) ID() *SHA1 {
	return c.id
}

// Author returns the author of the commit.
func (c *Commit) Author() *Signature {
	return c.author
}

// Committer returns the committer of the commit.
func (c *Commit) Committer() *Signature {
	return c.committer
}

// Message returns the full commit message.
func (c *Commit) Message() string {
	return c.message
}

// Summary returns first line of commit message.
func (c *Commit) Summary() string {
	return strings.Split(c.message, "\n")[0]
}

// ParentsCount returns number of parents of the commit.
// It returns 0 if this is the root commit, otherwise returns 1, 2, etc.
func (c *Commit) ParentsCount() int {
	return len(c.parents)
}

// ParentID returns the SHA-1 hash of the n-th parent (0-based) of this commit.
// It returns an ErrParentNotExist if no such parent exists.
func (c *Commit) ParentID(n int) (*SHA1, error) {
	if n >= len(c.parents) {
		return nil, ErrParentNotExist
	}
	return c.parents[n], nil
}

// Parent returns the n-th parent commit (0-based) of this commit.
// It returns ErrRevisionNotExist if no such parent exists.
func (c *Commit) Parent(n int, opts ...CatFileCommitOptions) (*Commit, error) {
	id, err := c.ParentID(n)
	if err != nil {
		return nil, err
	}

	return c.repo.CatFileCommit(id.String(), opts...)
}

// CommitByPath returns the commit of the path in the state of this commit.
func (c *Commit) CommitByPath(opts ...CommitByRevisionOptions) (*Commit, error) {
	return c.repo.CommitByRevision(c.id.String(), opts...)
}

// CommitsByPage returns a paginated list of commits in the state of this commit.
// The returned list is in reverse chronological order.
func (c *Commit) CommitsByPage(page, size int, opts ...CommitsByPageOptions) ([]*Commit, error) {
	return c.repo.CommitsByPage(c.id.String(), page, size, opts...)
}

// SearchCommits searches commit message with given pattern. The returned list is in reverse
// chronological order.
func (c *Commit) SearchCommits(pattern string, opts ...SearchCommitsOptions) ([]*Commit, error) {
	return c.repo.SearchCommits(c.id.String(), pattern, opts...)
}

// ShowNameStatus returns name status of the commit.
func (c *Commit) ShowNameStatus(opts ...ShowNameStatusOptions) (*NameStatus, error) {
	return c.repo.ShowNameStatus(c.id.String(), opts...)
}

// CommitsCount returns number of total commits up to this commit.
func (c *Commit) CommitsCount(opts ...RevListCountOptions) (int64, error) {
	return c.repo.RevListCount([]string{c.id.String()}, opts...)
}

// FilesChangedSince returns a list of files changed after given commit ID.
func (c *Commit) FilesChangedAfter(after string, opts ...DiffNameOnlyOptions) ([]string, error) {
	return c.repo.DiffNameOnly(after, c.id.String(), opts...)
}

// CommitsAfter returns a list of commits after given commit ID up to this commit. The returned
// list is in reverse chronological order.
func (c *Commit) CommitsAfter(after string, opts ...RevListOptions) ([]*Commit, error) {
	return c.repo.RevList([]string{after + "..." + c.id.String()}, opts...)
}

// Ancestors returns a list of ancestors of this commit in reverse chronological order.
func (c *Commit) Ancestors(opts ...LogOptions) ([]*Commit, error) {
	if c.ParentsCount() == 0 {
		return []*Commit{}, nil
	}

	var opt LogOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	opt.Skip++

	return c.repo.Log(c.id.String(), opt)
}

type limitWriter struct {
	W io.Writer
	N int64
}

func (w *limitWriter) Write(p []byte) (int, error) {
	if w.N <= 0 {
		return len(p), nil
	}

	limit := int64(len(p))
	if limit > w.N {
		limit = w.N
	}
	n, err := w.W.Write(p[:limit])
	w.N -= int64(n)

	// Prevent "short write" error
	return len(p), err
}

// IsImageFile returns true if the commit is an image blob.
func (c *Commit) IsImageFile(name string) (bool, error) {
	blob, err := c.Blob(name)
	if err != nil {
		return false, err
	}

	buf := new(bytes.Buffer)
	buf.Grow(512)
	stdout := &limitWriter{
		W: buf,
		N: int64(buf.Cap()),
	}

	err = blob.Pipeline(stdout, ioutil.Discard)
	if err != nil {
		return false, err
	}

	return strings.Contains(http.DetectContentType(buf.Bytes()), "image/"), nil
}
