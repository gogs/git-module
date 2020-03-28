// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"fmt"
	"time"
)

// UnescapeChars reverses escaped characters.
func UnescapeChars(in []byte) []byte {
	if bytes.ContainsAny(in, "\\\t") {
		return in
	}

	out := bytes.Replace(in, escapedSlash, regularSlash, -1)
	out = bytes.Replace(out, escapedTab, regularTab, -1)
	return out
}

// Predefine []byte variables to avoid runtime allocations.
var (
	escapedSlash = []byte(`\\`)
	regularSlash = []byte(`\`)
	escapedTab   = []byte(`\t`)
	regularTab   = []byte("\t")
)

// parseTree parses tree information from the (uncompressed) raw data of the tree object.
func parseTree(t *Tree, data []byte) ([]*TreeEntry, error) {
	entries := make([]*TreeEntry, 0, 10)
	l := len(data)
	pos := 0
	for pos < l {
		entry := new(TreeEntry)
		entry.parent = t
		step := 6
		switch string(data[pos : pos+step]) {
		case "100644", "100664":
			entry.mode = EntryBlob
			entry.typ = ObjectBlob
		case "100755":
			entry.mode = EntryExec
			entry.typ = ObjectBlob
		case "120000":
			entry.mode = EntrySymlink
			entry.typ = ObjectBlob
		case "160000":
			entry.mode = EntryCommit
			entry.typ = ObjectCommit

			step = 8
		case "040000":
			entry.mode = EntryTree
			entry.typ = ObjectTree
		default:
			return nil, fmt.Errorf("unknown type: %v", string(data[pos:pos+step]))
		}
		pos += step + 6 // Skip string type of entry type.

		step = 40
		id, err := NewIDFromString(string(data[pos : pos+step]))
		if err != nil {
			return nil, err
		}
		entry.id = id
		pos += step + 1 // Skip half of SHA1.

		step = bytes.IndexByte(data[pos:], '\n')

		// In case entry name is surrounded by double quotes(it happens only in git-shell).
		if data[pos] == '"' {
			entry.name = string(UnescapeChars(data[pos+1 : pos+step-1]))
		} else {
			entry.name = string(data[pos : pos+step])
		}

		pos += step + 1
		entries = append(entries, entry)
	}
	return entries, nil
}

// LsTreeOptions contains optional arguments for listing trees.
// Docs: https://git-scm.com/docs/git-ls-tree
type LsTreeOptions struct {
	// The timeout duration before giving up for each shell command execution.
	// The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// LsTree returns the tree object in the repository by given revision.
func (r *Repository) LsTree(rev string, opts ...LsTreeOptions) (*Tree, error) {
	var opt LsTreeOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	var err error
	rev, err = r.RevParse(rev, RevParseOptions{Timeout: opt.Timeout}) //nolint
	if err != nil {
		return nil, err
	}
	t := &Tree{
		id:   MustIDFromString(rev),
		repo: r,
	}

	stdout, err := NewCommand("ls-tree", rev).RunInDirWithTimeout(opt.Timeout, r.path)
	if err != nil {
		return nil, err
	}

	t.entries, err = parseTree(t, stdout)
	if err != nil {
		return nil, err
	}

	return t, nil
}
