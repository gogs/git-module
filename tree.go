// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"strings"
	"sync"
)

// Tree represents a flat directory listing in Git.
type Tree struct {
	id     *SHA1
	parent *Tree

	repo *Repository

	entries     Entries
	entriesOnce sync.Once
	entriesErr  error
}

// Subtree returns a subtree by given subpath of the tree.
func (t *Tree) Subtree(subpath string, opts ...LsTreeOptions) (*Tree, error) {
	if len(subpath) == 0 {
		return t, nil
	}

	paths := strings.Split(subpath, "/")
	var (
		err error
		g   = t
		p   = t
		e   *TreeEntry
	)
	for _, name := range paths {
		e, err = p.TreeEntry(name, opts...)
		if err != nil {
			return nil, err
		}

		g = &Tree{
			id:     e.id,
			parent: p,
			repo:   t.repo,
		}
		p = g
	}
	return g, nil
}

// Entries returns all entries of the tree.
func (t *Tree) Entries(opts ...LsTreeOptions) (Entries, error) {
	t.entriesOnce.Do(func() {
		if t.entries != nil {
			return
		}

		var tt *Tree
		tt, t.entriesErr = t.repo.LsTree(t.id.String(), opts...)
		if t.entriesErr != nil {
			return
		}
		t.entries = tt.entries
	})

	return t.entries, t.entriesErr
}
