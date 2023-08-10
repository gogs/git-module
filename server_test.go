// Copyright 2023 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateServerInfo(t *testing.T) {
	tests := []struct {
		path string
	}{
		{
			path: repoPath,
		},
	}
	for _, test := range tests {
		t.Run("update-server-info", func(t *testing.T) {
			err := os.RemoveAll(filepath.Join(test.path, "info"))
			require.NoError(t, err)
			err = UpdateServerInfo(test.path, UpdateServerInfoOptions{Force: true})
			require.NoError(t, err)
			assert.True(t, isFile(filepath.Join(test.path, "info", "refs")))
		})
	}
}
