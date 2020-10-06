// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var oneRowBlame = `2c49687c5b06776a44e2a8c0635428f647909472 3 3 4
author ᴜɴᴋɴᴡᴏɴ
author-mail <u@gogs.io>
author-time 1585383299
author-tz +0800
committer GitHub
committer-mail <noreply@github.com>
committer-time 1585383299
committer-tz +0800
summary ci: migrate from Travis to GitHub Actions (#50)
previous 0d17b78404b7432905a58a235d875e9d28969ee3 README.md
filename README.md
	[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/gogs/git-module/Go?logo=github&style=for-the-badge)](https://github.com/gogs/git-module/actions?query=workflow%3AGo)
`

var twoRowsBlame = `f29bce1e3a666c02175d080892be185405dd3af4 1 1 2
author Unknwon
author-mail <u@gogs.io>
author-time 1573967409
author-tz -0800
committer Unknwon
committer-mail <u@gogs.io>
committer-time 1573967409
committer-tz -0800
summary README:  update badges
previous 065699e51f42559ab0c3ad22c1f2c789b2def8fb README.md
filename README.md
	# Git Module 
f29bce1e3a666c02175d080892be185405dd3af4 2 2
	
`

var commit1 = &Commit{
	ID:      MustIDFromString("f29bce1e3a666c02175d080892be185405dd3af4"),
	Message: "README:  update badges",
	Author: &Signature{
		Name:  "Unknwon",
		Email: "<u@gogs.io>",
		When:  time.Unix(1573967409, 0).In(time.FixedZone("Fixed", -8*60*60)),
	},
	Committer: &Signature{
		Name:  "Unknwon",
		Email: "<u@gogs.io>",
		When:  time.Unix(1573967409, 0).In(time.FixedZone("Fixed", -8*60*60)),
	},
	parents: []*SHA1{MustIDFromString("065699e51f42559ab0c3ad22c1f2c789b2def8fb")},
}

var commit2 = &Commit{
	ID:      MustIDFromString("2c49687c5b06776a44e2a8c0635428f647909472"),
	Message: "ci: migrate from Travis to GitHub Actions (#50)",
	Author: &Signature{
		Name:  "ᴜɴᴋɴᴡᴏɴ",
		Email: "<u@gogs.io>",
		When:  time.Unix(1585383299, 0).In(time.FixedZone("Fixed", 8*60*60)),
	},
	Committer: &Signature{
		Name:  "GitHub",
		Email: "<noreply@github.com>",
		When:  time.Unix(1585383299, 0).In(time.FixedZone("Fixed", 8*60*60)),
	},
	parents: []*SHA1{MustIDFromString("0d17b78404b7432905a58a235d875e9d28969ee3")},
}

func TestOneRowBlame(t *testing.T) {
	blame, _ := BlameContent([]byte(oneRowBlame))
	var expect = createBlame()

	expect.commits[3] = commit2

	assert.Equal(t, expect, blame)
}

func TestMultipleRowsBlame(t *testing.T) {
	blame, _ := BlameContent([]byte(twoRowsBlame + oneRowBlame))
	var expect = createBlame()

	expect.commits[1] = commit1
	expect.commits[2] = commit1
	expect.commits[3] = commit2

	assert.Equal(t, expect, blame)
}

func TestRepository_BlameFile(t *testing.T) {
	blame, _ := testrepo.BlameFile("master", "pom.xml")
	assert.Greater(t, len(blame.commits), 0)
}

func TestRepository_BlameNotExistFile(t *testing.T) {
	_, err := testrepo.BlameFile("master", "0")
	assert.Error(t, err)
}
