package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStashWorktreeError(t *testing.T) {
	_, err := testrepo.StashList()
	assert.Errorf(t, err, "StashList() should return an error when not run in a work tree")
}

func TestStash(t *testing.T) {
	tmp := t.TempDir()
	path, err := filepath.Abs(repoPath)
	require.NoError(t, err)

	require.NoError(t, Clone("file://"+path, tmp))

	repo, err := Open(tmp)
	require.NoError(t, err)

	err = os.WriteFile(tmp+"/resources/newfile", []byte("hello, world!"), 0o644)
	require.NoError(t, err)

	f, err := os.OpenFile(tmp+"/README.txt", os.O_APPEND|os.O_WRONLY, 0o644)
	require.NoError(t, err)

	_, err = f.WriteString("\n\ngit-module")
	require.NoError(t, err)

	f.Close()
	err = repo.Add(AddOptions{All: true})
	require.NoError(t, err)

	err = repo.StashPush("")
	require.NoError(t, err)

	f, err = os.OpenFile(tmp+"/README.txt", os.O_APPEND|os.O_WRONLY, 0o644)
	require.NoError(t, err)

	_, err = f.WriteString("\n\nstash 1")
	require.NoError(t, err)

	f.Close()
	err = repo.Add(AddOptions{All: true})
	require.NoError(t, err)

	err = repo.StashPush("custom message")
	require.NoError(t, err)

	want := []*Stash{
		{
			Index:   0,
			Message: "On master: custom message",
			Files:   []string{"README.txt"},
		},
		{
			Index:   1,
			Message: "WIP on master: cfc3b29 Add files with same SHA",
			Files:   []string{"README.txt", "resources/newfile"},
		},
	}

	stash, err := repo.StashList(StashListOptions{
		CommandOptions: CommandOptions{
			Envs: []string{"GIT_CONFIG_GLOBAL=/dev/null"},
		},
	})
	require.NoError(t, err)
	require.Equalf(t, want, stash, "StashList() got = %v, want %v", stash, want)

	wantDiff := &Diff{
		totalAdditions: 4,
		totalDeletions: 0,
		isIncomplete:   false,
		Files: []*DiffFile{
			{
				Name:     "README.txt",
				Type:     DiffFileChange,
				Index:    "72e29aca01368bc0aca5d599c31fa8705b11787d",
				OldIndex: "adfd6da3c0a3fb038393144becbf37f14f780087",
				Sections: []*DiffSection{
					{
						Lines: []*DiffLine{
							{
								Type:    DiffLineSection,
								Content: `@@ -13,3 +13,6 @@ As a quick reminder, this came from one of three locations in either SSH, Git, o`,
							},
							{
								Type:      DiffLinePlain,
								Content:   " We can, as an example effort, even modify this README and change it as if it were source code for the purposes of the class.",
								LeftLine:  13,
								RightLine: 13,
							},
							{
								Type:      DiffLinePlain,
								Content:   " ",
								LeftLine:  14,
								RightLine: 14,
							},
							{
								Type:      DiffLinePlain,
								Content:   " This demo also includes an image with changes on a branch for examination of image diff on GitHub.",
								LeftLine:  15,
								RightLine: 15,
							},
							{
								Type:      DiffLineAdd,
								Content:   "+",
								LeftLine:  0,
								RightLine: 16,
							},
							{
								Type:      DiffLineAdd,
								Content:   "+",
								LeftLine:  0,
								RightLine: 17,
							},
							{
								Type:      DiffLineAdd,
								Content:   "+git-module",
								LeftLine:  0,
								RightLine: 18,
							},
						},
						numAdditions: 3,
						numDeletions: 0,
					},
				},
				numAdditions: 3,
				numDeletions: 0,
				oldName:      "README.txt",
				mode:         0o100644,
				oldMode:      0o100644,
				isBinary:     false,
				isSubmodule:  false,
				isIncomplete: false,
			},
			{
				Name:     "resources/newfile",
				Type:     DiffFileAdd,
				Index:    "30f51a3fba5274d53522d0f19748456974647b4f",
				OldIndex: "0000000000000000000000000000000000000000",
				Sections: []*DiffSection{
					{
						Lines: []*DiffLine{
							{
								Type:    DiffLineSection,
								Content: "@@ -0,0 +1 @@",
							},
							{
								Type:      DiffLineAdd,
								Content:   "+hello, world!",
								LeftLine:  0,
								RightLine: 1,
							},
						},
						numAdditions: 1,
						numDeletions: 0,
					},
				},
				numAdditions: 1,
				numDeletions: 0,
				oldName:      "resources/newfile",
				mode:         0o100644,
				oldMode:      0o100644,
				isBinary:     false,
				isSubmodule:  false,
				isIncomplete: false,
			},
		},
	}

	diff, err := repo.StashDiff(want[1].Index, 0, 0, 0, DiffOptions{
		CommandOptions: CommandOptions{
			Envs: []string{"GIT_CONFIG_GLOBAL=/dev/null"},
		},
	})
	require.NoError(t, err)
	require.Equalf(t, wantDiff, diff, "StashDiff() got = %v, want %v", diff, wantDiff)
}
