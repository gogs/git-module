// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubmoduleEntry_ID(t *testing.T) {
	e := SubmoduleEntry{
		id: MustIDFromString(EmptyID),
	}
	assert.Equal(t, EmptyID, e.ID().String())
}
