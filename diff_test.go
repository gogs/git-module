// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffSection_NumLines(t *testing.T) {
	section := &DiffSection{
		Lines: []*DiffLine{
			{
				Type:      DiffLineAdd,
				Content:   "a line",
				LeftLine:  1,
				RightLine: 10,
			},
		},
	}

	assert.Equal(t, 1, section.NumLines())
}

func TestDiffSection_Line(t *testing.T) {
	lineDelete := &DiffLine{
		Type:      DiffLineDelete,
		Content:   `-  <groupId>com.ambientideas</groupId>`,
		LeftLine:  4,
		RightLine: 0,
	}
	lineAdd := &DiffLine{
		Type:      DiffLineAdd,
		Content:   `+  <groupId>com.github</groupId>`,
		LeftLine:  0,
		RightLine: 4,
	}
	section := &DiffSection{
		Lines: []*DiffLine{
			{
				Type:    DiffLineSection,
				Content: "@@ -1,7 +1,7 @@",
			}, {
				Type:      DiffLinePlain,
				Content:   ` <project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"`,
				LeftLine:  1,
				RightLine: 1,
			}, {
				Type:      DiffLinePlain,
				Content:   `   xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/maven-v4_0_0.xsd">`,
				LeftLine:  2,
				RightLine: 2,
			}, {
				Type:      DiffLinePlain,
				Content:   `   <modelVersion>4.0.0</modelVersion>`,
				LeftLine:  3,
				RightLine: 3,
			},
			lineDelete,
			lineAdd,
			{
				Type:      DiffLinePlain,
				Content:   `   <artifactId>egitdemo</artifactId>`,
				LeftLine:  5,
				RightLine: 5,
			}, {
				Type:      DiffLinePlain,
				Content:   `   <packaging>jar</packaging>`,
				LeftLine:  6,
				RightLine: 6,
			}, {
				Type:      DiffLinePlain,
				Content:   `   <version>1.0-SNAPSHOT</version>`,
				LeftLine:  7,
				RightLine: 7,
			},
		},
	}

	assert.Equal(t, lineDelete, section.Line(lineDelete.Type, 4))
	assert.Equal(t, lineAdd, section.Line(lineAdd.Type, 4))
}

func TestDiffFile(t *testing.T) {
	file := &DiffFile{
		Name:  ".gitmodules",
		Type:  DiffFileAdd,
		Index: "6abde17",
		Sections: []*DiffSection{
			{
				Lines: []*DiffLine{
					{
						Type:    DiffLineSection,
						Content: "@@ -0,0 +1,3 @@",
					}, {
						Type:      DiffLineAdd,
						Content:   `+[submodule "gogs/docs-api"]`,
						LeftLine:  0,
						RightLine: 1,
					}, {
						Type: DiffLineAdd,
						Content: `+	path = gogs/docs-api`,
						LeftLine:  0,
						RightLine: 2,
					},
				},
			},
		},
		numAdditions: 2,
		numDeletions: 0,
		oldName:      "",
		isBinary:     false,
		isSubmodule:  false,
		isIncomplete: true,
	}

	assert.Equal(t, 1, file.NumSections())
	assert.Equal(t, 2, file.NumAdditions())
	assert.Equal(t, 0, file.NumDeletions())
	assert.True(t, file.IsCreated())
	assert.False(t, file.IsDeleted())
	assert.False(t, file.IsRenamed())
	assert.Empty(t, file.OldName())
	assert.False(t, file.IsBinary())
	assert.False(t, file.IsSubmodule())
	assert.True(t, file.IsIncomplete())
}

func TestDiff(t *testing.T) {
	diff := &Diff{
		Files: []*DiffFile{
			{
				Name:         "run.sh",
				Type:         DiffFileRename,
				Index:        "",
				Sections:     nil,
				numAdditions: 0,
				numDeletions: 0,
				oldName:      "runme.sh",
				isBinary:     false,
				isSubmodule:  false,
				isIncomplete: false,
			},
		},
		totalAdditions: 10,
		totalDeletions: 20,
		isIncomplete:   false,
	}

	assert.Equal(t, 1, diff.NumFiles())
	assert.Equal(t, 10, diff.TotalAdditions())
	assert.Equal(t, 20, diff.TotalDeletions())
	assert.False(t, diff.IsIncomplete())
}

func TestStreamParseDiff(t *testing.T) {
	tests := []struct {
		input        string
		maxFileLines int
		maxLineChars int
		maxFiles     int
		expDiff      *Diff
	}{
		{
			input: `diff --git a/.gitmodules b/.gitmodules
new file mode 100644
index 0000000..6abde17
--- /dev/null
+++ b/.gitmodules
@@ -0,0 +1,3 @@
+[submodule "gogs/docs-api"]
+	path = gogs/docs-api
+	url = https://github.com/gogs/docs-api.git
diff --git a/gogs/docs-api b/gogs/docs-api
new file mode 160000
index 0000000..6b08f76
--- /dev/null
+++ b/gogs/docs-api
@@ -0,0 +1 @@
+Subproject commit 6b08f76a5313fa3d26859515b30aa17a5faa2807`,
			expDiff: &Diff{
				Files: []*DiffFile{
					{
						Name:  ".gitmodules",
						Type:  DiffFileAdd,
						Index: "6abde17",
						Sections: []*DiffSection{
							{
								Lines: []*DiffLine{
									{
										Type:    DiffLineSection,
										Content: "@@ -0,0 +1,3 @@",
									}, {
										Type:      DiffLineAdd,
										Content:   `+[submodule "gogs/docs-api"]`,
										LeftLine:  0,
										RightLine: 1,
									}, {
										Type: DiffLineAdd,
										Content: `+	path = gogs/docs-api`,
										LeftLine:  0,
										RightLine: 2,
									}, {
										Type: DiffLineAdd,
										Content: `+	url = https://github.com/gogs/docs-api.git`,
										LeftLine:  0,
										RightLine: 3,
									},
								},
								numAdditions: 3,
								numDeletions: 0,
							},
						},
						numAdditions: 3,
						numDeletions: 0,
						oldName:      "",
						isBinary:     false,
						isSubmodule:  false,
						isIncomplete: false,
					},
					{
						Name:  "gogs/docs-api",
						Type:  DiffFileAdd,
						Index: "6b08f76",
						Sections: []*DiffSection{
							{
								Lines: []*DiffLine{
									{
										Type:    DiffLineSection,
										Content: "@@ -0,0 +1 @@",
									}, {
										Type:      DiffLineAdd,
										Content:   `+Subproject commit 6b08f76a5313fa3d26859515b30aa17a5faa2807`,
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
						oldName:      "",
						isBinary:     false,
						isSubmodule:  true,
						isIncomplete: false,
					},
				},
				totalAdditions: 4,
				totalDeletions: 0,
				isIncomplete:   false,
			},
		},
		{
			input: `diff --git a/pom.xml b/pom.xml
index ee791be..9997571 100644
--- a/pom.xml
+++ b/pom.xml
@@ -1,7 +1,7 @@
 <project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
   xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/maven-v4_0_0.xsd">
   <modelVersion>4.0.0</modelVersion>
-  <groupId>com.ambientideas</groupId>
+  <groupId>com.github</groupId>
   <artifactId>egitdemo</artifactId>
   <packaging>jar</packaging>
   <version>1.0-SNAPSHOT</version>`,
			expDiff: &Diff{
				Files: []*DiffFile{
					{
						Name:  "pom.xml",
						Type:  DiffFileChange,
						Index: "9997571",
						Sections: []*DiffSection{
							{
								Lines: []*DiffLine{
									{
										Type:    DiffLineSection,
										Content: "@@ -1,7 +1,7 @@",
									}, {
										Type:      DiffLinePlain,
										Content:   ` <project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"`,
										LeftLine:  1,
										RightLine: 1,
									}, {
										Type:      DiffLinePlain,
										Content:   `   xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/maven-v4_0_0.xsd">`,
										LeftLine:  2,
										RightLine: 2,
									}, {
										Type:      DiffLinePlain,
										Content:   `   <modelVersion>4.0.0</modelVersion>`,
										LeftLine:  3,
										RightLine: 3,
									}, {
										Type:      DiffLineDelete,
										Content:   `-  <groupId>com.ambientideas</groupId>`,
										LeftLine:  4,
										RightLine: 0,
									}, {
										Type:      DiffLineAdd,
										Content:   `+  <groupId>com.github</groupId>`,
										LeftLine:  0,
										RightLine: 4,
									}, {
										Type:      DiffLinePlain,
										Content:   `   <artifactId>egitdemo</artifactId>`,
										LeftLine:  5,
										RightLine: 5,
									}, {
										Type:      DiffLinePlain,
										Content:   `   <packaging>jar</packaging>`,
										LeftLine:  6,
										RightLine: 6,
									}, {
										Type:      DiffLinePlain,
										Content:   `   <version>1.0-SNAPSHOT</version>`,
										LeftLine:  7,
										RightLine: 7,
									},
								},
								numAdditions: 1,
								numDeletions: 1,
							},
						},
						numAdditions: 1,
						numDeletions: 1,
						oldName:      "",
						isBinary:     false,
						isSubmodule:  false,
						isIncomplete: false,
					},
				},
				totalAdditions: 1,
				totalDeletions: 1,
				isIncomplete:   false,
			},
		},
		{
			input: `diff --git a/img/sourcegraph.png b/img/sourcegraph.png
new file mode 100644
index 0000000..2ce9188
Binary files /dev/null and b/img/sourcegraph.png differ`,
			expDiff: &Diff{
				Files: []*DiffFile{
					{
						Name:         "img/sourcegraph.png",
						Type:         DiffFileAdd,
						Index:        "2ce9188",
						Sections:     nil,
						numAdditions: 0,
						numDeletions: 0,
						oldName:      "",
						isBinary:     true,
						isSubmodule:  false,
						isIncomplete: false,
					},
				},
				totalAdditions: 0,
				totalDeletions: 0,
				isIncomplete:   false,
			},
		},
		{
			input: `diff --git a/fix.txt b/fix.txt
deleted file mode 100644
index e69de29..0000000`,
			expDiff: &Diff{
				Files: []*DiffFile{
					{
						Name:         "fix.txt",
						Type:         DiffFileDelete,
						Index:        "e69de29",
						Sections:     nil,
						numAdditions: 0,
						numDeletions: 0,
						oldName:      "",
						isBinary:     false,
						isSubmodule:  false,
						isIncomplete: false,
					},
				},
				totalAdditions: 0,
				totalDeletions: 0,
				isIncomplete:   false,
			},
		},
		{
			input: `diff --git a/runme.sh b/run.sh
similarity index 100%
rename from runme.sh
rename to run.sh`,
			expDiff: &Diff{
				Files: []*DiffFile{
					{
						Name:         "run.sh",
						Type:         DiffFileRename,
						Index:        "",
						Sections:     nil,
						numAdditions: 0,
						numDeletions: 0,
						oldName:      "runme.sh",
						isBinary:     false,
						isSubmodule:  false,
						isIncomplete: false,
					},
				},
				totalAdditions: 0,
				totalDeletions: 0,
				isIncomplete:   false,
			},
		},
		{
			input: `
diff --git a/dir/file.txt b/dir/file.txt
index b6fc4c620b67d95f953a5c1c1230aaab5db5a1b0..ab80bda5dd90d8b42be25ac2c7a071b722171f09 100644
--- a/dir/file.txt
+++ b/dir/file.txt
@@ -1 +1,3 @@
-hello
\ No newline at end of file
+hello
+
+fdsfdsfds
\ No newline at end of file`,
			expDiff: &Diff{
				Files: []*DiffFile{
					{
						Name:  "dir/file.txt",
						Type:  DiffFileChange,
						Index: "ab80bda5dd90d8b42be25ac2c7a071b722171f09",
						Sections: []*DiffSection{
							{
								Lines: []*DiffLine{
									{
										Type:    DiffLineSection,
										Content: "@@ -1 +1,3 @@",
									}, {
										Type:      DiffLineDelete,
										Content:   `-hello`,
										LeftLine:  1,
										RightLine: 0,
									}, {
										Type:      DiffLineAdd,
										Content:   `+hello`,
										LeftLine:  0,
										RightLine: 1,
									}, {
										Type:      DiffLineAdd,
										Content:   `+`,
										LeftLine:  0,
										RightLine: 2,
									}, {
										Type:      DiffLineAdd,
										Content:   `+fdsfdsfds`,
										LeftLine:  0,
										RightLine: 3,
									},
								},
								numAdditions: 3,
								numDeletions: 1,
							},
						},
						numAdditions: 3,
						numDeletions: 1,
						oldName:      "",
						isBinary:     false,
						isSubmodule:  false,
						isIncomplete: false,
					},
				},
				totalAdditions: 3,
				totalDeletions: 1,
				isIncomplete:   false,
			},
		},
		{
			input: `diff --git a/.travis.yml b/.travis.yml
index 335db7ea..51d7543e 100644
--- a/.travis.yml
+++ b/.travis.yml
@@ -1,9 +1,6 @@
 sudo: false
 language: go
 go:
-  - 1.4.x
-  - 1.5.x
-  - 1.6.x
   - 1.7.x
   - 1.8.x
   - 1.9.x
@@ -12,6 +9,7 @@ go:
   - 1.12.x
   - 1.13.x
 
+install: go get -v ./...
 script: 
   - go get golang.org/x/tools/cmd/cover
   - go get github.com/smartystreets/goconvey`,
			maxFileLines: 2,
			expDiff: &Diff{
				Files: []*DiffFile{
					{
						Name:  ".travis.yml",
						Type:  DiffFileChange,
						Index: "51d7543e",
						Sections: []*DiffSection{
							{
								Lines: []*DiffLine{
									{
										Type:    DiffLineSection,
										Content: "@@ -1,9 +1,6 @@",
									}, {
										Type:      DiffLinePlain,
										Content:   ` sudo: false`,
										LeftLine:  1,
										RightLine: 1,
									}, {
										Type:      DiffLinePlain,
										Content:   ` language: go`,
										LeftLine:  2,
										RightLine: 2,
									}, {
										Type:      DiffLinePlain,
										Content:   ` go:`,
										LeftLine:  3,
										RightLine: 3,
									}, {
										Type:      DiffLineDelete,
										Content:   `-  - 1.4.x`,
										LeftLine:  4,
										RightLine: 0,
									}, {
										Type:      DiffLineDelete,
										Content:   `-  - 1.5.x`,
										LeftLine:  5,
										RightLine: 0,
									}, {
										Type:      DiffLineDelete,
										Content:   `-  - 1.6.x`,
										LeftLine:  6,
										RightLine: 0,
									}, {
										Type:      DiffLinePlain,
										Content:   `   - 1.7.x`,
										LeftLine:  7,
										RightLine: 4,
									}, {
										Type:      DiffLinePlain,
										Content:   `   - 1.8.x`,
										LeftLine:  8,
										RightLine: 5,
									}, {
										Type:      DiffLinePlain,
										Content:   `   - 1.9.x`,
										LeftLine:  9,
										RightLine: 6,
									},
								},
								numAdditions: 0,
								numDeletions: 3,
							},
						},
						numAdditions: 0,
						numDeletions: 3,
						oldName:      "",
						isBinary:     false,
						isSubmodule:  false,
						isIncomplete: true,
					},
				},
				totalAdditions: 0,
				totalDeletions: 3,
				isIncomplete:   true,
			},
		},
		{
			input: `diff --git a/.gitmodules b/.gitmodules
new file mode 100644
index 0000000..6abde17
--- /dev/null
+++ b/.gitmodules
@@ -0,0 +1,3 @@
+[submodule "gogs/docs-api"]
+	path = gogs/docs-api
+	url = https://github.com/gogs/docs-api.git`,
			maxLineChars: 30,
			expDiff: &Diff{
				Files: []*DiffFile{
					{
						Name:  ".gitmodules",
						Type:  DiffFileAdd,
						Index: "6abde17",
						Sections: []*DiffSection{
							{
								Lines: []*DiffLine{
									{
										Type:    DiffLineSection,
										Content: "@@ -0,0 +1,3 @@",
									}, {
										Type:      DiffLineAdd,
										Content:   `+[submodule "gogs/docs-api"]`,
										LeftLine:  0,
										RightLine: 1,
									}, {
										Type: DiffLineAdd,
										Content: `+	path = gogs/docs-api`,
										LeftLine:  0,
										RightLine: 2,
									},
								},
								numAdditions: 2,
								numDeletions: 0,
							},
						},
						numAdditions: 2,
						numDeletions: 0,
						oldName:      "",
						isBinary:     false,
						isSubmodule:  false,
						isIncomplete: true,
					},
				},
				totalAdditions: 2,
				totalDeletions: 0,
				isIncomplete:   true,
			},
		},
		{
			input: `diff --git a/.gitmodules b/.gitmodules
new file mode 100644
index 0000000..6abde17
--- /dev/null
+++ b/.gitmodules
@@ -0,0 +1,3 @@
+[submodule "gogs/docs-api"]
+	path = gogs/docs-api
+	url = https://github.com/gogs/docs-api.git
diff --git a/gogs/docs-api b/gogs/docs-api
new file mode 160000
index 0000000..6b08f76
--- /dev/null
+++ b/gogs/docs-api
@@ -0,0 +1 @@
+Subproject commit 6b08f76a5313fa3d26859515b30aa17a5faa2807`,
			maxFileLines: 2,
			maxLineChars: 30,
			maxFiles:     1,
			expDiff: &Diff{
				Files: []*DiffFile{
					{
						Name:  ".gitmodules",
						Type:  DiffFileAdd,
						Index: "6abde17",
						Sections: []*DiffSection{
							{
								Lines: []*DiffLine{
									{
										Type:    DiffLineSection,
										Content: "@@ -0,0 +1,3 @@",
									}, {
										Type:      DiffLineAdd,
										Content:   `+[submodule "gogs/docs-api"]`,
										LeftLine:  0,
										RightLine: 1,
									}, {
										Type: DiffLineAdd,
										Content: `+	path = gogs/docs-api`,
										LeftLine:  0,
										RightLine: 2,
									},
								},
								numAdditions: 2,
								numDeletions: 0,
							},
						},
						numAdditions: 2,
						numDeletions: 0,
						oldName:      "",
						isBinary:     false,
						isSubmodule:  false,
						isIncomplete: true,
					},
				},
				totalAdditions: 2,
				totalDeletions: 0,
				isIncomplete:   true,
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			done := make(chan SteamParseDiffResult)
			go StreamParseDiff(strings.NewReader(test.input), done, test.maxFiles, test.maxFileLines, test.maxLineChars)
			result := <-done
			if result.Err != nil {
				t.Fatal(result.Err)
			}

			assert.Equal(t, test.expDiff, result.Diff)
		})
	}
}
