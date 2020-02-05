// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

// ObjectType is the type of a Git objet.
type ObjectType string

// A list of object types.
const (
	ObjectCommit ObjectType = "commit"
	ObjectTree   ObjectType = "tree"
	ObjectBlob   ObjectType = "blob"
	ObjectTag    ObjectType = "tag"
)
