package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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
			_, err := UpdateServerInfo(test.path, UpdateServerInfoOptions{Force: true})
			assert.NoError(t, err)
			_, err = os.Stat(filepath.Join(test.path, "info", "refs"))
			assert.NoError(t, err)
		})
	}
}
