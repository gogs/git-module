// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

func (r *Repository) getTree(id SHA1) (*Tree, error) {
	treePath := filepathFromSHA1(r.path, id.String())
	if isFile(treePath) {
		_, err := NewCommand("ls-tree", id.String()).RunInDir(r.path)
		if err != nil {
			return nil, ErrRevisionNotExist{id.String(), ""}
		}
	}

	return NewTree(r, id), nil
}

// Find the tree object in the repository.
func (r *Repository) GetTree(idStr string) (*Tree, error) {
	id, err := NewIDFromString(idStr)
	if err != nil {
		return nil, err
	}
	return r.getTree(id)
}
