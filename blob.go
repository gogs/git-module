// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"context"
	"io"
)

// Blob is a blob object.
type Blob struct {
	*TreeEntry
}

// Bytes reads and returns the content of the blob all at once in bytes. This
// can be very slow and memory consuming for huge content.
func (b *Blob) Bytes(ctx context.Context) ([]byte, error) {
	stdout := new(bytes.Buffer)

	// Preallocate memory to save ~50% memory usage on big files.
	if size := b.Size(ctx); size > 0 && size < int64(^uint(0)>>1) {
		stdout.Grow(int(size))
	}

	if err := b.Pipeline(ctx, stdout, nil); err != nil {
		return nil, err
	}
	return stdout.Bytes(), nil
}

// Pipeline reads the content of the blob and pipes stdout and stderr to
// supplied io.Writer.
func (b *Blob) Pipeline(ctx context.Context, stdout, stderr io.Writer) error {
	return gitPipeline(ctx, b.parent.repo.path, []string{"show", b.id.String()}, nil, stdout, stderr, nil)
}
