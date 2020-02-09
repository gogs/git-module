package git

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCommit(t *testing.T) {
	c, err := testrepo.CatFileCommit("435ffceb7ba576c937e922766e37d4f7abdcc122")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("ID", func(t *testing.T) {
		assert.Equal(t, "435ffceb7ba576c937e922766e37d4f7abdcc122", c.ID().String())
	})

	author := &Signature{
		Name:  "Jordan McCullough",
		Email: "jordan@github.com",
		When:  time.Unix(1415213395, 0),
	}
	t.Run("Author", func(t *testing.T) {
		assert.Equal(t, author.Name, c.Author().Name)
		assert.Equal(t, author.Email, c.Author().Email)
		assert.Equal(t, author.When.Unix(), c.Author().When.Unix())
	})

	t.Run("Committer", func(t *testing.T) {
		assert.Equal(t, author.Name, c.Committer().Name)
		assert.Equal(t, author.Email, c.Committer().Email)
		assert.Equal(t, author.When.Unix(), c.Committer().When.Unix())
	})

	t.Run("Message", func(t *testing.T) {
		message := `Merge pull request #35 from githubtraining/travis-yml-docker

Add special option flag for Travis Docker use case`
		assert.Equal(t, message, c.Message())
	})

	t.Run("Summary", func(t *testing.T) {
		assert.Equal(t, "Merge pull request #35 from githubtraining/travis-yml-docker", c.Summary())
	})
}

func TestCommit_Parent(t *testing.T) {
	c, err := testrepo.CatFileCommit("435ffceb7ba576c937e922766e37d4f7abdcc122")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("ParentsCount", func(t *testing.T) {
		assert.Equal(t, 2, c.ParentsCount())
	})

	t.Run("Parent", func(t *testing.T) {
		t.Run("no such parent", func(t *testing.T) {
			_, err := c.Parent(c.ParentsCount() + 1)
			assert.NotNil(t, err)
			assert.Equal(t, `revision does not exist [rev: , path: ]`, err.Error())
		})

		tests := []struct {
			n           int
			expParentID string
		}{
			{
				n:           0,
				expParentID: "a13dba1e469944772490909daa58c53ac8fa4b0d",
			},
			{
				n:           1,
				expParentID: "7c5ee6478d137417ae602140c615e33aed91887c",
			},
		}
		for _, test := range tests {
			t.Run("", func(t *testing.T) {
				p, err := c.Parent(test.n)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, test.expParentID, p.ID().String())
			})
		}
	})
}

func TestCommit_CommitByPath(t *testing.T) {
	c, err := testrepo.CatFileCommit("435ffceb7ba576c937e922766e37d4f7abdcc122")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path        string
		expCommitID string
	}{
		{
			path:        "", // No path gets back to the commit itself
			expCommitID: "435ffceb7ba576c937e922766e37d4f7abdcc122",
		},
		{
			path:        "resources/labels.properties",
			expCommitID: "755fd577edcfd9209d0ac072eed3b022cbe4d39b",
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			cc, err := c.CommitByPath(CommitByRevisionOptions{
				Path: test.path,
			})
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expCommitID, cc.ID().String())
		})
	}
}

func TestCommit_CommitsByPage(t *testing.T) {
	// There are at most 5 commits can be used for pagination before this commit.
	c, err := testrepo.CatFileCommit("f5ed01959cffa4758ca0a49bf4c34b138d7eab0a")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		page         int
		size         int
		path         string
		expCommitIDs []string
	}{
		{
			page: 0,
			size: 2,
			path: "",
			expCommitIDs: []string{
				"f5ed01959cffa4758ca0a49bf4c34b138d7eab0a",
				"9cdb160ee4118035bf73c744e3bf72a1ba16484a",
			},
		},
		{
			page: 1,
			size: 2,
			path: "",
			expCommitIDs: []string{
				"f5ed01959cffa4758ca0a49bf4c34b138d7eab0a",
				"9cdb160ee4118035bf73c744e3bf72a1ba16484a",
			},
		},
		{
			page: 2,
			size: 2,
			path: "",
			expCommitIDs: []string{
				"dc64fe4ab8618a5be491a9fca46f1585585ea44e",
				"32c273781bab599b955ce7c59d92c39bedf35db0",
			},
		},
		{
			page: 3,
			size: 2,
			path: "",
			expCommitIDs: []string{
				"755fd577edcfd9209d0ac072eed3b022cbe4d39b",
			},
		},
		{
			page:         4,
			size:         2,
			path:         "",
			expCommitIDs: []string{},
		},

		{
			page: 2,
			size: 2,
			path: "src",
			expCommitIDs: []string{
				"755fd577edcfd9209d0ac072eed3b022cbe4d39b",
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			commits, err := c.CommitsByPage(test.page, test.size, CommitsByPageOptions{
				Path: test.path,
			})
			if err != nil {
				t.Fatal(err)
			}

			ids := make([]string, len(commits))
			for i := range commits {
				ids[i] = commits[i].ID().String()
			}

			assert.Equal(t, test.expCommitIDs, ids)
		})
	}
}
