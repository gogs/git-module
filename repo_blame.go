// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"time"
)

// BlameOptions contains optional arguments for blaming a file.
// Docs: https://git-scm.com/docs/git-blame
type BlameOptions struct {
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// BlameFile returns blame results of the file with the given revision of the
// repository.
func (r *Repository) BlameFile(rev, file string, opts ...BlameOptions) (*Blame, error) {
	var opt BlameOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	stdout, err := NewCommand("blame", "-l", "-s", rev, "--", file).RunInDirWithTimeout(opt.Timeout, r.path)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(stdout, []byte{'\n'})
	blame := &Blame{
		lines: make([]*Commit, 0, len(lines)),
	}
	for _, line := range lines {
		if len(line) < 40 {
			break
		}
		id := line[:40]

		// Earliest commit is indicated by a leading "^"
		if id[0] == '^' {
			id = id[1:]
		}
		commit, err := r.CatFileCommit(string(id), CatFileCommitOptions{Timeout: opt.Timeout}) //nolint
		if err != nil {
			return nil, err
		}
		blame.lines = append(blame.lines, commit)
	}
	return blame, nil
}
