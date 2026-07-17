package git

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommit_Submodule(t *testing.T) {
	ctx := context.Background()

	c, err := testrepo.CatFileCommit(ctx, testrepoMarks[29].String())
	if err != nil {
		t.Fatal(err)
	}

	mod, err := c.Submodule(ctx, "gogs/docs-api")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "gogs/docs-api", mod.Name)
	assert.Equal(t, "https://github.com/gogs/docs-api.git", mod.URL)
	assert.Equal(t, submoduleSHA.String(), mod.Commit)

	_, err = c.Submodule(ctx, "404")
	assert.Equal(t, ErrSubmoduleNotExist, err)
}
