// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTree_TreeEntry(t *testing.T) {
	tree, err := testrepo.LsTree("master")
	if err != nil {
		t.Fatal(err)
	}

	e, err := tree.TreeEntry("")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, tree.id, e.ID())
	assert.Equal(t, ObjectTree, e.Type())
	assert.True(t, e.IsTree())
}
