package git

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTreeEntry(t *testing.T) {
	id := MustIDFromString(testrepoMarks[12].String())
	e := &TreeEntry{
		mode: EntrySymlink,
		typ:  ObjectTree,
		id:   id,
		name: "go.mod",
	}

	assert.False(t, e.IsTree())
	assert.False(t, e.IsBlob())
	assert.False(t, e.IsExec())
	assert.True(t, e.IsSymlink())
	assert.False(t, e.IsCommit())

	assert.Equal(t, ObjectTree, e.Type())
	assert.Equal(t, e.id, e.ID())
	assert.Equal(t, "go.mod", e.Name())
}

func TestTreeEntry_Size(t *testing.T) {
	ctx := context.Background()
	tree, err := testrepo.LsTree(ctx, testrepoMarks[30].String())
	require.NoError(t, err)

	es, err := tree.Entries(ctx)
	require.NoError(t, err)

	t.Run("blob", func(t *testing.T) {
		var entry *TreeEntry
		for _, e := range es {
			if e.Name() == "README.txt" {
				entry = e
				break
			}
		}
		require.NotNil(t, entry, "entry README.txt not found")
		assert.Equal(t, int64(795), entry.Size(ctx))
	})

	t.Run("tree returns zero", func(t *testing.T) {
		var entry *TreeEntry
		for _, e := range es {
			if e.IsTree() {
				entry = e
				break
			}
		}
		require.NotNil(t, entry, "tree entry not found")
		assert.Equal(t, int64(0), entry.Size(ctx))
	})
}

func TestEntries_Sort(t *testing.T) {
	ctx := context.Background()
	tree, err := testrepo.LsTree(ctx, testrepoMarks[30].String())
	if err != nil {
		t.Fatal(err)
	}

	es, err := tree.Entries(ctx)
	if err != nil {
		t.Fatal(err)
	}

	es.Sort()

	expEntries := []*TreeEntry{
		{
			mode: EntryTree,
			typ:  ObjectTree,
			id:   MustIDFromString(testrepoMarks[41].String()),
			name: "gogs",
		}, {
			mode: EntryTree,
			typ:  ObjectTree,
			id:   MustIDFromString(testrepoMarks[42].String()),
			name: "img",
		}, {
			mode: EntryTree,
			typ:  ObjectTree,
			id:   MustIDFromString(testrepoMarks[44].String()),
			name: "resources",
		}, {
			mode: EntryTree,
			typ:  ObjectTree,
			id:   MustIDFromString(testrepoMarks[47].String()),
			name: "src",
		}, {
			mode: EntryBlob,
			typ:  ObjectBlob,
			id:   MustIDFromString(testrepoMarks[34].String()),
			name: ".DS_Store",
		}, {
			mode: EntryBlob,
			typ:  ObjectBlob,
			id:   MustIDFromString(testrepoMarks[35].String()),
			name: ".gitattributes",
		}, {
			mode: EntryBlob,
			typ:  ObjectBlob,
			id:   MustIDFromString(testrepoMarks[36].String()),
			name: ".gitignore",
		}, {
			mode: EntryBlob,
			typ:  ObjectBlob,
			id:   MustIDFromString(testrepoMarks[37].String()),
			name: ".gitmodules",
		}, {
			mode: EntryBlob,
			typ:  ObjectBlob,
			id:   MustIDFromString(testrepoMarks[38].String()),
			name: ".travis.yml",
		}, {
			mode: EntryBlob,
			typ:  ObjectBlob,
			id:   MustIDFromString(testrepoMarks[39].String()),
			name: "README.txt",
		}, {
			mode: EntryBlob,
			typ:  ObjectBlob,
			id:   MustIDFromString(testrepoMarks[40].String()),
			name: "build.gradle",
		}, {
			mode: EntryBlob,
			typ:  ObjectBlob,
			id:   MustIDFromString(testrepoMarks[43].String()),
			name: "pom.xml",
		}, {
			mode: EntryExec,
			typ:  ObjectBlob,
			id:   MustIDFromString(testrepoMarks[45].String()),
			name: "run.sh",
		},
	}
	for i := range expEntries {
		assert.Equal(t, expEntries[i].Mode(), es[i].Mode(), "idx: %d", i)
		assert.Equal(t, expEntries[i].Type(), es[i].Type(), "idx: %d", i)
		assert.Equal(t, expEntries[i].ID().String(), es[i].ID().String(), "idx: %d", i)
		assert.Equal(t, expEntries[i].Name(), es[i].Name(), "idx: %d", i)
	}
}

func TestEntries_CommitsInfo(t *testing.T) {
	ctx := context.Background()
	tree, err := testrepo.LsTree(ctx, testrepoMarks[32].String())
	if err != nil {
		t.Fatal(err)
	}

	c, err := testrepo.CatFileCommit(ctx, tree.id.String())
	if err != nil {
		t.Fatal(err)
	}

	t.Run("general directory", func(t *testing.T) {
		es, err := tree.Entries(ctx)
		if err != nil {
			t.Fatal(err)
		}

		infos, err := es.CommitsInfo(ctx, c)
		if err != nil {
			t.Fatal(err)
		}

		expInfos := []*EntryCommitInfo{
			{
				Entry: &TreeEntry{
					name: ".DS_Store",
				},
				Commit: &Commit{
					ID: MustIDFromString(testrepoMarks[28].String()),
				},
			}, {
				Entry: &TreeEntry{
					name: ".gitattributes",
				},
				Commit: &Commit{
					ID: MustIDFromString(testrepoMarks[15].String()),
				},
			}, {
				Entry: &TreeEntry{
					name: ".gitignore",
				},
				Commit: &Commit{
					ID: MustIDFromString(testrepoMarks[16].String()),
				},
			}, {
				Entry: &TreeEntry{
					name: ".gitmodules",
				},
				Commit: &Commit{
					ID: MustIDFromString(testrepoMarks[29].String()),
				},
			}, {
				Entry: &TreeEntry{
					name: ".travis.yml",
				},
				Commit: &Commit{
					ID: MustIDFromString(testrepoMarks[23].String()),
				},
			}, {
				Entry: &TreeEntry{
					name: "README.txt",
				},
				Commit: &Commit{
					ID: MustIDFromString(testrepoMarks[20].String()),
				},
			}, {
				Entry: &TreeEntry{
					name: "build.gradle",
				},
				Commit: &Commit{
					ID: MustIDFromString(testrepoMarks[17].String()),
				},
			}, {
				Entry: &TreeEntry{
					name: "gogs",
				},
				Commit: &Commit{
					ID: MustIDFromString(testrepoMarks[29].String()),
				},
			}, {
				Entry: &TreeEntry{
					name: "img",
				},
				Commit: &Commit{
					ID: MustIDFromString(testrepoMarks[28].String()),
				},
			}, {
				Entry: &TreeEntry{
					name: "pom.xml",
				},
				Commit: &Commit{
					ID: MustIDFromString(testrepoMarks[26].String()),
				},
			}, {
				Entry: &TreeEntry{
					name: "resources",
				},
				Commit: &Commit{
					ID: MustIDFromString(testrepoMarks[1].String()),
				},
			}, {
				Entry: &TreeEntry{
					name: "run.sh",
				},
				Commit: &Commit{
					ID: MustIDFromString(testrepoMarks[30].String()),
				},
			}, {
				Entry: &TreeEntry{
					name: "sameSHAs",
				},
				Commit: &Commit{
					ID: MustIDFromString(testrepoMarks[32].String()),
				},
			}, {
				Entry: &TreeEntry{
					name: "src",
				},
				Commit: &Commit{
					ID: MustIDFromString(testrepoMarks[25].String()),
				},
			},
		}
		for i := range expInfos {
			assert.Equal(t, expInfos[i].Entry.Name(), infos[i].Entry.Name(), "idx: %d", i)
			assert.Equal(t, expInfos[i].Commit.ID.String(), infos[i].Commit.ID.String(), "idx: %d", i)
		}
	})

	t.Run("directory with submodule", func(t *testing.T) {
		subtree, err := tree.Subtree(ctx, "gogs")
		if err != nil {
			t.Fatal(err)
		}

		es, err := subtree.Entries(ctx)
		if err != nil {
			t.Fatal(err)
		}

		infos, err := es.CommitsInfo(ctx, c, CommitsInfoOptions{
			Path: "gogs",
		})
		if err != nil {
			t.Fatal(err)
		}

		expInfos := []*EntryCommitInfo{
			{
				Entry: &TreeEntry{
					name: "docs-api",
				},
				Commit: &Commit{
					ID: MustIDFromString(testrepoMarks[29].String()),
				},
			},
		}
		for i := range expInfos {
			assert.Equal(t, expInfos[i].Entry.Name(), infos[i].Entry.Name(), "idx: %d", i)
			assert.Equal(t, expInfos[i].Commit.ID.String(), infos[i].Commit.ID.String(), "idx: %d", i)
		}
	})

	t.Run("directory with files that have the same SHA", func(t *testing.T) {
		subtree, err := tree.Subtree(ctx, "sameSHAs")
		if err != nil {
			t.Fatal(err)
		}

		es, err := subtree.Entries(ctx)
		if err != nil {
			t.Fatal(err)
		}

		infos, err := es.CommitsInfo(ctx, c, CommitsInfoOptions{
			Path: "sameSHAs",
		})
		if err != nil {
			t.Fatal(err)
		}

		expInfos := []*EntryCommitInfo{
			{
				Entry: &TreeEntry{
					name: "file1.txt",
				},
				Commit: &Commit{
					ID: MustIDFromString(testrepoMarks[32].String()),
				},
			}, {
				Entry: &TreeEntry{
					name: "file2.txt",
				},
				Commit: &Commit{
					ID: MustIDFromString(testrepoMarks[32].String()),
				},
			},
		}
		for i := range expInfos {
			assert.Equal(t, expInfos[i].Entry.Name(), infos[i].Entry.Name(), "idx: %d", i)
			assert.Equal(t, expInfos[i].Commit.ID.String(), infos[i].Commit.ID.String(), "idx: %d", i)
		}
	})
}
