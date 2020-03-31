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

func TestTree_Blob(t *testing.T) {
	tree, err := testrepo.LsTree("d58e3ef9f123eea6857161c79275ee22b228f659")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("not a blob", func(t *testing.T) {
		_, err := tree.Blob("src")
		assert.Equal(t, ErrNotBlob, err)
	})

	t.Run("get a blob", func(t *testing.T) {
		b, err := tree.Blob("README.txt")
		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, b.IsBlob())
	})

	t.Run("get an executable as blob", func(t *testing.T) {
		b, err := tree.Blob("run.sh")
		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, b.IsExec())
	})
}
