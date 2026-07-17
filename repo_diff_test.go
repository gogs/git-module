package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepository_Diff(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		rev          string
		maxFiles     int
		maxFileLines int
		maxLineChars int
		opt          DiffOptions
		expDiff      *Diff
	}{
		{
			rev: testrepoMarks[27].String(),
			expDiff: &Diff{
				Files: []*DiffFile{
					{
						Name:         "fix.txt",
						Type:         DiffFileDelete,
						Index:        testEmptyShaID,
						OldIndex:     testrepoMarks[54].String(),
						Sections:     nil,
						numAdditions: 0,
						numDeletions: 0,
						oldName:      "fix.txt",
						isBinary:     false,
						isSubmodule:  false,
						isIncomplete: false,
						mode:         0100644,
						oldMode:      0100644,
					},
				},
				totalAdditions: 0,
				totalDeletions: 0,
				isIncomplete:   false,
			},
		},
		{
			rev: testrepoMarks[1].String(),
			expDiff: &Diff{
				Files: []*DiffFile{
					{
						Name:     "README.txt",
						Type:     DiffFileAdd,
						Index:    testrepoMarks[51].String(),
						OldIndex: testEmptyShaID,
						Sections: []*DiffSection{
							{
								Lines: []*DiffLine{
									{
										Type:    DiffLineSection,
										Content: "@@ -0,0 +1,11 @@",
									},
									{
										Type:      DiffLineAdd,
										Content:   "+This is a sample project students can use during Matthew's Git class.",
										LeftLine:  0,
										RightLine: 1,
									},
									{
										Type:      DiffLineAdd,
										Content:   "+",
										LeftLine:  0,
										RightLine: 2,
									},
									{
										Type:      DiffLineAdd,
										Content:   "+We can have a bit of fun with this repo, knowing that we can always reset it to a known good state.  We can apply labels, and branch, then add new code and merge it in to the master branch.",
										LeftLine:  0,
										RightLine: 3,
									},
									{
										Type:      DiffLineAdd,
										Content:   "+",
										LeftLine:  0,
										RightLine: 4,
									},
									{
										Type:      DiffLineAdd,
										Content:   "+As a quick reminder, this came from one of three locations in either SSH, Git, or HTTPS format:",
										LeftLine:  0,
										RightLine: 5,
									},
									{
										Type:      DiffLineAdd,
										Content:   "+",
										LeftLine:  0,
										RightLine: 6,
									},
									{
										Type:      DiffLineAdd,
										Content:   "+* git@github.com:matthewmccullough/hellogitworld.git",
										LeftLine:  0,
										RightLine: 7,
									},
									{
										Type:      DiffLineAdd,
										Content:   "+* git://github.com/matthewmccullough/hellogitworld.git",
										LeftLine:  0,
										RightLine: 8,
									},
									{
										Type:      DiffLineAdd,
										Content:   "+* https://matthewmccullough@github.com/matthewmccullough/hellogitworld.git",
										LeftLine:  0,
										RightLine: 9,
									},
									{
										Type:      DiffLineAdd,
										Content:   "+",
										LeftLine:  0,
										RightLine: 10,
									},
									{
										Type:      DiffLineAdd,
										Content:   "+We can, as an example effort, even modify this README and change it as if it were source code for the purposes of the class.",
										LeftLine:  0,
										RightLine: 11,
									},
								},
								numAdditions: 11,
							},
						},
						numAdditions: 11,
						numDeletions: 0,
						oldName:      "README.txt",
						isBinary:     false,
						isSubmodule:  false,
						isIncomplete: false,
						mode:         0100644,
						oldMode:      0100644,
					},
					{
						Name:     "resources/labels.properties",
						Type:     DiffFileAdd,
						Index:    testrepoMarks[52].String(),
						OldIndex: testEmptyShaID,
						Sections: []*DiffSection{
							{
								Lines: []*DiffLine{
									{
										Type:    DiffLineSection,
										Content: "@@ -0,0 +1,4 @@",
									},
									{
										Type:      DiffLineAdd,
										Content:   "+app.title=Our App",
										LeftLine:  0,
										RightLine: 1,
									},
									{
										Type:      DiffLineAdd,
										Content:   "+app.welcome=Welcome to the application",
										LeftLine:  0,
										RightLine: 2,
									},
									{
										Type:      DiffLineAdd,
										Content:   "+app.goodbye=We hope you enjoyed using our application",
										LeftLine:  0,
										RightLine: 3,
									},
									{
										Type:      DiffLineAdd,
										Content:   "+cli.usage=This application doesn't use a command line interface",
										LeftLine:  0,
										RightLine: 4,
									},
								},
								numAdditions: 4,
							},
						},
						numAdditions: 4,
						numDeletions: 0,
						oldName:      "resources/labels.properties",
						isBinary:     false,
						isSubmodule:  false,
						isIncomplete: false,
						mode:         0100644,
						oldMode:      0100644,
					},
					{
						Name:     "src/Main.groovy",
						Type:     DiffFileAdd,
						Index:    testrepoMarks[53].String(),
						OldIndex: testEmptyShaID,
						Sections: []*DiffSection{
							{
								Lines: []*DiffLine{
									{
										Type:    DiffLineSection,
										Content: "@@ -0,0 +1,6 @@",
									},
									{
										Type:      DiffLineAdd,
										Content:   "+def name = \"Matthew\"",
										LeftLine:  0,
										RightLine: 1,
									},
									{
										Type:      DiffLineAdd,
										Content:   "+",
										LeftLine:  0,
										RightLine: 2,
									},
									{
										Type:      DiffLineAdd,
										Content:   "+println \"Hello ${name}\"",
										LeftLine:  0,
										RightLine: 3,
									},
									{
										Type:      DiffLineAdd,
										Content:   "+",
										LeftLine:  0,
										RightLine: 4,
									},
									{
										Type:      DiffLineAdd,
										Content:   "+int programmingPoints = 10",
										LeftLine:  0,
										RightLine: 5,
									},
									{
										Type:      DiffLineAdd,
										Content:   "+println \"${name} has at least ${programmingPoints} programming points.\"",
										LeftLine:  0,
										RightLine: 6,
									},
								},
								numAdditions: 6,
							},
						},
						numAdditions: 6,
						numDeletions: 0,
						oldName:      "src/Main.groovy",
						isBinary:     false,
						isSubmodule:  false,
						isIncomplete: false,
						mode:         0100644,
						oldMode:      0100644,
					},
				},
				totalAdditions: 21,
				totalDeletions: 0,
				isIncomplete:   false,
			},
		},
		{
			rev: testrepoMarks[27].String(),
			opt: DiffOptions{
				Base: testrepoMarks[26].String(),
			},
			expDiff: &Diff{
				Files: []*DiffFile{
					{
						Name:         "fix.txt",
						Type:         DiffFileDelete,
						Index:        testEmptyShaID,
						OldIndex:     testrepoMarks[54].String(),
						Sections:     nil,
						numAdditions: 0,
						numDeletions: 0,
						oldName:      "fix.txt",
						isBinary:     false,
						isSubmodule:  false,
						isIncomplete: false,
						mode:         0100644,
						oldMode:      0100644,
					},
				},
				totalAdditions: 0,
				totalDeletions: 0,
				isIncomplete:   false,
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			diff, err := testrepo.Diff(ctx, test.rev, test.maxFiles, test.maxFileLines, test.maxLineChars, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expDiff, diff)
		})
	}
}

func TestRepository_RawDiff(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid revision", func(t *testing.T) {
		err := testrepo.RawDiff(ctx, "bad_revision", "bad_diff_type", nil)
		assert.Equal(t, ErrRevisionNotExist, err)
	})

	t.Run("invalid diffType", func(t *testing.T) {
		err := testrepo.RawDiff(ctx, testrepoMarks[27].String(), "bad_diff_type", nil)
		assert.Equal(t, errors.New("invalid diffType: bad_diff_type"), err)
	})

	tests := []struct {
		rev       string
		diffType  RawDiffFormat
		opt       RawDiffOptions
		expOutput string
	}{
		{
			rev:      testrepoMarks[27].String(),
			diffType: RawDiffNormal,
			expOutput: fmt.Sprintf(`diff --git a/fix.txt b/fix.txt
deleted file mode 100644
index %s..%s
`, testrepoMarks[54].String(), testEmptyShaID),
		},
		{
			rev:      testrepoMarks[27].String(),
			diffType: RawDiffPatch,
			expOutput: fmt.Sprintf(`Date: Sun, 9 Feb 2020 17:22:24 +0800
Subject: [PATCH] Delete fix.txt

---
 fix.txt | 0
 1 file changed, 0 insertions(+), 0 deletions(-)
 delete mode 100644 fix.txt

diff --git a/fix.txt b/fix.txt
deleted file mode 100644
index %s..%s
`, testrepoMarks[54].String(), testEmptyShaID),
		},
		{
			rev:      testrepoMarks[1].String(),
			diffType: RawDiffNormal,
			expOutput: fmt.Sprintf(`commit %[1]s
Author: Matthew McCullough <matthewm@ambientideas.com>
Date:   Mon Nov 24 21:22:01 2008 -0700

    Addition of the README and basic Groovy source samples.
    
    - Addition of the README.txt file explaining what this repository is all about.
    - Addition of Groovy sample source.
    - Addition of sample resource Properties file.

diff --git a/README.txt b/README.txt
new file mode 100644
index %[2]s..%[3]s
--- /dev/null
+++ b/README.txt
@@ -0,0 +1,11 @@
+This is a sample project students can use during Matthew's Git class.
+
+We can have a bit of fun with this repo, knowing that we can always reset it to a known good state.  We can apply labels, and branch, then add new code and merge it in to the master branch.
+
+As a quick reminder, this came from one of three locations in either SSH, Git, or HTTPS format:
+
+* git@github.com:matthewmccullough/hellogitworld.git
+* git://github.com/matthewmccullough/hellogitworld.git
+* https://matthewmccullough@github.com/matthewmccullough/hellogitworld.git
+
+We can, as an example effort, even modify this README and change it as if it were source code for the purposes of the class.
\ No newline at end of file
diff --git a/resources/labels.properties b/resources/labels.properties
new file mode 100644
index %[2]s..%[4]s
--- /dev/null
+++ b/resources/labels.properties
@@ -0,0 +1,4 @@
+app.title=Our App
+app.welcome=Welcome to the application
+app.goodbye=We hope you enjoyed using our application
+cli.usage=This application doesn't use a command line interface
diff --git a/src/Main.groovy b/src/Main.groovy
new file mode 100644
index %[2]s..%[5]s
--- /dev/null
+++ b/src/Main.groovy
@@ -0,0 +1,6 @@
+def name = "Matthew"
+
+println "Hello ${name}"
+
+int programmingPoints = 10
+println "${name} has at least ${programmingPoints} programming points."
\ No newline at end of file
`, testrepoMarks[1].String(), testEmptyShaID, testrepoMarks[51].String(), testrepoMarks[52].String(), testrepoMarks[53].String()),
		},
		{
			rev:      testrepoMarks[1].String(),
			diffType: RawDiffPatch,
			expOutput: fmt.Sprintf(`Date: Mon, 24 Nov 2008 21:22:01 -0700
Subject: [PATCH] Addition of the README and basic Groovy source samples.

- Addition of the README.txt file explaining what this repository is all about.
- Addition of Groovy sample source.
- Addition of sample resource Properties file.
---
 README.txt                  | 11 +++++++++++
 resources/labels.properties |  4 ++++
 src/Main.groovy             |  6 ++++++
 3 files changed, 21 insertions(+)
 create mode 100644 README.txt
 create mode 100644 resources/labels.properties
 create mode 100644 src/Main.groovy

diff --git a/README.txt b/README.txt
new file mode 100644
index %[2]s..%[3]s
--- /dev/null
+++ b/README.txt
@@ -0,0 +1,11 @@
+This is a sample project students can use during Matthew's Git class.
+
+We can have a bit of fun with this repo, knowing that we can always reset it to a known good state.  We can apply labels, and branch, then add new code and merge it in to the master branch.
+
+As a quick reminder, this came from one of three locations in either SSH, Git, or HTTPS format:
+
+* git@github.com:matthewmccullough/hellogitworld.git
+* git://github.com/matthewmccullough/hellogitworld.git
+* https://matthewmccullough@github.com/matthewmccullough/hellogitworld.git
+
+We can, as an example effort, even modify this README and change it as if it were source code for the purposes of the class.
\ No newline at end of file
diff --git a/resources/labels.properties b/resources/labels.properties
new file mode 100644
index %[2]s..%[4]s
--- /dev/null
+++ b/resources/labels.properties
@@ -0,0 +1,4 @@
+app.title=Our App
+app.welcome=Welcome to the application
+app.goodbye=We hope you enjoyed using our application
+cli.usage=This application doesn't use a command line interface
diff --git a/src/Main.groovy b/src/Main.groovy
new file mode 100644
index %[2]s..%[5]s
--- /dev/null
+++ b/src/Main.groovy
@@ -0,0 +1,6 @@
+def name = "Matthew"
+
+println "Hello ${name}"
+
+int programmingPoints = 10
+println "${name} has at least ${programmingPoints} programming points."
\ No newline at end of file
`, testrepoMarks[1].String(), testEmptyShaID, testrepoMarks[51].String(), testrepoMarks[52].String(), testrepoMarks[53].String()),
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			buf := new(bytes.Buffer)
			err := testrepo.RawDiff(ctx, test.rev, test.diffType, buf, test.opt)
			if err != nil {
				t.Fatal(err)
			}
			output := buf.String()

			// Only check the content after "Date:" line, which is deterministic.
			i := strings.Index(output, "Date:")
			if i > 0 && test.diffType == RawDiffPatch {
				output = output[i:]
			}

			assert.Equal(t, test.expOutput, output)
		})
	}
}

func TestRepository_DiffBinary(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		base      string
		head      string
		opt       DiffBinaryOptions
		expOutput string
	}{
		{
			base: testrepoMarks[28].String(),
			head: testrepoMarks[29].String(),
			expOutput: fmt.Sprintf(`diff --git a/.gitmodules b/.gitmodules
new file mode 100644
index %[1]s..%[2]s
--- /dev/null
+++ b/.gitmodules
@@ -0,0 +1,3 @@
+[submodule "gogs/docs-api"]
+	path = gogs/docs-api
+	url = https://github.com/gogs/docs-api.git
diff --git a/gogs/docs-api b/gogs/docs-api
new file mode 160000
index %[1]s..%[3]s
--- /dev/null
+++ b/gogs/docs-api
@@ -0,0 +1 @@
+Subproject commit %[3]s
`, testEmptyShaID, testrepoMarks[37].String(), submoduleSHA.String()),
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			p, err := testrepo.DiffBinary(ctx, test.base, test.head, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expOutput, string(p))
		})
	}
}
