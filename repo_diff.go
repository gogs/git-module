package git

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"unicode/utf8"

	"github.com/Unknwon/com"
	"github.com/gogits/chardet"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

type DiffLineType uint8

const (
	DIFF_LINE_PLAIN DiffLineType = iota + 1
	DIFF_LINE_ADD
	DIFF_LINE_DEL
	DIFF_LINE_SECTION
)

type DiffFileType uint8

const (
	DIFF_FILE_ADD DiffFileType = iota + 1
	DIFF_FILE_CHANGE
	DIFF_FILE_DEL
	DIFF_FILE_RENAME
)

type DiffLine struct {
	LeftIdx  int
	RightIdx int
	Type     DiffLineType
	Content  string
}

func (d *DiffLine) GetType() int {
	return int(d.Type)
}

type DiffSection struct {
	Name  string
	Lines []*DiffLine
}

type DiffFile struct {
	Name               string
	OldName            string
	Index              string // 40-byte SHA, Changed/New: new SHA; Deleted: old SHA
	Addition, Deletion int
	Type               DiffFileType
	IsCreated          bool
	IsDeleted          bool
	IsBin              bool
	IsRenamed          bool
	IsSubmodule        bool
	Sections           []*DiffSection
	IsIncomplete       bool
}

func (diffFile *DiffFile) GetType() int {
	return int(diffFile.Type)
}

type Diff struct {
	TotalAddition, TotalDeletion int
	Files                        []*DiffFile
	IsIncomplete                 bool
}

func (diff *Diff) NumFiles() int {
	return len(diff.Files)
}

const DIFF_HEAD = "diff --git "
const AnsiCharset = "ascii"

func DetectEncoding(content []byte) (string, error) {
	if utf8.Valid(content) {
		return "UTF-8", nil
	}

	result, err := chardet.NewTextDetector().DetectBest(content)
	if result.Charset != "UTF-8" {
		return AnsiCharset, err
	}

	return result.Charset, err
}

func ParsePatch(maxLines, maxLineCharacteres, maxFiles int, reader io.Reader) (*Diff, error) {
	var (
		diff = &Diff{Files: make([]*DiffFile, 0)}

		curFile    *DiffFile
		curSection = &DiffSection{
			Lines: make([]*DiffLine, 0, 10),
		}

		leftLine, rightLine int
		lineCount           int
		curFileLinesCount   int
	)

	input := bufio.NewReader(reader)
	isEOF := false
	for !isEOF {
		line, err := input.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				isEOF = true
			} else {
				return nil, fmt.Errorf("ReadString: %v", err)
			}
		}

		if len(line) > 0 && line[len(line)-1] == '\n' {
			// Remove line break.
			line = line[:len(line)-1]
		}

		if strings.HasPrefix(line, "+++ ") || strings.HasPrefix(line, "--- ") || len(line) == 0 {
			continue
		}

		curFileLinesCount++
		lineCount++

		// Diff data too large, we only show the first about maxlines lines
		if curFileLinesCount >= maxLines || len(line) >= maxLineCharacteres {
			curFile.IsIncomplete = true
		}

		switch {
		case line[0] == ' ':
			diffLine := &DiffLine{Type: DIFF_LINE_PLAIN, Content: line, LeftIdx: leftLine, RightIdx: rightLine}
			leftLine++
			rightLine++
			curSection.Lines = append(curSection.Lines, diffLine)
			continue
		case line[0] == '@':
			curSection = &DiffSection{}
			curFile.Sections = append(curFile.Sections, curSection)
			ss := strings.Split(line, "@@")
			diffLine := &DiffLine{Type: DIFF_LINE_SECTION, Content: line}
			curSection.Lines = append(curSection.Lines, diffLine)

			// Parse line number.
			ranges := strings.Split(ss[1][1:], " ")
			leftLine, _ = com.StrTo(strings.Split(ranges[0], ",")[0][1:]).Int()
			if len(ranges) > 1 {
				rightLine, _ = com.StrTo(strings.Split(ranges[1], ",")[0]).Int()
			} else {
				rightLine = leftLine
			}
			continue
		case line[0] == '+':
			curFile.Addition++
			diff.TotalAddition++
			diffLine := &DiffLine{Type: DIFF_LINE_ADD, Content: line, RightIdx: rightLine}
			rightLine++
			curSection.Lines = append(curSection.Lines, diffLine)
			continue
		case line[0] == '-':
			curFile.Deletion++
			diff.TotalDeletion++
			diffLine := &DiffLine{Type: DIFF_LINE_DEL, Content: line, LeftIdx: leftLine}
			if leftLine > 0 {
				leftLine++
			}
			curSection.Lines = append(curSection.Lines, diffLine)
		case strings.HasPrefix(line, "Binary"):
			curFile.IsBin = true
			continue
		}

		// Get new file.
		if strings.HasPrefix(line, DIFF_HEAD) {
			middle := -1

			// Note: In case file name is surrounded by double quotes (it happens only in git-shell).
			// e.g. diff --git "a/xxx" "b/xxx"
			hasQuote := line[len(DIFF_HEAD)] == '"'
			if hasQuote {
				middle = strings.Index(line, ` "b/`)
			} else {
				middle = strings.Index(line, " b/")
			}

			beg := len(DIFF_HEAD)
			a := line[beg+2 : middle]
			b := line[middle+3:]
			if hasQuote {
				a = string(UnescapeChars([]byte(a[1 : len(a)-1])))
				b = string(UnescapeChars([]byte(b[1 : len(b)-1])))
			}

			curFile = &DiffFile{
				Name:     a,
				Type:     DIFF_FILE_CHANGE,
				Sections: make([]*DiffSection, 0, 10),
			}
			diff.Files = append(diff.Files, curFile)
			if len(diff.Files) >= maxFiles {
				diff.IsIncomplete = true
				io.Copy(ioutil.Discard, reader)
				break
			}
			curFileLinesCount = 0

			// Check file diff type and submodule.
		CHECK_TYPE:
			for {
				line, err := input.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						isEOF = true
					} else {
						return nil, fmt.Errorf("ReadString: %v", err)
					}
				}

				switch {
				case strings.HasPrefix(line, "new file"):
					curFile.Type = DIFF_FILE_ADD
					curFile.IsCreated = true
					curFile.IsSubmodule = strings.HasSuffix(line, " 160000\n")
				case strings.HasPrefix(line, "deleted"):
					curFile.Type = DIFF_FILE_DEL
					curFile.IsDeleted = true
					curFile.IsSubmodule = strings.HasSuffix(line, " 160000\n")
				case strings.HasPrefix(line, "index"):
					if curFile.IsDeleted {
						curFile.Index = line[6:46]
					} else if len(line) >= 88 {
						curFile.Index = line[49:88]
					} else {
						curFile.Index = curFile.Name
					}
					break CHECK_TYPE
				case strings.HasPrefix(line, "similarity index 100%"):
					curFile.Type = DIFF_FILE_RENAME
					curFile.IsRenamed = true
					curFile.OldName = curFile.Name
					curFile.Name = b
					curFile.Index = b
					break CHECK_TYPE
				}
			}
		}
	}

	// FIXME: detect encoding while parsing.
	var buf bytes.Buffer
	for _, f := range diff.Files {
		buf.Reset()
		for _, sec := range f.Sections {
			for _, l := range sec.Lines {
				buf.WriteString(l.Content)
				buf.WriteString("\n")
			}
		}
		charsetLabel, err := DetectEncoding(buf.Bytes())
		if charsetLabel != "UTF-8" && err == nil {
			encoding, _ := charset.Lookup(charsetLabel)
			if encoding != nil {
				d := encoding.NewDecoder()
				for _, sec := range f.Sections {
					for _, l := range sec.Lines {
						if c, _, err := transform.String(d, l.Content); err == nil {
							l.Content = c
						}
					}
				}
			}
		}
	}
	return diff, nil
}

func (repo *Repository) GetDiffRange(beforeCommitID, afterCommitID string, maxLines, maxLineCharacteres, maxFiles int) (*Diff, error) {
	commit, err := repo.GetCommit(afterCommitID)
	if err != nil {
		return nil, err
	}

	var diff *Command
	// if "after" commit given
	if len(beforeCommitID) == 0 {
		// First commit of repository.
		if commit.ParentCount() == 0 {
			diff = NewCommand("show", "--full-index", afterCommitID)
		} else {
			c, _ := commit.Parent(0)
			diff = NewCommand("diff", "--full-index", "-M", c.ID.String(), afterCommitID)
		}
	} else {
		diff = NewCommand("diff", "--full-index", "-M", beforeCommitID, afterCommitID)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	err = diff.RunInDirPipeline(repo.Path, stdout, stderr)
	if err != nil {
		return nil, fmt.Errorf("StdoutPipe: %v", err)
	}

	diffContent, err := ParsePatch(maxLines, maxLineCharacteres, maxFiles, stdout)
	if err != nil {
		return nil, fmt.Errorf("ParsePatch: %v", err)
	}

	return diffContent, nil
}

type RawDiffType string

const (
	RAW_DIFF_NORMAL RawDiffType = "diff"
	RAW_DIFF_PATCH  RawDiffType = "patch"
)

// GetRawDiff dumps diff results of repository in given commit ID to io.Writer.
// TODO: move this function to gogits/git-module
func (repo *Repository) GetRawDiff(commitID string, diffType RawDiffType, writer io.Writer) error {
	commit, err := repo.GetCommit(commitID)
	if err != nil {
		return fmt.Errorf("GetCommit: %v", err)
	}

	var cmd *Command
	switch diffType {
	case RAW_DIFF_NORMAL:
		if commit.ParentCount() == 0 {
			cmd = NewCommand("show", commitID)
		} else {
			c, _ := commit.Parent(0)
			cmd = NewCommand("diff", "-M", c.ID.String(), commitID)
		}
	case RAW_DIFF_PATCH:
		if commit.ParentCount() == 0 {
			cmd = NewCommand("format-patch", "--no-signature", "--stdout", "--root", commitID)
		} else {
			c, _ := commit.Parent(0)
			query := fmt.Sprintf("%s...%s", commitID, c.ID.String())
			cmd = NewCommand("format-patch", "--no-signature", "--stdout", query)
		}
	default:
		return fmt.Errorf("invalid diffType: %s", diffType)
	}

	stderr := &bytes.Buffer{}
	if err = cmd.RunInDirPipeline(repo.Path, writer, stderr); err != nil {
		return fmt.Errorf("Run: %v - %s", err, stderr)
	}
	return nil
}

func (repo *Repository) GetDiffCommit(commitID string, maxLines, maxLineCharacteres, maxFiles int) (*Diff, error) {
	return repo.GetDiffRange("", commitID, maxLines, maxLineCharacteres, maxFiles)
}
