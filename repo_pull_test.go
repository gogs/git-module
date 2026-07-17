package git

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_MergeBase(t *testing.T) {
	ctx := context.Background()

	t.Run("bad revision", func(t *testing.T) {
		// "bad_revision" doesn't exist, so git fails with exit status 128 (fatal),
		// not exit status 1 (no merge base).
		mb, err := testrepo.MergeBase(ctx, testrepoMarks[30].String(), "bad_revision")
		assert.Error(t, err)
		assert.Empty(t, mb)
	})

	tests := []struct {
		base          string
		head          string
		opt           MergeBaseOptions
		wantMergeBase string
	}{
		{
			base:          testrepoMarks[29].String(),
			head:          testrepoMarks[30].String(),
			wantMergeBase: testrepoMarks[29].String(),
		},
		{
			base:          "master",
			head:          "release-1.0",
			wantMergeBase: testrepoMarks[30].String(),
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			mb, err := testrepo.MergeBase(ctx, test.base, test.head, test.opt)
			require.NoError(t, err)
			assert.Equal(t, test.wantMergeBase, mb)
		})
	}
}
