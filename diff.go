// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bufio"
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
	DiffLineDel
	DiffLineSection
)

// DiffFileType is the file status in diff.
type DiffFileType uint8

// A list of different file statuses.
const (
	DiffFileAdd DiffFileType = iota + 1
	DiffFileChange
	DiffFileDel
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
	name  string
	lines []*DiffLine
}

// Name returns the name of the section.
func (s *DiffSection) Name() string {
	return s.name
}

// Lines returns lines in the section.
func (s *DiffSection) Lines() []*DiffLine {
	return s.lines
}

// Line returns a specific line by given type and line number in a section.
func (s *DiffSection) Line(typ DiffLineType, line int) *DiffLine {
	var (
		difference    = 0
		addCount      = 0
		delCount      = 0
		matchDiffLine *DiffLine
	)

loop:
	for _, diffLine := range s.lines {
		switch diffLine.typ {
		case DiffLineAdd:
			addCount++
		case DiffLineDel:
			delCount++
		default:
			if matchDiffLine != nil {
				break loop
			}
			difference = diffLine.rightLine - diffLine.leftLine
			addCount = 0
			delCount = 0
		}

		switch typ {
		case DiffLineDel:
			if diffLine.rightLine == 0 && diffLine.leftLine == line-difference {
				matchDiffLine = diffLine
			}
		case DiffLineAdd:
			if diffLine.leftLine == 0 && diffLine.rightLine == line+difference {
				matchDiffLine = diffLine
			}
		}
	}

	if addCount == delCount {
		return matchDiffLine
	}
	return nil
}

// DiffFile represents a file in diff.
type DiffFile struct {
	name     string
	typ      DiffFileType
	index    string // 40-byte SHA, Changed/New: new SHA; Deleted: old SHA
	sections []*DiffSection

	numAdditions int
	numDeletions int

	isCreated bool
	isDeleted bool
	isRenamed bool
	oldName   string

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
	return f.isCreated
}

// IsDeleted returns true if the file has been deleted.
func (f *DiffFile) IsDeleted() bool {
	return f.isDeleted
}

// IsRenamed returns true if the file has been renamed.
func (f *DiffFile) IsRenamed() bool {
	return f.isRenamed
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

// SteamParsePatch parses the diff read from the given io.Reader. It does parse-on-read to minimize
// the time spent on huge diffs. It accepts a channel to notify and send error (if any) to the caller
// when the process is done.
func SteamParsePatch(r io.Reader, done chan<- error, maxLines, maxLineChars, maxFiles int) *Diff {
	var (
		diff = new(Diff)

		curFile    *DiffFile
		curSection = &DiffSection{
			lines: make([]*DiffLine, 0, 10),
		}

		leftLine, rightLine int
		lineCount           int
		curFileLineCount    int
	)
	input := bufio.NewReader(r)
	isEOF := false
	for !isEOF {
		// TODO: would input.ReadBytes be more memory-efficient?
		line, err := input.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				isEOF = true
			} else {
				done <- fmt.Errorf("ReadString: %v", err)
				return nil
			}
		}

		// Remove line break.
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}

		if strings.HasPrefix(line, "+++ ") || strings.HasPrefix(line, "--- ") || len(line) == 0 {
			continue
		}

		curFileLineCount++
		lineCount++

		// Diff data too large, we only show the first about maxlines lines
		if curFileLineCount >= maxLines || len(line) >= maxLineChars {
			curFile.isIncomplete = true
		}

		switch {
		case line[0] == ' ':
			diffLine := &DiffLine{typ: DiffLinePlain, content: line, leftLine: leftLine, rightLine: rightLine}
			leftLine++
			rightLine++
			curSection.lines = append(curSection.lines, diffLine)
			continue
		case line[0] == '@':
			curSection = &DiffSection{}
			curFile.sections = append(curFile.sections, curSection)
			ss := strings.Split(line, "@@")
			diffLine := &DiffLine{typ: DiffLineSection, content: line}
			curSection.lines = append(curSection.lines, diffLine)

			// Parse line number.
			ranges := strings.Split(ss[1][1:], " ")
			leftLine, _ = strconv.Atoi(strings.Split(ranges[0], ",")[0][1:])
			if len(ranges) > 1 {
				rightLine, _ = strconv.Atoi(strings.Split(ranges[1], ",")[0])
			} else {
				rightLine = leftLine
			}
			continue
		case line[0] == '+':
			curFile.numAdditions++
			diff.totalAdditions++
			diffLine := &DiffLine{typ: DiffLineAdd, content: line, rightLine: rightLine}
			rightLine++
			curSection.lines = append(curSection.lines, diffLine)
			continue
		case line[0] == '-':
			curFile.numDeletions++
			diff.totalDeletions++
			diffLine := &DiffLine{typ: DiffLineDel, content: line, leftLine: leftLine}
			if leftLine > 0 {
				leftLine++
			}
			curSection.lines = append(curSection.lines, diffLine)
		case strings.HasPrefix(line, "Binary"):
			curFile.isBinary = true
			continue
		}

		// Get new file.
		const diffHead = "diff --git "
		if strings.HasPrefix(line, diffHead) {
			middle := -1

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
				name:     a,
				typ:      DiffFileChange,
				sections: make([]*DiffSection, 0, 10),
			}
			diff.files = append(diff.files, curFile)
			if len(diff.files) >= maxFiles {
				diff.isIncomplete = true
				_, _ = io.Copy(ioutil.Discard, r)
				break
			}
			curFileLineCount = 0

			// Check file diff type and submodule.
		checkType:
			for {
				line, err := input.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						isEOF = true
					} else {
						done <- fmt.Errorf("ReadString: %v", err)
						return nil
					}
				}

				switch {
				case strings.HasPrefix(line, "new file"):
					curFile.typ = DiffFileAdd
					curFile.isCreated = true
					curFile.isSubmodule = strings.HasSuffix(line, " 160000\n")
				case strings.HasPrefix(line, "deleted"):
					curFile.typ = DiffFileDel
					curFile.isDeleted = true
					curFile.isSubmodule = strings.HasSuffix(line, " 160000\n")
				case strings.HasPrefix(line, "index"):
					if curFile.isDeleted {
						curFile.index = line[6:46]
					} else if len(line) >= 88 {
						curFile.index = line[49:88]
					} else {
						curFile.index = curFile.name
					}
					break checkType
				case strings.HasPrefix(line, "similarity index 100%"):
					curFile.typ = DiffFileRename
					curFile.isRenamed = true
					curFile.oldName = curFile.name
					curFile.name = b
					curFile.index = b
					break checkType
				case strings.HasPrefix(line, "old mode"):
					break checkType
				}
			}
		}
	}

	done <- nil
	return diff
}
