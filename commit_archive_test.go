package git

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func tempPath() string {
	return filepath.Join(os.TempDir(), strconv.Itoa(int(time.Now().UnixNano())))
}

func TestCommit_Archive(t *testing.T) {
	ctx := context.Background()
	for _, format := range []ArchiveFormat{
		ArchiveZip,
		ArchiveTarGz,
	} {
		t.Run(string(format), func(t *testing.T) {
			c, err := testrepo.CatFileCommit(ctx, testrepoMarks[1].String())
			if err != nil {
				t.Fatal(err)
			}

			dst := tempPath()
			defer func() {
				_ = os.Remove(dst)
			}()

			assert.Nil(t, c.Archive(ctx, format, dst))
		})
	}
}
