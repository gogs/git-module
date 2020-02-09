// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bufio"
	"bytes"
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

	numAdditions int
	numDeletions int
}

// Lines returns lines in the section.
func (s *DiffSection) Lines() []*DiffLine {
	return s.lines
}

// NumLines returns the number of lines in the section.
func (s *DiffSection) NumLines() int {
	return len(s.lines)
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

type diffParser struct {
	*bufio.Reader
	maxFiles     int
	maxFileLines int
	maxLineChars int

	// The next line that hasn't been processed. It is used to determine what kind of process should go in.
	buffer []byte
	isEOF  bool
}

func (p *diffParser) readLine() error {
	if p.buffer != nil {
		return nil
	}

	var err error
	p.buffer, err = p.ReadBytes('\n')
	if err != nil {
		if err != io.EOF {
			return fmt.Errorf("read string: %v", err)
		}

		p.isEOF = true
	}

	// Remove line break
	if len(p.buffer) > 0 && p.buffer[len(p.buffer)-1] == '\n' {
		p.buffer = p.buffer[:len(p.buffer)-1]
	}
	return nil
}

var diffHead = []byte("diff --git ")

func (p *diffParser) parseFileHeader() (*DiffFile, error) {
	line := string(p.buffer)
	p.buffer = nil

	// Note: In case file name is surrounded by double quotes (it happens only in git-shell).
	// e.g. diff --git "a/xxx" "b/xxx"
	var middle int
	hasQuote := line[len(diffHead)] == '"'
	if hasQuote {
		middle = strings.Index(line, ` "b/`)
	} else {
		middle = strings.Index(line, ` b/`)
	}

	beg := len(diffHead)
	a := line[beg+2 : middle]
	b := line[middle+3:]
	if hasQuote {
		a = string(UnescapeChars([]byte(a[1 : len(a)-1])))
		b = string(UnescapeChars([]byte(b[1 : len(b)-1])))
	}

	file := &DiffFile{
		name: a,
		typ:  DiffFileChange,
	}

	// Check file diff type and submodule
	var err error
checkType:
	for !p.isEOF {
		if err = p.readLine(); err != nil {
			return nil, err
		}

		line := string(p.buffer)
		p.buffer = nil

		if len(line) == 0 {
			continue
		}

		switch {
		case strings.HasPrefix(line, "new file"):
			file.typ = DiffFileAdd
			file.isSubmodule = strings.HasSuffix(line, " 160000")
		case strings.HasPrefix(line, "deleted"):
			file.typ = DiffFileDelete
			file.isSubmodule = strings.HasSuffix(line, " 160000")
		case strings.HasPrefix(line, "index"): // e.g. index ee791be..9997571 100644
			fields := strings.Fields(line[6:])
			shas := strings.Split(fields[0], "..")
			if len(shas) != 2 {
				return nil, errors.New("malformed index: expect two SHAs in the form of <old>..<new>")
			}

			if file.IsDeleted() {
				file.index = shas[0]
			} else {
				file.index = shas[1]
			}
			break checkType
		case strings.HasPrefix(line, "similarity index 100%"):
			file.typ = DiffFileRename
			file.oldName = a
			file.name = b
			break checkType
		case strings.HasPrefix(line, "old mode"):
			break checkType
		}
	}

	return file, nil
}

func (p *diffParser) parseSection() (_ *DiffSection, isIncomplete bool, _ error) {
	line := string(p.buffer)
	p.buffer = nil

	section := &DiffSection{
		lines: []*DiffLine{
			{
				typ:     DiffLineSection,
				content: line,
			},
		},
	}

	// Parse line number, e.g. @@ -0,0 +1,3 @@
	var leftLine, rightLine int
	ss := strings.Split(line, "@@")
	ranges := strings.Split(ss[1][1:], " ")
	leftLine, _ = strconv.Atoi(strings.Split(ranges[0], ",")[0][1:])
	if len(ranges) > 1 {
		rightLine, _ = strconv.Atoi(strings.Split(ranges[1], ",")[0])
	} else {
		rightLine = leftLine
	}

	var err error
	for !p.isEOF {
		if err = p.readLine(); err != nil {
			return nil, false, err
		}

		if len(p.buffer) == 0 {
			p.buffer = nil
			continue
		}

		// Make sure we're still in the section. If not, we're done with this section.
		if p.buffer[0] != ' ' &&
			p.buffer[0] != '+' &&
			p.buffer[0] != '-' {
			return section, false, nil
		}

		line := string(p.buffer)
		p.buffer = nil

		// Too many characters in a single diff line
		if p.maxLineChars > 0 && len(line) > p.maxLineChars {
			return section, true, nil
		}

		switch line[0] {
		case ' ':
			section.lines = append(section.lines, &DiffLine{
				typ:       DiffLinePlain,
				content:   line,
				leftLine:  leftLine,
				rightLine: rightLine,
			})
			leftLine++
			rightLine++
		case '+':
			section.lines = append(section.lines, &DiffLine{
				typ:       DiffLineAdd,
				content:   line,
				rightLine: rightLine,
			})
			section.numAdditions++
			rightLine++
		case '-':
			section.lines = append(section.lines, &DiffLine{
				typ:      DiffLineDelete,
				content:  line,
				leftLine: leftLine,
			})
			section.numDeletions++
			if leftLine > 0 {
				leftLine++
			}
		}
	}

	return section, false, nil
}

func (p *diffParser) parse() (*Diff, error) {
	diff := new(Diff)
	file := new(DiffFile)
	currentFileLines := 0

	var err error
	for !p.isEOF {
		if err = p.readLine(); err != nil {
			return nil, err
		}

		if len(p.buffer) == 0 ||
			bytes.HasPrefix(p.buffer, []byte("+++ ")) ||
			bytes.HasPrefix(p.buffer, []byte("--- ")) {
			p.buffer = nil
			continue
		}

		// Found new file
		if bytes.HasPrefix(p.buffer, diffHead) {
			// Check if reached maximum number of files
			if p.maxFiles > 0 && len(diff.files) >= p.maxFiles {
				diff.isIncomplete = true
				_, _ = io.Copy(ioutil.Discard, p)
				break
			}

			file, err = p.parseFileHeader()
			if err != nil {
				return nil, err
			}
			diff.files = append(diff.files, file)

			currentFileLines = 0
			continue
		}

		if file == nil || file.isIncomplete {
			p.buffer = nil
			continue
		}

		if bytes.HasPrefix(p.buffer, []byte("Binary")) {
			p.buffer = nil
			file.isBinary = true
			continue
		}

		// Loop until we found section header
		if p.buffer[0] != '@' {
			p.buffer = nil
			continue
		}

		// Too many diff lines for the file
		if p.maxFileLines > 0 && currentFileLines > p.maxFileLines {
			file.isIncomplete = true
			diff.isIncomplete = true
			continue
		}

		section, isIncomplete, err := p.parseSection()
		if err != nil {
			return nil, err
		}
		file.sections = append(file.sections, section)
		file.numAdditions += section.numAdditions
		file.numDeletions += section.numDeletions
		diff.totalAdditions += section.numAdditions
		diff.totalDeletions += section.numDeletions
		currentFileLines += section.NumLines()
		if isIncomplete {
			file.isIncomplete = true
			diff.isIncomplete = true
		}
	}

	return diff, nil
}

// StreamParseDiff parses the diff read from the given io.Reader. It does parse-on-read to minimize
// the time spent on huge diffs. It accepts a channel to notify and send error (if any) to the caller
// when the process is done. Therefore, this method should be called in a goroutine asynchronously.
func StreamParseDiff(r io.Reader, done chan<- SteamParsePatchResult, maxFiles, maxFileLines, maxLineChars int) {
	p := &diffParser{
		Reader:       bufio.NewReader(r),
		maxFiles:     maxFiles,
		maxFileLines: maxFileLines,
		maxLineChars: maxLineChars,
	}
	diff, err := p.parse()
	done <- SteamParsePatchResult{
		Diff: diff,
		Err:  err,
	}
	return
}
