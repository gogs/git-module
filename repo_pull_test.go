// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepository_MergeBase(t *testing.T) {
	ctx := context.Background()

	t.Run("no merge base", func(t *testing.T) {
		mb, err := testrepo.MergeBase(ctx, "0eedd79eba4394bbef888c804e899731644367fe", "bad_revision")
		assert.Equal(t, ErrNoMergeBase, err)
		assert.Empty(t, mb)
	})

	tests := []struct {
		base         string
		head         string
		opt          MergeBaseOptions
		expMergeBase string
	}{
		{
			base:         "4e59b72440188e7c2578299fc28ea425fbe9aece",
			head:         "0eedd79eba4394bbef888c804e899731644367fe",
			expMergeBase: "4e59b72440188e7c2578299fc28ea425fbe9aece",
		},
		{
			base:         "master",
			head:         "release-1.0",
			expMergeBase: "0eedd79eba4394bbef888c804e899731644367fe",
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			mb, err := testrepo.MergeBase(ctx, test.base, test.head, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expMergeBase, mb)
		})
	}
}
