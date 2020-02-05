// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import "bytes"

// Tag contains information of a Git tag.
type Tag struct {
	typ      ObjectType
	id       *SHA1
	commitID *SHA1 // The ID of the underlying commit
	refspec  string
	tagger   *Signature
	message  string

	repo *Repository
}

func (tag *Tag) Commit(opts ...CatFileCommitOptions) (*Commit, error) {
	return tag.repo.CatFileCommit(tag.commitID.String(), opts...)
}

// parseTag parses tag information from the (uncompressed) raw data of the tag object.
// It assumes "\n\n" separates the header from the rest of the message.
func parseTag(data []byte) (*Tag, error) {
	tag := new(Tag)
	// we now have the contents of the commit object. Let's investigate.
	nextline := 0
l:
	for {
		eol := bytes.IndexByte(data[nextline:], '\n')
		switch {
		case eol > 0:
			line := data[nextline : nextline+eol]
			spacepos := bytes.IndexByte(line, ' ')
			reftype := line[:spacepos]
			switch string(reftype) {
			case "object":
				id, err := NewIDFromString(string(line[spacepos+1:]))
				if err != nil {
					return nil, err
				}
				tag.commitID = id
			case "type":
				// A commit can have one or more parents
				tag.typ = ObjectType(line[spacepos+1:])
			case "tagger":
				sig, err := parseSignature(line[spacepos+1:])
				if err != nil {
					return nil, err
				}
				tag.tagger = sig
			}
			nextline += eol + 1
		case eol == 0:
			tag.message = string(data[nextline+1:])
			break l
		default:
			break l
		}
	}
	return tag, nil
}
