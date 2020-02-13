package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepository_LsTree(t *testing.T) {
	// Make sure it does not blow up
	tree, err := testrepo.LsTree("master", LsTreeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, tree)

	// Tree ID for "gogs/" directory
	tree, err = testrepo.LsTree("fcf7087e732bfe3c25328248a9bf8c3ccd85bed4", LsTreeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, tree)
}
