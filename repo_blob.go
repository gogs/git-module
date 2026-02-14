// Copyright 2022 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"context"
)

// CatFileBlobOptions contains optional arguments for verifying the objects.
//
// Docs: https://git-scm.com/docs/git-cat-file#Documentation/git-cat-file.txt
type CatFileBlobOptions struct {
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// CatFileBlob returns the blob corresponding to the given revision of the repository.
func (r *Repository) CatFileBlob(ctx context.Context, rev string, opts ...CatFileBlobOptions) (*Blob, error) {
	var opt CatFileBlobOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	rev, err := r.RevParse(ctx, rev, RevParseOptions{CommandOptions: opt.CommandOptions}) //nolint
	if err != nil {
		return nil, err
	}

	typ, err := r.CatFileType(ctx, rev, CatFileTypeOptions{CommandOptions: opt.CommandOptions})
	if err != nil {
		return nil, err
	}

	if typ != ObjectBlob {
		return nil, ErrNotBlob
	}

	return &Blob{
		TreeEntry: &TreeEntry{
			mode: EntryBlob,
			typ:  ObjectBlob,
			id:   MustIDFromString(rev),
			parent: &Tree{
				repo: r,
			},
		},
	}, nil
}
