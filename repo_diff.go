// Copyright 2017 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"fmt"
	"io"
	"time"
)

// DiffRangeOptions contains optional arguments for parsing diff.
// Docs: https://git-scm.com/docs/git-diff#Documentation/git-diff.txt---full-index
type DiffRangeOptions struct {
	// The commit ID to used for computing diff between a range of commits (base, revision]. When not set,
	// only computes diff for a single commit at revision.
	Base string
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// Diff returns a parsed diff object between given commits.
func (r *Repository) Diff(rev string, maxLines, maxLineChars, maxFiles int, opts ...DiffRangeOptions) (*Diff, error) {
	var opt DiffRangeOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	commit, err := r.CatFileCommit(rev, CatFileCommitOptions{Timeout: opt.Timeout})
	if err != nil {
		return nil, err
	}

	cmd := NewCommand()
	if opt.Base == "" {
		// First commit of repository
		if commit.ParentsCount() == 0 {
			cmd.AddArgs("show", "--full-index", rev)
		} else {
			c, _ := commit.Parent(0)
			cmd.AddArgs("diff", "--full-index", "-M", c.id.String(), rev)
		}
	} else {
		cmd.AddArgs("diff", "--full-index", "-M", opt.Base, rev)
	}

	stdout, w := io.Pipe()
	done := make(chan error)
	var diff *Diff
	go func() {
		diff = SteamParsePatch(stdout, done, maxLines, maxLineChars, maxFiles)
	}()

	stderr := new(bytes.Buffer)
	err = cmd.RunInDirPipelineWithTimeout(2*time.Minute, w, stderr, r.path)
	_ = w.Close() // Close writer to exit parsing goroutine
	if err != nil {
		return nil, concatenateError(err, stderr.String())
	}

	return diff, <-done
}

// RawDiffFormat is the format of a raw diff.
type RawDiffFormat string

const (
	RawDiffNormal RawDiffFormat = "diff"
	RawDiffPatch  RawDiffFormat = "patch"
)

// RawDiffOptions contains optional arguments for dumpping a raw diff.
// Docs: https://git-scm.com/docs/git-format-patch
type RawDiffOptions struct {
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// GetRawDiff dumps diff of repository in given revision directly to given io.Writer.
func (r *Repository) GetRawDiff(rev string, diffType RawDiffFormat, w io.Writer, opts ...RawDiffOptions) error {
	var opt RawDiffOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	commit, err := r.CatFileCommit(rev, CatFileCommitOptions{Timeout: opt.Timeout})
	if err != nil {
		return err
	}

	cmd := NewCommand()
	switch diffType {
	case RawDiffNormal:
		if commit.ParentsCount() == 0 {
			cmd.AddArgs("show", rev)
		} else {
			c, _ := commit.Parent(0)
			cmd.AddArgs("diff", "-M", c.id.String(), rev)
		}
	case RawDiffPatch:
		if commit.ParentsCount() == 0 {
			cmd.AddArgs("format-patch", "--no-signature", "--stdout", "--root", rev)
		} else {
			c, _ := commit.Parent(0)
			query := fmt.Sprintf("%s...%s", rev, c.id.String())
			cmd.AddArgs("format-patch", "--no-signature", "--stdout", query)
		}
	default:
		return fmt.Errorf("invalid diffType: %s", diffType)
	}

	stderr := new(bytes.Buffer)
	if err = cmd.RunInDirPipeline(r.path, w, stderr); err != nil {
		return concatenateError(err, stderr.String())
	}
	return nil
}
