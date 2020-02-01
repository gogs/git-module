// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bufio"
	"container/list"
	"io"
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

// ParentCount returns number of parents of the commit.
// It returns 0 if this is the root commit, otherwise returns 1, 2, etc.
func (c *Commit) ParentCount() int {
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

	dataRc, err := blob.Data()
	if err != nil {
		return false
	}
	buf := make([]byte, 1024)
	n, _ := dataRc.Read(buf)
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

// Submodules contains information of submodules.
type Submodules = *objectCache

// Submodules returns submodules found in this commit.
func (c *Commit) Submodules() (Submodules, error) {
	c.submodulesOnce.Do(func() {
		var e *TreeEntry
		e, c.submodulesErr = c.GetTreeEntryByPath(".gitmodules")
		if c.submodulesErr != nil {
			return
		}

		var r io.Reader
		r, c.submodulesErr = e.Blob().Data()
		if c.submodulesErr != nil {
			return
		}

		scanner := bufio.NewScanner(r)
		c.submodules = newObjectCache()
		var inSection bool
		var path string
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "[submodule") {
				inSection = true
				continue
			}
			if inSection {
				fields := strings.Split(scanner.Text(), "=")
				k := strings.TrimSpace(fields[0])
				if k == "path" {
					path = strings.TrimSpace(fields[1])
				} else if k == "url" {
					c.submodules.Set(path, &Submodule{
						Name: path,
						URL:  strings.TrimSpace(fields[1])},
					)
					inSection = false
				}
			}
		}
	})

	return c.submodules, c.submodulesErr
}

// Submodule returns submodule by given name.
func (c *Commit) Submodule(name string) (*Submodule, error) {
	mods, err := c.Submodules()
	if err != nil {
		return nil, err
	}

	m, has := mods.Get(name)
	if has {
		return m.(*Submodule), nil
	}
	return nil, nil
}
