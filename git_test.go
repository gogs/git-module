package git

import (
	"bytes"
	"fmt"
	"testing"

	goversion "github.com/mcuadros/go-version"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestSetOutput(t *testing.T) {
	assert.Nil(t, logOutput)

	var buf bytes.Buffer
	SetOutput(&buf)

	assert.NotNil(t, logOutput)
}

func TestSetPrefix(t *testing.T) {
	old := logPrefix
	new := "[custom] "
	SetPrefix(new)
	defer SetPrefix(old)

	assert.Equal(t, new, logPrefix)
}

func Test_log(t *testing.T) {
	tests := []struct {
		format    string
		args      []interface{}
		expOutput string
	}{
		{
			format:    "",
			expOutput: "[git-module] \n",
		},
		{
			format:    "something",
			expOutput: "[git-module] something\n",
		},
		{
			format:    "val: %v",
			args:      []interface{}{123},
			expOutput: "[git-module] val: 123\n",
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			var buf bytes.Buffer
			SetOutput(&buf)

			log(test.format, test.args...)
			assert.Equal(t, test.expOutput, buf.String())
		})
	}
}

func TestBinVersion(t *testing.T) {
	g := errgroup.Group{}
	for i := 0; i < 30; i++ {
		g.Go(func() error {
			version, err := BinVersion()
			assert.Nil(t, err)

			if !goversion.Compare(version, "1.8.3", ">=") {
				return fmt.Errorf("version: expected >= 1.8.3 but got %q", version)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}
}
