// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"context"
	"path/filepath"
	"strings"
)

// ArchiveFormat is the format of an archive.
type ArchiveFormat string

// A list of formats can be created by Git for an archive.
const (
	ArchiveZip   ArchiveFormat = "zip"
	ArchiveTarGz ArchiveFormat = "tar.gz"
)

// Archive creates given format of archive to the destination.
func (c *Commit) Archive(ctx context.Context, format ArchiveFormat, dst string) error {
	prefix := filepath.Base(strings.TrimSuffix(c.repo.path, ".git")) + "/"
	_, err := exec(ctx, c.repo.path, []string{
		"archive",
		"--prefix=" + prefix,
		"--format=" + string(format),
		"-o", dst,
		"--end-of-options",
		c.ID.String(),
	}, nil)
	return err
}
