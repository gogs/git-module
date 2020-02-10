package git

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHook(t *testing.T) {
	path := filepath.Join(os.TempDir(), strconv.Itoa(int(time.Now().Unix())))
	h := &Hook{
		name:     HookPreReceive,
		path:     path,
		isSample: false,
		content:  "test content",
	}

	assert.Equal(t, HookPreReceive, h.Name())
	assert.Equal(t, path, h.Path())
	assert.False(t, h.IsSample())
	assert.Equal(t, "test content", h.Content())
}

func TestHook_Update(t *testing.T) {
	path := filepath.Join(os.TempDir(), strconv.Itoa(int(time.Now().Unix())))
	defer func() {
		_ = os.Remove(path)
	}()

	h := &Hook{
		name:     HookPreReceive,
		path:     path,
		isSample: false,
	}
	err := h.Update("test content")
	if err != nil {
		t.Fatal(err)
	}

	p, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "test content", string(p))
}
