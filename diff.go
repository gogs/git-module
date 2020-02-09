// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

// DiffLineType is the line type in diff.
type DiffLineType uint8

// A list of different line types.
const (
	DiffLinePlain DiffLineType = iota + 1
	DiffLineAdd
	DiffLineDelete
	DiffLineSection
)

// DiffFileType is the file status in diff.
type DiffFileType uint8

// A list of different file statuses.
const (
	DiffFileAdd DiffFileType = iota + 1
	DiffFileChange
	DiffFileDelete
	DiffFileRename
)

// DiffLine represents a line in diff.
type DiffLine struct {
	typ       DiffLineType
	content   string
	leftLine  int
	rightLine int
}

// Type returns the type of the line.
func (l *DiffLine) Type() DiffLineType {
	return l.typ
}

// Content returns the content of the line.
func (l *DiffLine) Content() string {
	return l.content
}

// Left returns the left line number.
func (l *DiffLine) Left() int {
	return l.leftLine
}

// Right returns the right line number.
func (l *DiffLine) Right() int {
	return l.rightLine
}

// DiffSection represents a section in diff.
type DiffSection struct {
	lines []*DiffLine
}

// Lines returns lines in the section.
func (s *DiffSection) Lines() []*DiffLine {
	return s.lines
}

// Line returns a specific line by given type and line number in a section.
func (s *DiffSection) Line(typ DiffLineType, line int) *DiffLine {
	var (
		difference      = 0
		addCount        = 0
		delCount        = 0
		matchedDiffLine *DiffLine
	)

loop:
	for _, diffLine := range s.lines {
		switch diffLine.typ {
		case DiffLineAdd:
			addCount++
		case DiffLineDelete:
			delCount++
		default:
			if matchedDiffLine != nil {
				break loop
			}
			difference = diffLine.rightLine - diffLine.leftLine
			addCount = 0
			delCount = 0
		}

		switch typ {
		case DiffLineDelete:
			if diffLine.rightLine == 0 && diffLine.leftLine == line-difference {
				matchedDiffLine = diffLine
			}
		case DiffLineAdd:
			if diffLine.leftLine == 0 && diffLine.rightLine == line+difference {
				matchedDiffLine = diffLine
			}
		}
	}

	if addCount == delCount {
		return matchedDiffLine
	}
	return nil
}

// DiffFile represents a file in diff.
type DiffFile struct {
	name     string
	typ      DiffFileType
	index    string // Changed/New: new SHA; Deleted: old SHA
	sections []*DiffSection

	numAdditions int
	numDeletions int

	oldName string

	isBinary     bool
	isSubmodule  bool
	isIncomplete bool
}

// Name returns the name of the file.
func (f *DiffFile) Name() string {
	return f.name
}

// Type returns the type of the file.
func (f *DiffFile) Type() DiffFileType {
	return f.typ
}

// Index returns the index (SHA1 hash) of the file.
func (f *DiffFile) Index() string {
	return f.index
}

// Sections returns the sections in the file.
func (f *DiffFile) Sections() []*DiffSection {
	return f.sections
}

// NumSections returns the number of sections in the file.
func (f *DiffFile) NumSections() int {
	return len(f.sections)
}

// NumAdditions returns the number of additions in the file.
func (f *DiffFile) NumAdditions() int {
	return f.numAdditions
}

// NumDeletions returns the number of deletions in the file.
func (f *DiffFile) NumDeletions() int {
	return f.numDeletions
}

// IsCreated returns true if the file is newly created.
func (f *DiffFile) IsCreated() bool {
	return f.typ == DiffFileAdd
}

// IsDeleted returns true if the file has been deleted.
func (f *DiffFile) IsDeleted() bool {
	return f.typ == DiffFileDelete
}

// IsRenamed returns true if the file has been renamed.
func (f *DiffFile) IsRenamed() bool {
	return f.typ == DiffFileRename
}

// OldName returns previous name before renaming.
func (f *DiffFile) OldName() string {
	return f.oldName
}

// IsBinary returns true if the file is in binary format.
func (f *DiffFile) IsBinary() bool {
	return f.isBinary
}

// IsSubmodule returns true if the file contains information of a submodule.
func (f *DiffFile) IsSubmodule() bool {
	return f.isSubmodule
}

// IsIncomplete returns true if the file is incomplete to the file diff.
func (f *DiffFile) IsIncomplete() bool {
	return f.isIncomplete
}

// Diff represents a Git diff.
type Diff struct {
	files []*DiffFile

	totalAdditions int
	totalDeletions int

	isIncomplete bool
}

// Files returns the files in the diff.
func (d *Diff) Files() []*DiffFile {
	return d.files
}

// NumFiles returns the number of files in the diff.
func (d *Diff) NumFiles() int {
	return len(d.files)
}

// TotalAdditions returns the total additions in the diff.
func (d *Diff) TotalAdditions() int {
	return d.totalAdditions
}

// TotalDeletions returns the total deletions in the diff.
func (d *Diff) TotalDeletions() int {
	return d.totalDeletions
}

// IsIncomplete returns true if the file is incomplete to the entire diff.
func (d *Diff) IsIncomplete() bool {
	return d.isIncomplete
}

// SteamParsePatchResult contains results of stream parsing a patch.
type SteamParsePatchResult struct {
	Diff *Diff
	Err  error
}

// SteamParsePatch parses the diff read from the given io.Reader. It does parse-on-read to minimize
// the time spent on huge diffs. It accepts a channel to notify and send error (if any) to the caller
// when the process is done.
func SteamParsePatch(r io.Reader, done chan<- SteamParsePatchResult, maxLines, maxLineChars, maxFiles int) {
	var (
		diff = new(Diff)

		curFile    = new(DiffFile)
		curSection = new(DiffSection)

		leftLine, rightLine int
		curFileLineCount    int
	)
	input := bufio.NewReader(r)
	isEOF := false
	for !isEOF {
		line, err := input.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				done <- SteamParsePatchResult{
					Err: fmt.Errorf("read string: %v", err),
				}
				return
			}

			isEOF = true
		}

		// Remove line break
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}

		if strings.HasPrefix(line, "+++ ") || strings.HasPrefix(line, "--- ") || len(line) == 0 {
			continue
		}

		if !curFile.isIncomplete {
			// Diff line too large
			if line[0] == ' ' || line[0] == '+' || line[0] == '-' {
				if (maxLines > 0 && curFileLineCount > maxLines) ||
					(maxLineChars > 0 && len(line) > maxLineChars) {
					curFile.isIncomplete = true
					diff.isIncomplete = true
					continue
				}
			}

			switch {
			case line[0] == '@':
				curSection = &DiffSection{
					lines: []*DiffLine{
						{
							typ:     DiffLineSection,
							content: line,
						},
					},
				}
				curFile.sections = append(curFile.sections, curSection)

				// Parse line number, e.g. @@ -0,0 +1,3 @@
				ss := strings.Split(line, "@@")
				ranges := strings.Split(ss[1][1:], " ")
				leftLine, _ = strconv.Atoi(strings.Split(ranges[0], ",")[0][1:])
				if len(ranges) > 1 {
					rightLine, _ = strconv.Atoi(strings.Split(ranges[1], ",")[0])
				} else {
					rightLine = leftLine
				}
				curFileLineCount++
				continue
			case line[0] == ' ':
				curSection.lines = append(curSection.lines, &DiffLine{
					typ:       DiffLinePlain,
					content:   line,
					leftLine:  leftLine,
					rightLine: rightLine,
				})
				leftLine++
				rightLine++
				curFileLineCount++
				continue
			case line[0] == '+':
				curSection.lines = append(curSection.lines, &DiffLine{
					typ:       DiffLineAdd,
					content:   line,
					rightLine: rightLine,
				})
				curFile.numAdditions++
				diff.totalAdditions++
				rightLine++
				curFileLineCount++
				continue
			case line[0] == '-':
				curSection.lines = append(curSection.lines, &DiffLine{
					typ:      DiffLineDelete,
					content:  line,
					leftLine: leftLine,
				})
				curFile.numDeletions++
				diff.totalDeletions++
				if leftLine > 0 {
					leftLine++
				}
				curFileLineCount++
				continue
			case strings.HasPrefix(line, "Binary"):
				curFile.isBinary = true
				continue
			}
		}

		// Get new file
		const diffHead = "diff --git "
		if strings.HasPrefix(line, diffHead) {
			if maxFiles > 0 && len(diff.files) >= maxFiles {
				diff.isIncomplete = true
				_, _ = io.Copy(ioutil.Discard, r)
				break
			}

			var middle int

			// Note: In case file name is surrounded by double quotes (it happens only in git-shell).
			// e.g. diff --git "a/xxx" "b/xxx"
			hasQuote := line[len(diffHead)] == '"'
			if hasQuote {
				middle = strings.Index(line, ` "b/`)
			} else {
				middle = strings.Index(line, " b/")
			}

			beg := len(diffHead)
			a := line[beg+2 : middle]
			b := line[middle+3:]
			if hasQuote {
				a = string(UnescapeChars([]byte(a[1 : len(a)-1])))
				b = string(UnescapeChars([]byte(b[1 : len(b)-1])))
			}

			curFile = &DiffFile{
				name: a,
				typ:  DiffFileChange,
			}
			diff.files = append(diff.files, curFile)
			curFileLineCount = 0

			// Check file diff type and submodule
		checkType:
			for !isEOF {
				line, err := input.ReadString('\n')
				if err != nil {
					if err != io.EOF {
						done <- SteamParsePatchResult{
							Err: fmt.Errorf("read string: %v", err),
						}
						return
					}

					isEOF = true
				}

				// Remove line break
				if len(line) > 0 && line[len(line)-1] == '\n' {
					line = line[:len(line)-1]
				}

				switch {
				case strings.HasPrefix(line, "new file"):
					curFile.typ = DiffFileAdd
					curFile.isSubmodule = strings.HasSuffix(line, " 160000\n")
				case strings.HasPrefix(line, "deleted"):
					curFile.typ = DiffFileDelete
					curFile.isSubmodule = strings.HasSuffix(line, " 160000\n")
				case strings.HasPrefix(line, "index"): // e.g. index ee791be..9997571 100644
					fields := strings.Fields(line[6:])
					shas := strings.Split(fields[0], "..")
					if len(shas) != 2 {
						done <- SteamParsePatchResult{
							Err: errors.New("malformed index: expect two SHAs in the form of <old>..<new>"),
						}
						return
					}

					if curFile.IsDeleted() {
						curFile.index = shas[0]
					} else {
						curFile.index = shas[1]
					}
					break checkType
				case strings.HasPrefix(line, "similarity index 100%"):
					curFile.typ = DiffFileRename
					curFile.oldName = curFile.name
					curFile.name = b
					break checkType
				case strings.HasPrefix(line, "old mode"):
					break checkType
				}
			}
		}
	}

	done <- SteamParsePatchResult{
		Diff: diff,
	}
	return
}
