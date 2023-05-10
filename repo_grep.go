// Copyright 2022 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// GrepOptions contains optional arguments for grep search over repository files.
//
// Docs: https://git-scm.com/docs/git-grep
type GrepOptions struct {
	// The tree to run the search. Defaults to "HEAD".
	Tree string
	// Limits the search to files in the specified pathspec.
	Pathspec string
	// Whether to do case insensitive search.
	IgnoreCase bool
	// Whether to match the pattern only at word boundaries.
	WordRegexp bool
	// Whether use extended regular expressions.
	ExtendedRegexp bool
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// GrepResult represents a single result from a grep search.
type GrepResult struct {
	// The tree of the file that matched, e.g. "HEAD".
	Tree string
	// The path of the file that matched.
	Path string
	// The line number of the match.
	Line int
	// The 1-indexed column number of the match.
	Column int
	// The text of the line that matched.
	Text string
}

func parseGrepLine(line string) (*GrepResult, error) {
	r := &GrepResult{}
	sp := strings.SplitN(line, ":", 5)
	var n int
	switch len(sp) {
	case 4:
		// HEAD
		r.Tree = "HEAD"
	case 5:
		// Tree included
		r.Tree = sp[0]
		n++
	default:
		return nil, fmt.Errorf("invalid grep line: %s", line)
	}
	r.Path = sp[n]
	n++
	r.Line, _ = strconv.Atoi(sp[n])
	n++
	r.Column, _ = strconv.Atoi(sp[n])
	n++
	r.Text = sp[n]
	return r, nil
}

// Grep returns the results of a grep search in the repository.
func (r *Repository) Grep(pattern string, opts ...GrepOptions) []*GrepResult {
	var opt GrepOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt.Tree == "" {
		opt.Tree = "HEAD"
	}

	cmd := NewCommand("grep").
		AddOptions(opt.CommandOptions).
		// Display full-name, line number and column number
		AddArgs("--full-name", "--line-number", "--column")
	if opt.IgnoreCase {
		cmd.AddArgs("--ignore-case")
	}
	if opt.WordRegexp {
		cmd.AddArgs("--word-regexp")
	}
	if opt.ExtendedRegexp {
		cmd.AddArgs("--extended-regexp")
	}
	cmd.AddArgs(pattern, opt.Tree)
	if opt.Pathspec != "" {
		cmd.AddArgs("--", opt.Pathspec)
	}

	stdout, err := cmd.RunInDirWithTimeout(opt.Timeout, r.path)
	if err != nil {
		return nil
	}

	var results []*GrepResult
	// Normalize line endings
	lines := strings.Split(strings.ReplaceAll(string(stdout), "\r", ""), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		r, err := parseGrepLine(line)
		if err == nil {
			results = append(results, r)
		}
	}
	return results
}
