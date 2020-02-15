// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTreeEntry(t *testing.T) {
	id := MustIDFromString("0eedd79eba4394bbef888c804e899731644367fe")
	e := &TreeEntry{
		mode: EntrySymlink,
		typ:  ObjectTree,
		id:   id,
		name: "go.mod",
	}

	assert.False(t, e.IsTree())
	assert.False(t, e.IsBlob())
	assert.False(t, e.IsExec())
	assert.True(t, e.IsSymlink())
	assert.False(t, e.IsCommit())

	assert.Equal(t, ObjectTree, e.Type())
	assert.Equal(t, e.id, e.ID())
	assert.Equal(t, "go.mod", e.Name())
}
