// Copyright 2022 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

// CatFileBlob returns the blob corresponding to the given revision of the repository.
func (repo *Repository) CatFileBlob(rev string) (*Blob, error) {
	typ, err := repo.CatFileType(rev)
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
				repo: repo,
			},
		},
	}, nil
}
