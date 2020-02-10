package git

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	path := os.TempDir()
	r := &Repository{
		path: path,
	}

	assert.Equal(t, path, r.Path())
}

func TestInit(t *testing.T) {
	tests := []struct {
		opt InitOptions
	}{
		{
			opt: InitOptions{},
		},
		{
			opt: InitOptions{
				Bare: true,
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			// Make sure it does not blow up
			path := tempPath()
			defer func() {
				_ = os.RemoveAll(path)
			}()

			if err := Init(path, test.opt); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestOpen(t *testing.T) {
	_, err := Open(testrepo.Path())
	assert.Nil(t, err)

	_, err = Open(tempPath())
	assert.Equal(t, os.ErrNotExist, err)
}

func TestClone(t *testing.T) {
	tests := []struct {
		opt CloneOptions
	}{
		{
			opt: CloneOptions{},
		},
		{
			opt: CloneOptions{
				Mirror: true,
				Bare:   true,
				Quiet:  true,
			},
		},
		{
			opt: CloneOptions{
				Branch: "develop",
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			// Make sure it does not blow up
			path := tempPath()
			defer func() {
				_ = os.RemoveAll(path)
			}()

			if err := Clone(testrepo.Path(), path, test.opt); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRepository_Fetch(t *testing.T) {
	path := tempPath()
	defer func() {
		_ = os.RemoveAll(path)
	}()

	if err := Clone(testrepo.Path(), path); err != nil {
		t.Fatal(err)
	}

	r, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		opt FetchOptions
	}{
		{
			opt: FetchOptions{},
		},
		{
			opt: FetchOptions{
				Prune: true,
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			// Make sure it does not blow up
			if err := r.Fetch(test.opt); err != nil {
				t.Fatal(err)
			}
		})
	}
}
