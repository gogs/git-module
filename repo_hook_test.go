// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepository_Hooks(t *testing.T) {
	t.Run("invalid hook", func(t *testing.T) {
		h, err := testrepo.Hook("", "bad_hook")
		assert.Equal(t, os.ErrNotExist, err)
		assert.Nil(t, h)
	})

	t.Run("no hooks directory", func(t *testing.T) {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		r, err := Open(wd)
		if err != nil {
			t.Fatal(err)
		}

		hooks, err := r.Hooks("")
		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, hooks)
	})

	t.Run("no hooks in the directory", func(t *testing.T) {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		r, err := Open(wd)
		if err != nil {
			t.Fatal(err)
		}

		dir := filepath.Join(r.Path(), DefaultHooksDir)
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			_ = os.RemoveAll(dir)
		}()

		hooks, err := r.Hooks("")
		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, hooks)
	})

	// Save "post-receive" hook with some content
	postReceiveHook := testrepo.NewHook(HookPostReceive)
	defer func() {
		_ = os.Remove(postReceiveHook.Path())
	}()

	err := postReceiveHook.Update("echo $1 $2 $3")
	if err != nil {
		t.Fatal(err)
	}

	hooks, err := testrepo.Hooks("")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(hooks))

	for i := range hooks {
		assert.NotEmpty(t, hooks[i].Content())
	}
}
