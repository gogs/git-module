// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"io"
)

// Blob is a blob object.
type Blob struct {
	*TreeEntry
}

// Bytes reads and returns the content of the blob all at once in bytes. This
// can be very slow and memory consuming for huge content.
func (b *Blob) Bytes() ([]byte, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	// Preallocate memory to save ~50% memory usage on big files.
	stdout.Grow(int(b.Size()))

	if err := b.Pipeline(stdout, stderr); err != nil {
		return nil, concatenateError(err, stderr.String())
	}
	return stdout.Bytes(), nil
}

// Pipeline reads the content of the blob and pipes stdout and stderr to
// supplied io.Writer.
func (b *Blob) Pipeline(stdout, stderr io.Writer) error {
	return NewCommand("show", b.id.String()).RunInDirPipeline(stdout, stderr, b.parent.repo.path)
}
