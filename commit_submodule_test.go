package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommit_Submodule(t *testing.T) {
	c, err := testrepo.CatFileCommit("4e59b72440188e7c2578299fc28ea425fbe9aece")
	if err != nil {
		t.Fatal(err)
	}

	mod, err := c.Submodule("gogs/docs-api")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "gogs/docs-api", mod.Name())
	assert.Equal(t, "https://github.com/gogs/docs-api.git", mod.URL())

	_, err = c.Submodule("404")
	assert.Equal(t, ErrSubmoduleNotExist, err)
}
