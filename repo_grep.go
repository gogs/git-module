package git

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type GrepOptions struct {
	// TreeID to search in
	TreeID string
	// Limits the search to files in the specified pathspec
	Pathspec string
	// Case insensitive search.
	IgnoreCase bool
	// Match the pattern only at word boundaries.
	WordMatch bool
	// Whether or not to use extended regular expressions.
	ExtendedRegex bool
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// GrepResult represents a single result from a grep search.
type GrepResult struct {
	// The TreeID of the file that matched. Could be `HEAD` or a tree ID.
	TreeID string
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
		// HEAD tree ID
		r.TreeID = "HEAD"
	case 5:
		// Tree ID included
		r.TreeID = sp[0]
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
func (r *Repository) Grep(pattern string, opts ...GrepOptions) ([]*GrepResult, error) {
	var opt GrepOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt.TreeID == "" {
		opt.TreeID = "HEAD"
	}

	cmd := NewCommand("grep").
		AddOptions(opt.CommandOptions).
		// Result full-name, line number & column number
		AddArgs("--full-name", "--line-number", "--column")
	if opt.IgnoreCase {
		cmd.AddArgs("-i")
	}
	if opt.WordMatch {
		cmd.AddArgs("-w")
	}
	if opt.ExtendedRegex {
		cmd.AddArgs("-E")
	}
	cmd.AddArgs(pattern, opt.TreeID)
	if opt.Pathspec != "" {
		cmd.AddArgs("--", opt.Pathspec)
	}

	results := make([]*GrepResult, 0)
	stdout, err := cmd.RunInDirWithTimeout(opt.Timeout, r.path)
	if err == nil && len(stdout) > 0 {
		// normalize line endings
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
	}

	return results, nil
}
