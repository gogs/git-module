// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	goversion "github.com/mcuadros/go-version"
)

// parseTag parses tag information from the (uncompressed) raw data of the tag
// object. It assumes "\n\n" separates the header from the rest of the message.
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

// getTag returns a tag by given SHA1 hash.
func (r *Repository) getTag(timeout time.Duration, id *SHA1) (*Tag, error) {
	t, ok := r.cachedTags.Get(id.String())
	if ok {
		log("Cached tag hit: %s", id)
		return t.(*Tag), nil
	}

	// Check tag type
	typ, err := r.CatFileType(id.String(), CatFileTypeOptions{Timeout: timeout})
	if err != nil {
		return nil, err
	}

	var tag *Tag
	switch typ {
	case ObjectCommit: // Tag is a commit
		tag = &Tag{
			typ:      ObjectCommit,
			id:       id,
			commitID: id,
			repo:     r,
		}

	case ObjectTag: // Tag is an annotation
		data, err := NewCommand("cat-file", "-p", id.String()).RunInDir(r.path)
		if err != nil {
			return nil, err
		}

		tag, err = parseTag(data)
		if err != nil {
			return nil, err
		}
		tag.typ = ObjectTag
		tag.id = id
		tag.repo = r
	default:
		return nil, fmt.Errorf("unsupported tag type: %s", ObjectType(typ))
	}

	r.cachedTags.Set(id.String(), tag)
	return tag, nil
}

// TagOptions contains optional arguments for getting a tag.
//
// Docs: https://git-scm.com/docs/git-cat-file
type TagOptions struct {
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// Tag returns a Git tag by given name, e.g. "v1.0.0".
func (r *Repository) Tag(name string, opts ...TagOptions) (*Tag, error) {
	var opt TagOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	refsepc := RefsTags + name
	refs, err := r.ShowRef(ShowRefOptions{
		Tags:           true,
		Patterns:       []string{refsepc},
		Timeout:        opt.Timeout,
		CommandOptions: opt.CommandOptions,
	})
	if err != nil {
		return nil, err
	} else if len(refs) == 0 {
		return nil, ErrReferenceNotExist
	}

	id, err := NewIDFromString(refs[0].ID)
	if err != nil {
		return nil, err
	}

	tag, err := r.getTag(opt.Timeout, id)
	if err != nil {
		return nil, err
	}
	tag.refspec = refsepc
	return tag, nil
}

// TagsOptions contains optional arguments for listing tags.
//
// Docs: https://git-scm.com/docs/git-tag#Documentation/git-tag.txt---list
type TagsOptions struct {
	// SortKet sorts tags with provided tag key, optionally prefixed with '-' to sort tags in descending order.
	SortKey string
	// Pattern filters tags matching the specified pattern.
	Pattern string
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// RepoTags returns a list of tags of the repository in given path.
func RepoTags(repoPath string, opts ...TagsOptions) ([]string, error) {
	var opt TagsOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	version, err := BinVersion()
	if err != nil {
		return nil, err
	}

	cmd := NewCommand("tag", "--list").AddOptions(opt.CommandOptions)

	var sorted bool
	if opt.SortKey != "" {
		cmd.AddArgs("--sort=" + opt.SortKey)
		sorted = true
	} else if goversion.Compare(version, "2.4.9", ">=") {
		cmd.AddArgs("--sort=-creatordate")
		sorted = true
	}

	if opt.Pattern != "" {
		cmd.AddArgs(opt.Pattern)
	}

	stdout, err := cmd.RunInDirWithTimeout(opt.Timeout, repoPath)
	if err != nil {
		return nil, err
	}

	tags := strings.Split(string(stdout), "\n")
	tags = tags[:len(tags)-1]

	if !sorted {
		goversion.Sort(tags)

		// Reverse order
		for i := 0; i < len(tags)/2; i++ {
			j := len(tags) - i - 1
			tags[i], tags[j] = tags[j], tags[i]
		}
	}

	return tags, nil
}

// Tags returns a list of tags of the repository.
func (r *Repository) Tags(opts ...TagsOptions) ([]string, error) {
	return RepoTags(r.path, opts...)
}

// CreateTagOptions contains optional arguments for creating a tag.
//
// Docs: https://git-scm.com/docs/git-tag
type CreateTagOptions struct {
	// Annotated marks a tag as annotated rather than lightweight.
	Annotated bool
	// Message specifies a tagging message for the annotated tag. It is ignored when tag is not annotated.
	Message string
	// Author is the author of the tag. It is ignored when tag is not annotated.
	Author *Signature
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// CreateTag creates a new tag on given revision.
func (r *Repository) CreateTag(name, rev string, opts ...CreateTagOptions) error {
	var opt CreateTagOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("tag").AddOptions(opt.CommandOptions)
	if opt.Annotated {
		cmd.AddArgs("-a", name)
		cmd.AddArgs("--message", opt.Message)
		if opt.Author != nil {
			cmd.AddCommitter(opt.Author)
		}
	} else {
		cmd.AddArgs(name)
	}

	cmd.AddArgs(rev)

	_, err := cmd.RunInDirWithTimeout(opt.Timeout, r.path)
	return err
}

// DeleteTagOptions contains optional arguments for deleting a tag.
//
// Docs: https://git-scm.com/docs/git-tag#Documentation/git-tag.txt---delete
type DeleteTagOptions struct {
	// The timeout duration before giving up for each shell command execution.
	// The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// DeleteTag deletes a tag from the repository.
func (r *Repository) DeleteTag(name string, opts ...DeleteTagOptions) error {
	var opt DeleteTagOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	_, err := NewCommand("tag", "--delete", name).RunInDirWithTimeout(opt.Timeout, r.path)
	return err
}
