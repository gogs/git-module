// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

// Blame contains information of a Git file blame.
type Blame struct {
	lines []*Commit
}

// Line returns the commit by given line number (1-based). It returns nil when
// no such line.
func (b *Blame) Line(i int) *Commit {
	if i <= 0 || len(b.lines) < i {
		return nil
	}
	return b.lines[i-1]
}
