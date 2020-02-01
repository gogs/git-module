// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"container/list"
	"net/http"
	"strings"
	"sync"
)

// Commit contains information of a Git commit.
type Commit struct {
	id        SHA1
	parents   []SHA1
	author    *Signature
	committer *Signature
	message   string

	Tree

	submodules     Submodules
	submodulesOnce sync.Once
	submodulesErr  error
}

// ID returns the SHA-1 hash of the commit.
func (c *Commit) ID() SHA1 {
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

// ParentID returns the SHA-1 hash of the n-th parent (0-based) of this commit.
// It returns ErrNotExist if no such parent exists.
func (c *Commit) ParentID(n int) (SHA1, error) {
	if n >= len(c.parents) {
		return SHA1{}, ErrNotExist{"", ""}
	}
	return c.parents[n], nil
}

// Parent returns the n-th parent commit (0-based) of this commit.
// It returns ErrNotExist if no such parent exists.
func (c *Commit) Parent(n int) (*Commit, error) {
	id, err := c.ParentID(n)
	if err != nil {
		return nil, err
	}

	return c.repo.getCommit(id)
}

// ParentsCount returns number of parents of the commit.
// It returns 0 if this is the root commit, otherwise returns 1, 2, etc.
func (c *Commit) ParentsCount() int {
	return len(c.parents)
}

func isImageFile(data []byte) (string, bool) {
	contentType := http.DetectContentType(data)
	if strings.Contains(contentType, "image/") {
		return contentType, true
	}
	return contentType, false
}

// IsImageFile returns true if the commit is a image blob.
func (c *Commit) IsImageFile(name string) bool {
	blob, err := c.GetBlobByPath(name)
	if err != nil {
		return false
	}

	r, err := blob.Reader()
	if err != nil {
		return false
	}
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	buf = buf[:n]
	_, isImage := isImageFile(buf)
	return isImage
}

// CommitByPath returns the commit of relative path.
func (c *Commit) CommitByPath(relpath string) (*Commit, error) {
	return c.repo.getCommitByPathWithID(c.id, relpath)
}

// CommitsCount returns number of total commits up to this commit.
func (c *Commit) CommitsCount() (int64, error) {
	return c.repo.CommitsCount(c.id.String())
}

// CommitsByPage returns a paginated list of commits with given page and size.
// The pagination starts from the newest to the oldest commit.
func (c *Commit) CommitsByPage(page, size int) (*list.List, error) {
	return c.repo.CommitsByRangeSize(c.id.String(), page, size)
}

// Ancestors returns a list of ancestors of this commit from the newest to the oldest.
func (c *Commit) Ancestors() (*list.List, error) {
	return c.repo.getCommitsBefore(c.id)
}

// AncestorsWithLimit returns a list of ancestors of this commit from the newest to the oldest
// until reached limited size of the list.
func (c *Commit) AncestorsWithLimit(limit int) (*list.List, error) {
	return c.repo.getCommitsBeforeLimit(c.id, limit)
}

// CommitsAfter returns a list of commits after given commit ID up to this commit.
// The returned list sorted from the newest to the oldest.
func (c *Commit) CommitsAfter(commitID string) (*list.List, error) {
	endCommit, err := c.repo.GetCommit(commitID)
	if err != nil {
		return nil, err
	}
	return c.repo.CommitsBetween(c, endCommit)
}

// SearchCommits searches commit message with given keyword. It returns a list of matched commits
// from the newest to the oldest
func (c *Commit) SearchCommits(keyword string) (*list.List, error) {
	return c.repo.searchCommits(c.id, keyword)
}

// FilesChangedSince returns a list of files changed since given commit ID.
func (c *Commit) FilesChangedSince(commitID string) ([]string, error) {
	return c.repo.getFilesChanged(commitID, c.id.String())
}
