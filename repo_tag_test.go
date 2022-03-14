// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepository_Tag(t *testing.T) {
	tests := []struct {
		name   string
		opt    TagOptions
		expTag *Tag
	}{
		{
			name: "v1.0.0",
			expTag: &Tag{
				typ:      ObjectCommit,
				id:       MustIDFromString("0eedd79eba4394bbef888c804e899731644367fe"),
				commitID: MustIDFromString("0eedd79eba4394bbef888c804e899731644367fe"),
				refspec:  "refs/tags/v1.0.0",
			},
		}, {
			name: "v1.1.0",
			expTag: &Tag{
				typ:      ObjectTag,
				id:       MustIDFromString("b39c8508bbc4b00ad2e24d358012ea123bcafd8d"),
				commitID: MustIDFromString("0eedd79eba4394bbef888c804e899731644367fe"),
				refspec:  "refs/tags/v1.1.0",
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			tag, err := testrepo.Tag(test.name, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expTag.Type(), tag.Type())
			assert.Equal(t, test.expTag.ID().String(), tag.ID().String())
			assert.Equal(t, test.expTag.CommitID().String(), tag.CommitID().String())
			assert.Equal(t, test.expTag.Refspec(), tag.Refspec())
		})
	}
}

func TestRepository_Tags(t *testing.T) {
	// Make sure it does not blow up
	tags, err := testrepo.Tags(TagsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEmpty(t, tags)
}

func TestRepository_CreateTag(t *testing.T) {
	r, cleanup, err := setupTempRepo()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	assert.False(t, r.HasReference(RefsTags+"v2.0.0"))

	err = r.CreateTag("v2.0.0", "master", CreateTagOptions{})
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, r.HasReference(RefsTags+"v2.0.0"))
}

func TestRepository_CreateAnnotatedTag(t *testing.T) {
	r, cleanup, err := setupTempRepo()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	assert.False(t, r.HasReference(RefsTags+"v2.0.0"))

	err = r.CreateTag("v2.0.0", "master", CreateTagOptions{
		Annotated: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, r.HasReference(RefsTags+"v2.0.0"))

	tag, err := r.Tag("v2.0.0")
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, tag.tagger)
}

func TestRepository_DeleteTag(t *testing.T) {
	r, cleanup, err := setupTempRepo()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	assert.True(t, r.HasReference(RefsTags+"v1.0.0"))

	err = r.DeleteTag("v1.0.0", DeleteTagOptions{})
	if err != nil {
		t.Fatal(err)
	}

	assert.False(t, r.HasReference(RefsTags+"v1.0.0"))
}
