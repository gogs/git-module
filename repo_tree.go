// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"time"
)

// LsTreeOptions contains optional arguments for listing trees.
// Docs: https://git-scm.com/docs/git-ls-tree
type LsTreeOptions struct {
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// LsTree returns the tree object in the repository by given revision.
func (r *Repository) LsTree(rev string, opts ...LsTreeOptions) (*Tree, error) {
	var opt LsTreeOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	_, err := NewCommand("ls-tree", rev).RunInDirWithTimeout(opt.Timeout, r.path)
	if err != nil {
		return nil, ErrRevisionNotExist{rev, ""}
	}

	rev, err = r.RevParse(rev, RevParseOptions{Timeout: opt.Timeout})
	if err != nil {
		return nil, err
	}

	return &Tree{
		id:   MustIDFromString(rev),
		repo: r,
	}, nil
}
