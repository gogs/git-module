// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
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

// CreateArchive creates given format of archive to the destination.
func (c *Commit) CreateArchive(format ArchiveFormat, dst string) error {
	prefix := filepath.Base(strings.TrimSuffix(c.repo.path, ".git")) + "/"
	_, err := NewCommand("archive",
		"--prefix="+prefix,
		"--format="+string(format),
		"-o", dst,
		c.ID.String(),
	).RunInDir(c.repo.path)
	return err
}
