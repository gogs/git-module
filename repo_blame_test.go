package git

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepository_Blame(t *testing.T) {
	ctx := context.Background()

	t.Run("bad file", func(t *testing.T) {
		_, err := testrepo.Blame(ctx, "", "404.txt")
		assert.Error(t, err)
	})

	blame, err := testrepo.Blame(ctx, testrepoMarks[32].String(), "README.txt")
	assert.Nil(t, err)

	// Assert representative commits
	// https://github.com/gogs/git-module-testrepo/blame/master/README.txt
	tests := []struct {
		line  int
		expID string
	}{
		{line: 1, expID: testrepoMarks[1].String()},
		{line: 3, expID: testrepoMarks[20].String()},
		{line: 5, expID: testrepoMarks[1].String()},
		{line: 13, expID: testrepoMarks[13].String()},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("Line %d", test.line), func(t *testing.T) {
			line := blame.Line(test.line)
			assert.Equal(t, test.expID, line.ID.String())
		})
	}
}
