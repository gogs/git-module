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
			n        int
			parentID string
		}{
			{
				n:        0,
				parentID: "a13dba1e469944772490909daa58c53ac8fa4b0d",
			},
			{
				n:        1,
				parentID: "7c5ee6478d137417ae602140c615e33aed91887c",
			},
		}
		for _, test := range tests {
			t.Run("", func(t *testing.T) {
				p, err := c.Parent(test.n)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, test.parentID, p.ID().String())
			})
		}
	})
}
