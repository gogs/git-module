package git

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommit(t *testing.T) {
	ctx := context.Background()
	c, err := testrepo.CatFileCommit(ctx, testrepoMarks[22].String())
	if err != nil {
		t.Fatal(err)
	}
	t.Run("ID", func(t *testing.T) {
		assert.Equal(t, testrepoMarks[22].String(), c.ID.String())
	})

	t.Run("Summary", func(t *testing.T) {
		assert.Equal(t, "Merge pull request #35 from githubtraining/travis-yml-docker", c.Summary())
	})
}

func TestCommit_Parent(t *testing.T) {
	ctx := context.Background()
	c, err := testrepo.CatFileCommit(ctx, testrepoMarks[22].String())
	if err != nil {
		t.Fatal(err)
	}

	t.Run("ParentsCount", func(t *testing.T) {
		assert.Equal(t, 2, c.ParentsCount())
	})

	t.Run("Parent", func(t *testing.T) {
		t.Run("no such parent", func(t *testing.T) {
			_, err := c.Parent(ctx, c.ParentsCount()+1)
			assert.Equal(t, ErrParentNotExist, err)
		})

		tests := []struct {
			n           int
			expParentID string
		}{
			{
				n:           0,
				expParentID: testrepoMarks[20].String(),
			},
			{
				n:           1,
				expParentID: testrepoMarks[21].String(),
			},
		}
		for _, test := range tests {
			t.Run("", func(t *testing.T) {
				p, err := c.Parent(ctx, test.n)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, test.expParentID, p.ID.String())
			})
		}
	})
}

func TestCommit_CommitByPath(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		id          string
		opt         CommitByRevisionOptions
		expCommitID string
	}{
		{
			id: testrepoMarks[12].String(),
			opt: CommitByRevisionOptions{
				Path: "", // No path gets back to the commit itself
			},
			expCommitID: testrepoMarks[12].String(),
		},
		{
			id: testrepoMarks[12].String(),
			opt: CommitByRevisionOptions{
				Path: "resources/labels.properties",
			},
			expCommitID: testrepoMarks[1].String(),
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			c, err := testrepo.CatFileCommit(ctx, test.id)
			if err != nil {
				t.Fatal(err)
			}

			cc, err := c.CommitByPath(ctx, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expCommitID, cc.ID.String())
		})
	}
}

// commitsToIDs returns a list of IDs for given commits.
func commitsToIDs(commits []*Commit) []string {
	ids := make([]string, len(commits))
	for i := range commits {
		ids[i] = commits[i].ID.String()
	}
	return ids
}

func TestCommit_CommitsByPage(t *testing.T) {
	ctx := context.Background()
	// There are at most 5 commits can be used for pagination before this commit.
	c, err := testrepo.CatFileCommit(ctx, testrepoMarks[5].String())
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		page         int
		size         int
		opt          CommitsByPageOptions
		expCommitIDs []string
	}{
		{
			page: 0,
			size: 2,
			expCommitIDs: []string{
				testrepoMarks[5].String(),
				testrepoMarks[4].String(),
			},
		},
		{
			page: 1,
			size: 2,
			expCommitIDs: []string{
				testrepoMarks[5].String(),
				testrepoMarks[4].String(),
			},
		},
		{
			page: 2,
			size: 2,
			expCommitIDs: []string{
				testrepoMarks[3].String(),
				testrepoMarks[2].String(),
			},
		},
		{
			page: 3,
			size: 2,
			expCommitIDs: []string{
				testrepoMarks[1].String(),
			},
		},
		{
			page:         4,
			size:         2,
			expCommitIDs: []string{},
		},

		{
			page: 2,
			size: 2,
			opt: CommitsByPageOptions{
				Path: "src",
			},
			expCommitIDs: []string{
				testrepoMarks[1].String(),
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			commits, err := c.CommitsByPage(ctx, test.page, test.size, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expCommitIDs, commitsToIDs(commits))
		})
	}
}

func TestCommit_SearchCommits(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		id           string
		pattern      string
		opt          SearchCommitsOptions
		expCommitIDs []string
	}{
		{
			id:      testrepoMarks[12].String(),
			pattern: "",
			expCommitIDs: []string{
				testrepoMarks[12].String(),
				testrepoMarks[11].String(),
				testrepoMarks[10].String(),
				testrepoMarks[9].String(),
				testrepoMarks[8].String(),
				testrepoMarks[7].String(),
				testrepoMarks[6].String(),
				testrepoMarks[5].String(),
				testrepoMarks[4].String(),
				testrepoMarks[3].String(),
				testrepoMarks[2].String(),
				testrepoMarks[1].String(),
			},
		},
		{
			id:      testrepoMarks[12].String(),
			pattern: "",
			opt: SearchCommitsOptions{
				MaxCount: 3,
			},
			expCommitIDs: []string{
				testrepoMarks[12].String(),
				testrepoMarks[11].String(),
				testrepoMarks[10].String(),
			},
		},

		{
			id:      testrepoMarks[12].String(),
			pattern: "feature",
			expCommitIDs: []string{
				testrepoMarks[12].String(),
				testrepoMarks[10].String(),
			},
		},
		{
			id:      testrepoMarks[12].String(),
			pattern: "feature",
			opt: SearchCommitsOptions{
				MaxCount: 1,
			},
			expCommitIDs: []string{
				testrepoMarks[12].String(),
			},
		},

		{
			id:      testrepoMarks[12].String(),
			pattern: "add.*",
			opt: SearchCommitsOptions{
				Path: "src",
			},
			expCommitIDs: []string{
				testrepoMarks[10].String(),
				testrepoMarks[9].String(),
				testrepoMarks[6].String(),
				testrepoMarks[2].String(),
				testrepoMarks[1].String(),
			},
		},
		{
			id:      testrepoMarks[12].String(),
			pattern: "add.*",
			opt: SearchCommitsOptions{
				MaxCount: 2,
				Path:     "src",
			},
			expCommitIDs: []string{
				testrepoMarks[10].String(),
				testrepoMarks[9].String(),
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			c, err := testrepo.CatFileCommit(ctx, test.id)
			if err != nil {
				t.Fatal(err)
			}

			commits, err := c.SearchCommits(ctx, test.pattern, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expCommitIDs, commitsToIDs(commits))
		})
	}
}

func TestCommit_ShowNameStatus(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		id        string
		opt       ShowNameStatusOptions
		expStatus *NameStatus
	}{
		{
			id: testrepoMarks[1].String(),
			expStatus: &NameStatus{
				Added: []string{
					"README.txt",
					"resources/labels.properties",
					"src/Main.groovy",
				},
			},
		},
		{
			id: testrepoMarks[2].String(),
			expStatus: &NameStatus{
				Modified: []string{
					"src/Main.groovy",
				},
			},
		},
		{
			id: testrepoMarks[3].String(),
			expStatus: &NameStatus{
				Added: []string{
					"src/Square.groovy",
				},
				Modified: []string{
					"src/Main.groovy",
				},
			},
		},
		{
			id: testrepoMarks[27].String(),
			expStatus: &NameStatus{
				Removed: []string{
					"fix.txt",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			c, err := testrepo.CatFileCommit(ctx, test.id)
			if err != nil {
				t.Fatal(err)
			}

			status, err := c.ShowNameStatus(ctx, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expStatus, status)
		})
	}
}

func TestCommit_CommitsCount(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		id       string
		opt      RevListCountOptions
		expCount int64
	}{
		{
			id:       testrepoMarks[1].String(),
			expCount: 1,
		},
		{
			id:       testrepoMarks[5].String(),
			expCount: 5,
		},
		{
			id:       testrepoMarks[27].String(),
			expCount: 27,
		},

		{
			id: testrepoMarks[21].String(),
			opt: RevListCountOptions{
				Path: "README.txt",
			},
			expCount: 3,
		},
		{
			id: testrepoMarks[21].String(),
			opt: RevListCountOptions{
				Path: "resources",
			},
			expCount: 1,
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			c, err := testrepo.CatFileCommit(ctx, test.id)
			if err != nil {
				t.Fatal(err)
			}

			count, err := c.CommitsCount(ctx, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expCount, count)
		})
	}
}

func TestCommit_FilesChangedAfter(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		id       string
		after    string
		opt      DiffNameOnlyOptions
		expFiles []string
	}{
		{
			id:       testrepoMarks[27].String(),
			after:    testrepoMarks[26].String(),
			expFiles: []string{"fix.txt"},
		},
		{
			id:       testrepoMarks[27].String(),
			after:    testrepoMarks[24].String(),
			expFiles: []string{"fix.txt", "pom.xml", "src/test/java/com/github/AppTest.java"},
		},

		{
			id:    testrepoMarks[27].String(),
			after: testrepoMarks[24].String(),
			opt: DiffNameOnlyOptions{
				Path: "src",
			},
			expFiles: []string{"src/test/java/com/github/AppTest.java"},
		},
		{
			id:    testrepoMarks[27].String(),
			after: testrepoMarks[24].String(),
			opt: DiffNameOnlyOptions{
				Path: "resources",
			},
			expFiles: []string{},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			c, err := testrepo.CatFileCommit(ctx, test.id)
			if err != nil {
				t.Fatal(err)
			}

			files, err := c.FilesChangedAfter(ctx, test.after, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expFiles, files)
		})
	}
}

func TestCommit_CommitsAfter(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		id           string
		after        string
		opt          RevListOptions
		expCommitIDs []string
	}{
		{
			id:    testrepoMarks[27].String(),
			after: testrepoMarks[24].String(),
			expCommitIDs: []string{
				testrepoMarks[27].String(),
				testrepoMarks[26].String(),
				testrepoMarks[25].String(),
			},
		},
		{
			id:    testrepoMarks[27].String(),
			after: testrepoMarks[24].String(),
			opt: RevListOptions{
				Path: "src",
			},
			expCommitIDs: []string{
				testrepoMarks[25].String(),
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			c, err := testrepo.CatFileCommit(ctx, test.id)
			if err != nil {
				t.Fatal(err)
			}

			commits, err := c.CommitsAfter(ctx, test.after, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expCommitIDs, commitsToIDs(commits))
		})
	}
}

func TestCommit_Ancestors(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		id           string
		opt          LogOptions
		expCommitIDs []string
	}{
		{
			id: testrepoMarks[12].String(),
			opt: LogOptions{
				MaxCount: 3,
			},
			expCommitIDs: []string{
				testrepoMarks[11].String(),
				testrepoMarks[10].String(),
				testrepoMarks[9].String(),
			},
		},
		{
			id:           testrepoMarks[1].String(),
			expCommitIDs: []string{},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			c, err := testrepo.CatFileCommit(ctx, test.id)
			if err != nil {
				t.Fatal(err)
			}

			commits, err := c.Ancestors(ctx, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expCommitIDs, commitsToIDs(commits))
		})
	}
}

func TestCommit_IsImageFile(t *testing.T) {
	ctx := context.Background()

	t.Run("not a blob", func(t *testing.T) {
		c, err := testrepo.CatFileCommit(ctx, testrepoMarks[29].String())
		if err != nil {
			t.Fatal(err)
		}

		isImage, err := c.IsImageFile(ctx, "gogs/docs-api")
		if err != nil {
			t.Fatal(err)
		}
		assert.False(t, isImage)
	})

	tests := []struct {
		id     string
		name   string
		expVal bool
	}{
		{
			id:     testrepoMarks[28].String(),
			name:   "README.txt",
			expVal: false,
		},
		{
			id:     testrepoMarks[28].String(),
			name:   "img/sourcegraph.png",
			expVal: true,
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			c, err := testrepo.CatFileCommit(ctx, test.id)
			if err != nil {
				t.Fatal(err)
			}

			isImage, err := c.IsImageFile(ctx, test.name)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expVal, isImage)
		})
	}
}

func TestCommit_IsImageFileByIndex(t *testing.T) {
	ctx := context.Background()

	t.Run("not a blob", func(t *testing.T) {
		c, err := testrepo.CatFileCommit(ctx, testrepoMarks[29].String())
		if err != nil {
			t.Fatal(err)
		}

		isImage, err := c.IsImageFileByIndex(ctx, testrepoMarks[40].String()) // "gogs"
		if err != nil {
			t.Fatal(err)
		}
		assert.False(t, isImage)
	})

	tests := []struct {
		id     string
		index  string
		expVal bool
	}{
		{
			id:     testrepoMarks[28].String(),
			index:  testrepoMarks[39].String(), // "README.txt"
			expVal: false,
		},
		{
			id:     testrepoMarks[28].String(),
			index:  testrepoMarks[49].String(), // "img/sourcegraph.png"
			expVal: true,
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			c, err := testrepo.CatFileCommit(ctx, test.id)
			if err != nil {
				t.Fatal(err)
			}

			isImage, err := c.IsImageFileByIndex(ctx, test.index)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expVal, isImage)
		})
	}
}
