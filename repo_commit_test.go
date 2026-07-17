package git

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_escapePath(t *testing.T) {
	tests := []struct {
		path    string
		expPath string
	}{
		{
			path:    "",
			expPath: "",
		},
		{
			path:    "normal",
			expPath: "normal",
		},
		{
			path:    ":normal",
			expPath: "\\:normal",
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, test.expPath, escapePath(test.path))
		})
	}
}

func TestRepository_CatFileCommit(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid revision", func(t *testing.T) {
		c, err := testrepo.CatFileCommit(ctx, "bad_revision")
		assert.Equal(t, ErrRevisionNotExist, err)
		assert.Nil(t, c)
	})

	c, err := testrepo.CatFileCommit(ctx, testrepoMarks[31].String())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, testrepoMarks[31].String(), c.ID.String())
	assert.Equal(t, "Add a symlink\n", c.Message)
}

func TestRepository_BranchCommit(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid branch", func(t *testing.T) {
		c, err := testrepo.BranchCommit(ctx, "refs/heads/release-1.0")
		assert.Equal(t, ErrRevisionNotExist, err)
		assert.Nil(t, c)
	})

	c, err := testrepo.BranchCommit(ctx, "release-1.0")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, testrepoMarks[30].String(), c.ID.String())
	assert.Equal(t, "Rename shell script\n", c.Message)
}

func TestRepository_TagCommit(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid branch", func(t *testing.T) {
		c, err := testrepo.BranchCommit(ctx, "refs/tags/v1.0.0")
		assert.Equal(t, ErrRevisionNotExist, err)
		assert.Nil(t, c)
	})

	c, err := testrepo.BranchCommit(ctx, "release-1.0")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, testrepoMarks[30].String(), c.ID.String())
	assert.Equal(t, "Rename shell script\n", c.Message)
}

func TestRepository_Log(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		rev          string
		opt          LogOptions
		expCommitIDs []string
	}{
		{
			rev: testrepoMarks[30].String(),
			opt: LogOptions{
				Since: time.Unix(1581250680, 0),
			},
			expCommitIDs: []string{
				testrepoMarks[30].String(),
				testrepoMarks[29].String(),
			},
		},
		{
			rev: testrepoMarks[30].String(),
			opt: LogOptions{
				Since: time.Now().AddDate(100, 0, 0),
			},
			expCommitIDs: []string{},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			commits, err := testrepo.Log(ctx, test.rev, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expCommitIDs, commitsToIDs(commits))
		})
	}
}

func TestRepository_CommitByRevision(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid revision", func(t *testing.T) {
		c, err := testrepo.CommitByRevision(ctx, "bad_revision")
		assert.Equal(t, ErrRevisionNotExist, err)
		assert.Nil(t, c)
	})

	tests := []struct {
		rev   string
		opt   CommitByRevisionOptions
		expID string
	}{
		{
			rev:   testrepoMarks[29].String()[:7],
			expID: testrepoMarks[29].String(),
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			c, err := testrepo.CommitByRevision(ctx, test.rev, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expID, c.ID.String())
		})
	}
}

func TestRepository_CommitsSince(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		rev          string
		since        time.Time
		opt          CommitsSinceOptions
		expCommitIDs []string
	}{
		{
			rev:   testrepoMarks[30].String(),
			since: time.Unix(1581250680, 0),
			expCommitIDs: []string{
				testrepoMarks[30].String(),
				testrepoMarks[29].String(),
			},
		},
		{
			rev:          testrepoMarks[30].String(),
			since:        time.Now().AddDate(100, 0, 0),
			expCommitIDs: []string{},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			commits, err := testrepo.CommitsSince(ctx, test.rev, test.since, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expCommitIDs, commitsToIDs(commits))
		})
	}
}

func TestRepository_DiffNameOnly(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		base     string
		head     string
		opt      DiffNameOnlyOptions
		expFiles []string
	}{
		{
			base:     testrepoMarks[26].String(),
			head:     testrepoMarks[27].String(),
			expFiles: []string{"fix.txt"},
		},
		{
			base: testrepoMarks[24].String(),
			head: testrepoMarks[27].String(),
			opt: DiffNameOnlyOptions{
				NeedsMergeBase: true,
			},
			expFiles: []string{"fix.txt", "pom.xml", "src/test/java/com/github/AppTest.java"},
		},

		{
			base: testrepoMarks[24].String(),
			head: testrepoMarks[27].String(),
			opt: DiffNameOnlyOptions{
				Path: "src",
			},
			expFiles: []string{"src/test/java/com/github/AppTest.java"},
		},
		{
			base: testrepoMarks[24].String(),
			head: testrepoMarks[27].String(),
			opt: DiffNameOnlyOptions{
				Path: "resources",
			},
			expFiles: []string{},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			files, err := testrepo.DiffNameOnly(ctx, test.base, test.head, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expFiles, files)
		})
	}
}

func TestRepository_RevListCount(t *testing.T) {
	ctx := context.Background()

	t.Run("no refspecs", func(t *testing.T) {
		count, err := testrepo.RevListCount(ctx, []string{})
		assert.Equal(t, errors.New("must have at least one refspec"), err)
		assert.Zero(t, count)
	})

	tests := []struct {
		refspecs []string
		opt      RevListCountOptions
		expCount int64
	}{
		{
			refspecs: []string{testrepoMarks[1].String()},
			expCount: 1,
		},
		{
			refspecs: []string{testrepoMarks[5].String()},
			expCount: 5,
		},
		{
			refspecs: []string{testrepoMarks[27].String()},
			expCount: 27,
		},

		{
			refspecs: []string{testrepoMarks[21].String()},
			opt: RevListCountOptions{
				Path: "README.txt",
			},
			expCount: 3,
		},
		{
			refspecs: []string{testrepoMarks[21].String()},
			opt: RevListCountOptions{
				Path: "resources",
			},
			expCount: 1,
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			count, err := testrepo.RevListCount(ctx, test.refspecs, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expCount, count)
		})
	}
}

func TestRepository_RevList(t *testing.T) {
	ctx := context.Background()

	t.Run("no refspecs", func(t *testing.T) {
		commits, err := testrepo.RevList(ctx, []string{})
		assert.Equal(t, errors.New("must have at least one refspec"), err)
		assert.Nil(t, commits)
	})

	tests := []struct {
		refspecs     []string
		opt          RevListOptions
		expCommitIDs []string
	}{
		{
			refspecs: []string{fmt.Sprintf("%s...%s", testrepoMarks[24].String(), testrepoMarks[27].String())},
			expCommitIDs: []string{
				testrepoMarks[27].String(),
				testrepoMarks[26].String(),
				testrepoMarks[25].String(),
			},
		},
		{
			refspecs: []string{fmt.Sprintf("%s...%s", testrepoMarks[24].String(), testrepoMarks[27].String())},
			opt: RevListOptions{
				Path: "src",
			},
			expCommitIDs: []string{
				testrepoMarks[25].String(),
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			commits, err := testrepo.RevList(ctx, test.refspecs, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expCommitIDs, commitsToIDs(commits))
		})
	}
}

func TestRepository_LatestCommitTime(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		opt     LatestCommitTimeOptions
		expTime time.Time
	}{
		{
			opt: LatestCommitTimeOptions{
				Branch: "release-1.0",
			},
			expTime: time.Unix(1581256638, 0),
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			got, err := testrepo.LatestCommitTime(ctx, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expTime.Unix(), got.Unix())
		})
	}
}
