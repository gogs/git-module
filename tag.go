// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

// Tag contains information of a Git tag.
type Tag struct {
	typ      ObjectType
	id       *SHA1
	commitID *SHA1 // The ID of the underlying commit
	refspec  string
	tagger   *Signature
	message  string

	repo *Repository
}

// Commit returns the underlying commit of the tag.
func (tag *Tag) Commit(opts ...CatFileCommitOptions) (*Commit, error) {
	return tag.repo.CatFileCommit(tag.commitID.String(), opts...)
}
