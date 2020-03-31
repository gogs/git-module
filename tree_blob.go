// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"path"
	"strings"
)

// TreeEntry returns the TreeEntry by given subpath of the tree.
func (t *Tree) TreeEntry(subpath string, opts ...LsTreeOptions) (*TreeEntry, error) {
	if len(subpath) == 0 {
		return &TreeEntry{
			id:   t.id,
			typ:  ObjectTree,
			mode: EntryTree,
		}, nil
	}

	subpath = path.Clean(subpath)
	paths := strings.Split(subpath, "/")
	var err error
	tree := t
	for i, name := range paths {
		// Reached end of the loop
		if i == len(paths)-1 {
			entries, err := tree.Entries(opts...)
			if err != nil {
				return nil, err
			}

			for _, v := range entries {
				if v.name == name {
					return v, nil
				}
			}
		} else {
			tree, err = tree.Subtree(name, opts...)
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, ErrRevisionNotExist
}

// Blob returns the blob object by given subpath of the tree.
func (t *Tree) Blob(subpath string, opts ...LsTreeOptions) (*Blob, error) {
	e, err := t.TreeEntry(subpath, opts...)
	if err != nil {
		return nil, err
	}

	if e.IsBlob() || e.IsExec() {
		return e.Blob(), nil
	}

	return nil, ErrNotBlob
}

// BlobByIndex returns blob object by given index.
func (t *Tree) BlobByIndex(index string) (*Blob, error) {
	typ, err := t.repo.CatFileType(index)
	if err != nil {
		return nil, err
	}

	if typ != ObjectBlob {
		return nil, ErrNotBlob
	}

	id, err := t.repo.RevParse(index)
	if err != nil {
		return nil, err
	}

	return &Blob{
		TreeEntry: &TreeEntry{
			mode:   EntryBlob,
			typ:    ObjectBlob,
			id:     MustIDFromString(id),
			parent: t,
		},
	}, nil
}
