// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
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

func TestCommit_CreateArchive(t *testing.T) {
	for _, format := range []ArchiveFormat{
		ArchiveZip,
		ArchiveTarGz,
	} {
		t.Run(string(format), func(t *testing.T) {
			c, err := testrepo.CatFileCommit("755fd577edcfd9209d0ac072eed3b022cbe4d39b")
			if err != nil {
				t.Fatal(err)
			}

			dst := tempPath()
			defer func() {
				_ = os.Remove(dst)
			}()

			assert.Nil(t, c.CreateArchive(format, dst))
		})
	}
}
