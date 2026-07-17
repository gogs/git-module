package git

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_CatFileBlob(t *testing.T) {
	ctx := context.Background()

	t.Run("not a blob", func(t *testing.T) {
		_, err := testrepo.CatFileBlob(ctx, testrepoMarks[47].String())
		assert.Equal(t, ErrNotBlob, err)
	})

	t.Run("get a blob, no full rev hash", func(t *testing.T) {
		b, err := testrepo.CatFileBlob(ctx, testrepoMarks[34].String()[:4])
		require.NoError(t, err)
		assert.True(t, b.IsBlob())
	})

	t.Run("get a blob", func(t *testing.T) {
		b, err := testrepo.CatFileBlob(ctx, testrepoMarks[34].String())
		require.NoError(t, err)
		assert.True(t, b.IsBlob())
	})
}
