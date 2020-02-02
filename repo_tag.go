// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"fmt"
	"strings"

	goversion "github.com/mcuadros/go-version"
)

func (r *Repository) CreateTag(name, revision string) error {
	_, err := NewCommand("tag", name, revision).RunInDir(r.path)
	return err
}

func (r *Repository) getTag(id *SHA1) (*Tag, error) {
	t, ok := r.cachedTags.Get(id.String())
	if ok {
		log("Cached tag hit: %s", id)
		return t.(*Tag), nil
	}

	// Get tag type
	tp, err := NewCommand("cat-file", "-t", id.String()).RunInDir(r.path)
	if err != nil {
		return nil, err
	}
	tp = bytes.TrimSpace(tp)

	// Tag is a commit.
	if ObjectType(tp) == ObjectCommit {
		tag := &Tag{
			ID:       id,
			CommitID: id,
			Type:     string(ObjectCommit),
			repo:     r,
		}

		r.cachedTags.Set(id.String(), tag)
		return tag, nil
	}

	// Tag with message.
	data, err := NewCommand("cat-file", "-p", id.String()).RunInDir(r.path)
	if err != nil {
		return nil, err
	}

	tag, err := parseTagData(data)
	if err != nil {
		return nil, err
	}

	tag.ID = id
	tag.repo = r

	r.cachedTags.Set(id.String(), tag)
	return tag, nil
}

// GetTag returns a Git tag by given name.
func (r *Repository) GetTag(name string) (*Tag, error) {
	stdout, err := NewCommand("show-ref", "--tags", name).RunInDir(r.path)
	if err != nil {
		return nil, err
	}

	id, err := NewIDFromString(strings.Split(string(stdout), " ")[0])
	if err != nil {
		return nil, err
	}

	tag, err := r.getTag(id)
	if err != nil {
		return nil, err
	}
	tag.Name = name
	return tag, nil
}

// GetTags returns all tags of the repository.
func (r *Repository) GetTags() ([]string, error) {
	cmd := NewCommand("tag", "-l")
	if goversion.Compare(gitVersion, "2.4.9", ">=") {
		cmd.AddArgs("--sort=-creatordate")
	}

	stdout, err := cmd.RunInDir(r.path)
	if err != nil {
		return nil, err
	}

	tags := strings.Split(string(stdout), "\n")
	tags = tags[:len(tags)-1]

	if goversion.Compare(gitVersion, "2.4.9", "<") {
		goversion.Sort(tags)

		// Reverse order
		for i := 0; i < len(tags)/2; i++ {
			j := len(tags) - i - 1
			tags[i], tags[j] = tags[j], tags[i]
		}
	}

	return tags, nil
}

type TagsResult struct {
	// Indicates whether results include the latest tag.
	HasLatest bool
	// If results do not include the latest tag, a indicator 'after' to go back.
	PreviousAfter string
	// Indicates whether results include the oldest tag.
	ReachEnd bool
	// List of returned tags.
	Tags []string
}

// GetTagsAfter returns list of tags 'after' (exlusive) given tag.
func (r *Repository) GetTagsAfter(after string, limit int) (*TagsResult, error) {
	allTags, err := r.GetTags()
	if err != nil {
		return nil, fmt.Errorf("GetTags: %v", err)
	}

	if limit < 0 {
		limit = 0
	}

	numAllTags := len(allTags)
	if len(after) == 0 && limit == 0 {
		return &TagsResult{
			HasLatest: true,
			ReachEnd:  true,
			Tags:      allTags,
		}, nil
	} else if len(after) == 0 && limit > 0 {
		endIdx := limit
		if limit >= numAllTags {
			endIdx = numAllTags
		}
		return &TagsResult{
			HasLatest: true,
			ReachEnd:  limit >= numAllTags,
			Tags:      allTags[:endIdx],
		}, nil
	}

	previousAfter := ""
	hasMatch := false
	tags := make([]string, 0, len(allTags))
	for i := range allTags {
		if hasMatch {
			tags = allTags[i:]
			break
		}
		if allTags[i] == after {
			hasMatch = true
			if limit > 0 && i-limit >= 0 {
				previousAfter = allTags[i-limit]
			}
			continue
		}
	}

	if !hasMatch {
		tags = allTags
	}

	// If all tags after match is equal to the limit, it reaches the oldest tag as well.
	if limit == 0 || len(tags) <= limit {
		return &TagsResult{
			HasLatest:     !hasMatch,
			PreviousAfter: previousAfter,
			ReachEnd:      true,
			Tags:          tags,
		}, nil
	}
	return &TagsResult{
		HasLatest:     !hasMatch,
		PreviousAfter: previousAfter,
		Tags:          tags[:limit],
	}, nil
}

// DeleteTag deletes a tag from the repository
func (r *Repository) DeleteTag(name string) error {
	cmd := NewCommand("tag", "-d")

	cmd.AddArgs(name)
	_, err := cmd.RunInDir(r.path)

	return err
}
