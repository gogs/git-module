// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlob(t *testing.T) {
	expOutput := `Copyright (c) 2015 All Gogs Contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.`

	blob := &Blob{
		repo: &Repository{},
		TreeEntry: &TreeEntry{
			ID: MustIDFromString("176d8dfe018c850d01851b05fb8a430096247353"), // Blob ID of "LICENSE" file
			ptree: &Tree{
				repo: &Repository{},
			},
		},
	}

	t.Run("get data all at once", func(t *testing.T) {
		p, err := blob.Bytes()
		assert.Nil(t, err)
		assert.Equal(t, expOutput, string(p))
	})

	t.Run("get data with pipeline", func(t *testing.T) {
		stdout := new(bytes.Buffer)
		err := blob.Pipeline(stdout, nil)
		assert.Nil(t, err)
		assert.Equal(t, expOutput, stdout.String())
	})
}
