package git

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffLine(t *testing.T) {
	line := &DiffLine{
		typ:       DiffLineAdd,
		content:   "a line",
		leftLine:  1,
		rightLine: 10,
	}

	assert.Equal(t, DiffLineAdd, line.Type())
	assert.Equal(t, "a line", line.Content())
	assert.Equal(t, 1, line.Left())
	assert.Equal(t, 10, line.Right())
}

func TestDiffSection_Lines(t *testing.T) {
	lines := []*DiffLine{
		{
			typ:       DiffLineAdd,
			content:   "a line",
			leftLine:  1,
			rightLine: 10,
		},
	}
	section := &DiffSection{
		lines: lines,
	}

	assert.Equal(t, lines, section.Lines())
}

func TestDiffSection_Line(t *testing.T) {
	lineDelete := &DiffLine{
		typ:       DiffLineDelete,
		content:   `-  <groupId>com.ambientideas</groupId>`,
		leftLine:  4,
		rightLine: 0,
	}
	lineAdd := &DiffLine{
		typ:       DiffLineAdd,
		content:   `+  <groupId>com.github</groupId>`,
		leftLine:  0,
		rightLine: 4,
	}
	section := &DiffSection{
		lines: []*DiffLine{
			{
				typ:       DiffLineSection,
				content:   "@@ -1,7 +1,7 @@",
				leftLine:  0,
				rightLine: 0,
			},
			{
				typ:       DiffLinePlain,
				content:   ` <project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"`,
				leftLine:  1,
				rightLine: 1,
			},
			{
				typ:       DiffLinePlain,
				content:   `   xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/maven-v4_0_0.xsd">`,
				leftLine:  2,
				rightLine: 2,
			},
			{
				typ:       DiffLinePlain,
				content:   `   <modelVersion>4.0.0</modelVersion>`,
				leftLine:  3,
				rightLine: 3,
			},
			lineDelete,
			lineAdd,
			{
				typ:       DiffLinePlain,
				content:   `   <artifactId>egitdemo</artifactId>`,
				leftLine:  5,
				rightLine: 5,
			},
			{
				typ:       DiffLinePlain,
				content:   `   <packaging>jar</packaging>`,
				leftLine:  6,
				rightLine: 6,
			},
			{
				typ:       DiffLinePlain,
				content:   `   <version>1.0-SNAPSHOT</version>`,
				leftLine:  7,
				rightLine: 7,
			},
		},
	}

	assert.Equal(t, lineDelete, section.Line(lineDelete.Type(), 4))
	assert.Equal(t, lineAdd, section.Line(lineAdd.Type(), 4))
}

func TestDiffFile(t *testing.T) {
	sections := []*DiffSection{
		{
			lines: []*DiffLine{
				{
					typ:       DiffLineSection,
					content:   "@@ -0,0 +1,3 @@",
					leftLine:  0,
					rightLine: 0,
				},
				{
					typ:       DiffLineAdd,
					content:   `+[submodule "gogs/docs-api"]`,
					leftLine:  0,
					rightLine: 1,
				},
				{
					typ: DiffLineAdd,
					content: `+	path = gogs/docs-api`,
					leftLine:  0,
					rightLine: 2,
				},
			},
		},
	}
	file := &DiffFile{
		name:         ".gitmodules",
		typ:          DiffFileAdd,
		index:        "6abde17",
		sections:     sections,
		numAdditions: 2,
		numDeletions: 0,
		oldName:      "",
		isBinary:     false,
		isSubmodule:  false,
		isIncomplete: true,
	}

	assert.Equal(t, ".gitmodules", file.Name())
	assert.Equal(t, DiffFileAdd, file.Type())
	assert.Equal(t, "6abde17", file.Index())
	assert.Equal(t, sections, file.Sections())
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
	files := []*DiffFile{
		{
			name:         "run.sh",
			typ:          DiffFileRename,
			index:        "",
			sections:     nil,
			numAdditions: 0,
			numDeletions: 0,
			oldName:      "runme.sh",
			isBinary:     false,
			isSubmodule:  false,
			isIncomplete: false,
		},
	}
	diff := &Diff{
		files:          files,
		totalAdditions: 10,
		totalDeletions: 20,
		isIncomplete:   false,
	}

	assert.Equal(t, files, diff.Files())
	assert.Equal(t, 1, diff.NumFiles())
	assert.Equal(t, 10, diff.TotalAdditions())
	assert.Equal(t, 20, diff.TotalDeletions())
	assert.False(t, diff.IsIncomplete())
}

func TestSteamParsePatch(t *testing.T) {
	tests := []struct {
		input        string
		maxLines     int
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
				files: []*DiffFile{
					{
						name:  ".gitmodules",
						typ:   DiffFileAdd,
						index: "6abde17",
						sections: []*DiffSection{
							{
								lines: []*DiffLine{
									{
										typ:       DiffLineSection,
										content:   "@@ -0,0 +1,3 @@",
										leftLine:  0,
										rightLine: 0,
									},
									{
										typ:       DiffLineAdd,
										content:   `+[submodule "gogs/docs-api"]`,
										leftLine:  0,
										rightLine: 1,
									},
									{
										typ: DiffLineAdd,
										content: `+	path = gogs/docs-api`,
										leftLine:  0,
										rightLine: 2,
									},
									{
										typ: DiffLineAdd,
										content: `+	url = https://github.com/gogs/docs-api.git`,
										leftLine:  0,
										rightLine: 3,
									},
								},
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
						name:  "gogs/docs-api",
						typ:   DiffFileAdd,
						index: "6b08f76",
						sections: []*DiffSection{
							{
								lines: []*DiffLine{
									{
										typ:       DiffLineSection,
										content:   "@@ -0,0 +1 @@",
										leftLine:  0,
										rightLine: 0,
									},
									{
										typ:       DiffLineAdd,
										content:   `+Subproject commit 6b08f76a5313fa3d26859515b30aa17a5faa2807`,
										leftLine:  0,
										rightLine: 1,
									},
								},
							},
						},
						numAdditions: 1,
						numDeletions: 0,
						oldName:      "",
						isBinary:     false,
						isSubmodule:  false,
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
				files: []*DiffFile{
					{
						name:  "pom.xml",
						typ:   DiffFileChange,
						index: "9997571",
						sections: []*DiffSection{
							{
								lines: []*DiffLine{
									{
										typ:       DiffLineSection,
										content:   "@@ -1,7 +1,7 @@",
										leftLine:  0,
										rightLine: 0,
									},
									{
										typ:       DiffLinePlain,
										content:   ` <project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"`,
										leftLine:  1,
										rightLine: 1,
									},
									{
										typ:       DiffLinePlain,
										content:   `   xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/maven-v4_0_0.xsd">`,
										leftLine:  2,
										rightLine: 2,
									},
									{
										typ:       DiffLinePlain,
										content:   `   <modelVersion>4.0.0</modelVersion>`,
										leftLine:  3,
										rightLine: 3,
									},
									{
										typ:       DiffLineDelete,
										content:   `-  <groupId>com.ambientideas</groupId>`,
										leftLine:  4,
										rightLine: 0,
									},
									{
										typ:       DiffLineAdd,
										content:   `+  <groupId>com.github</groupId>`,
										leftLine:  0,
										rightLine: 4,
									},
									{
										typ:       DiffLinePlain,
										content:   `   <artifactId>egitdemo</artifactId>`,
										leftLine:  5,
										rightLine: 5,
									},
									{
										typ:       DiffLinePlain,
										content:   `   <packaging>jar</packaging>`,
										leftLine:  6,
										rightLine: 6,
									},
									{
										typ:       DiffLinePlain,
										content:   `   <version>1.0-SNAPSHOT</version>`,
										leftLine:  7,
										rightLine: 7,
									},
								},
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
				files: []*DiffFile{
					{
						name:         "img/sourcegraph.png",
						typ:          DiffFileAdd,
						index:        "2ce9188",
						sections:     nil,
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
				files: []*DiffFile{
					{
						name:         "fix.txt",
						typ:          DiffFileDelete,
						index:        "e69de29",
						sections:     nil,
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
				files: []*DiffFile{
					{
						name:         "run.sh",
						typ:          DiffFileRename,
						index:        "",
						sections:     nil,
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
			input: `diff --git a/.gitmodules b/.gitmodules
new file mode 100644
index 0000000..6abde17
--- /dev/null
+++ b/.gitmodules
@@ -0,0 +1,3 @@
+[submodule "gogs/docs-api"]
+	path = gogs/docs-api
+	url = https://github.com/gogs/docs-api.git`,
			maxLines: 2,
			expDiff: &Diff{
				files: []*DiffFile{
					{
						name:  ".gitmodules",
						typ:   DiffFileAdd,
						index: "6abde17",
						sections: []*DiffSection{
							{
								lines: []*DiffLine{
									{
										typ:       DiffLineSection,
										content:   "@@ -0,0 +1,3 @@",
										leftLine:  0,
										rightLine: 0,
									},
									{
										typ:       DiffLineAdd,
										content:   `+[submodule "gogs/docs-api"]`,
										leftLine:  0,
										rightLine: 1,
									},
									{
										typ: DiffLineAdd,
										content: `+	path = gogs/docs-api`,
										leftLine:  0,
										rightLine: 2,
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
+	url = https://github.com/gogs/docs-api.git`,
			maxLineChars: 30,
			expDiff: &Diff{
				files: []*DiffFile{
					{
						name:  ".gitmodules",
						typ:   DiffFileAdd,
						index: "6abde17",
						sections: []*DiffSection{
							{
								lines: []*DiffLine{
									{
										typ:       DiffLineSection,
										content:   "@@ -0,0 +1,3 @@",
										leftLine:  0,
										rightLine: 0,
									},
									{
										typ:       DiffLineAdd,
										content:   `+[submodule "gogs/docs-api"]`,
										leftLine:  0,
										rightLine: 1,
									},
									{
										typ: DiffLineAdd,
										content: `+	path = gogs/docs-api`,
										leftLine:  0,
										rightLine: 2,
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
			maxLines:     2,
			maxLineChars: 30,
			maxFiles:     1,
			expDiff: &Diff{
				files: []*DiffFile{
					{
						name:  ".gitmodules",
						typ:   DiffFileAdd,
						index: "6abde17",
						sections: []*DiffSection{
							{
								lines: []*DiffLine{
									{
										typ:       DiffLineSection,
										content:   "@@ -0,0 +1,3 @@",
										leftLine:  0,
										rightLine: 0,
									},
									{
										typ:       DiffLineAdd,
										content:   `+[submodule "gogs/docs-api"]`,
										leftLine:  0,
										rightLine: 1,
									},
									{
										typ: DiffLineAdd,
										content: `+	path = gogs/docs-api`,
										leftLine:  0,
										rightLine: 2,
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
			done := make(chan SteamParsePatchResult)
			go SteamParsePatch(strings.NewReader(test.input), done, test.maxLines, test.maxLineChars, test.maxFiles)
			result := <-done
			if result.Err != nil {
				t.Fatal(result.Err)
			}

			assert.Equal(t, test.expDiff, result.Diff)
		})
	}
}
