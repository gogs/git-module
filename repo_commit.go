// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"container/list"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const RemotePrefix = "refs/remotes/"

// getRefCommitID returns the last commit ID string of given reference (branch or tag).
func (r *Repository) getRefCommitID(name string) (string, error) {
	stdout, err := NewCommand("show-ref", "--verify", name).RunInDir(r.path)
	if err != nil {
		if strings.Contains(err.Error(), "not a valid ref") {
			return "", ErrNotExist{name, ""}
		}
		return "", err
	}
	return strings.Split(string(stdout), " ")[0], nil
}

// GetBranchCommitID returns last commit ID string of given branch.
func (r *Repository) GetBranchCommitID(name string) (string, error) {
	return r.getRefCommitID(BranchPrefix + name)
}

// GetTagCommitID returns last commit ID string of given tag.
func (r *Repository) GetTagCommitID(name string) (string, error) {
	return r.getRefCommitID(TagPrefix + name)
}

// GetRemoteBranchCommitID returns last commit ID string of given remote branch.
func (r *Repository) GetRemoteBranchCommitID(name string) (string, error) {
	return r.getRefCommitID(RemotePrefix + name)
}

// parseCommitData parses commit information from the (uncompressed) raw
// data from the commit object.
// \n\n separate headers from message
func parseCommitData(data []byte) (*Commit, error) {
	commit := new(Commit)
	commit.parents = make([]SHA1, 0, 1)
	// we now have the contents of the commit object. Let's investigate...
	nextline := 0
l:
	for {
		eol := bytes.IndexByte(data[nextline:], '\n')
		switch {
		case eol > 0:
			line := data[nextline : nextline+eol]
			spacepos := bytes.IndexByte(line, ' ')
			reftype := line[:spacepos]
			switch string(reftype) {
			case "tree", "object":
				id, err := NewIDFromString(string(line[spacepos+1:]))
				if err != nil {
					return nil, err
				}
				commit.Tree.ID = id
			case "parent":
				// A commit can have one or more parents
				oid, err := NewIDFromString(string(line[spacepos+1:]))
				if err != nil {
					return nil, err
				}
				commit.parents = append(commit.parents, oid)
			case "author", "tagger":
				sig, err := newSignatureFromCommitline(line[spacepos+1:])
				if err != nil {
					return nil, err
				}
				commit.author = sig
			case "committer":
				sig, err := newSignatureFromCommitline(line[spacepos+1:])
				if err != nil {
					return nil, err
				}
				commit.committer = sig
			}
			nextline += eol + 1
		case eol == 0:
			commit.message = string(data[nextline+1:])
			break l
		default:
			break l
		}
	}
	return commit, nil
}

func (r *Repository) getCommit(id SHA1) (*Commit, error) {
	c, ok := r.commitCache.Get(id.String())
	if ok {
		log("Hit cache: %s", id)
		return c.(*Commit), nil
	}

	data, err := NewCommand("cat-file", "commit", id.String()).RunInDir(r.path)
	if err != nil {
		if strings.Contains(err.Error(), "exit status 128") {
			return nil, ErrNotExist{id.String(), ""}
		}
		return nil, err
	}

	commit, err := parseCommitData(data)
	if err != nil {
		return nil, err
	}
	commit.repo = r
	commit.id = id

	r.commitCache.Set(id.String(), commit)
	return commit, nil
}

// CommitByID returns commit object of by ID string.
func (r *Repository) CommitByID(commitID string) (*Commit, error) {
	var err error
	commitID, err = r.RevParse(commitID)
	if err != nil {
		return nil, err
	}
	id, err := NewIDFromString(commitID)
	if err != nil {
		return nil, err
	}

	return r.getCommit(id)
}

// GetBranchCommit returns the last commit of given branch.
func (r *Repository) GetBranchCommit(name string) (*Commit, error) {
	commitID, err := r.GetBranchCommitID(name)
	if err != nil {
		return nil, err
	}
	return r.CommitByID(commitID)
}

// GetTagCommit returns the commit of given tag.
func (r *Repository) GetTagCommit(name string) (*Commit, error) {
	commitID, err := r.GetTagCommitID(name)
	if err != nil {
		return nil, err
	}
	return r.CommitByID(commitID)
}

// GetRemoteBranchCommit returns the last commit of given remote branch.
func (r *Repository) GetRemoteBranchCommit(name string) (*Commit, error) {
	commitID, err := r.GetRemoteBranchCommitID(name)
	if err != nil {
		return nil, err
	}
	return r.CommitByID(commitID)
}

func (r *Repository) getCommitByPathWithID(id SHA1, relpath string) (*Commit, error) {
	// File name starts with ':' must be escaped.
	if relpath[0] == ':' {
		relpath = `\` + relpath
	}

	stdout, err := NewCommand("log", "-1", prettyLogFormat, id.String(), "--", relpath).RunInDir(r.path)
	if err != nil {
		return nil, err
	}

	id, err = NewIDFromString(string(stdout))
	if err != nil {
		return nil, err
	}

	return r.getCommit(id)
}

// GetCommitByPath returns the last commit of relative path.
func (r *Repository) GetCommitByPath(relpath string) (*Commit, error) {
	stdout, err := NewCommand("log", "-1", prettyLogFormat, "--", relpath).RunInDir(r.path)
	if err != nil {
		return nil, err
	}

	commits, err := r.parsePrettyFormatLogToList(stdout)
	if err != nil {
		return nil, err
	}
	return commits.Front().Value.(*Commit), nil
}

func (r *Repository) CommitsByRangeSize(revision string, page, size int) (*list.List, error) {
	stdout, err := NewCommand("log", revision, "--skip="+strconv.Itoa((page-1)*size),
		"--max-count="+strconv.Itoa(size), prettyLogFormat).RunInDir(r.path)
	if err != nil {
		return nil, err
	}
	return r.parsePrettyFormatLogToList(stdout)
}

var DefaultCommitsPageSize = 30

func (r *Repository) CommitsByRange(revision string, page int) (*list.List, error) {
	return r.CommitsByRangeSize(revision, page, DefaultCommitsPageSize)
}

func (r *Repository) searchCommits(id SHA1, keyword string) (*list.List, error) {
	stdout, err := NewCommand("log", id.String(), "-100", "-i", "--grep="+keyword, prettyLogFormat).RunInDir(r.path)
	if err != nil {
		return nil, err
	}
	return r.parsePrettyFormatLogToList(stdout)
}

func (r *Repository) getFilesChanged(id1 string, id2 string) ([]string, error) {
	stdout, err := NewCommand("diff", "--name-only", id1, id2).RunInDir(r.path)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(stdout), "\n"), nil
}

func commitsCount(repoPath, revision, relpath string) (int64, error) {
	cmd := NewCommand("rev-list", "--count").AddArgs(revision)
	if len(relpath) > 0 {
		cmd.AddArgs("--", relpath)
	}

	stdout, err := cmd.RunInDir(repoPath)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(strings.TrimSpace(string(stdout)), 10, 64)
}

// CommitsCount returns number of total commits up to given revision.
func CommitsCount(repoPath, revision string) (int64, error) {
	return commitsCount(repoPath, revision, "")
}

// CommitsCount returns number of total commits up to given revision of the repository.
func (r *Repository) CommitsCount(revision string) (int64, error) {
	return CommitsCount(r.path, revision)
}

func (r *Repository) FileCommitsCount(revision, file string) (int64, error) {
	return commitsCount(r.path, revision, file)
}

func (r *Repository) CommitsByFileAndRangeSize(revision, file string, page, size int) (*list.List, error) {
	stdout, err := NewCommand("log", revision, "--skip="+strconv.Itoa((page-1)*size),
		"--max-count="+strconv.Itoa(size), prettyLogFormat, "--", file).RunInDir(r.path)
	if err != nil {
		return nil, err
	}
	return r.parsePrettyFormatLogToList(stdout)
}

func (r *Repository) CommitsByFileAndRange(revision, file string, page int) (*list.List, error) {
	return r.CommitsByFileAndRangeSize(revision, file, page, DefaultCommitsPageSize)
}

func (r *Repository) FilesCountBetween(startCommitID, endCommitID string) (int, error) {
	stdout, err := NewCommand("diff", "--name-only", startCommitID+"..."+endCommitID).RunInDir(r.path)
	if err != nil {
		return 0, err
	}
	return len(strings.Split(string(stdout), "\n")) - 1, nil
}

// CommitsBetween returns a list that contains commits between [last, before).
func (r *Repository) CommitsBetween(last *Commit, before *Commit) (*list.List, error) {
	stdout, err := NewCommand("rev-list", before.id.String()+"..."+last.id.String()).RunInDir(r.path)
	if err != nil {
		return nil, err
	}
	return r.parsePrettyFormatLogToList(bytes.TrimSpace(stdout))
}

func (r *Repository) CommitsBetweenIDs(last, before string) (*list.List, error) {
	lastCommit, err := r.CommitByID(last)
	if err != nil {
		return nil, err
	}
	beforeCommit, err := r.CommitByID(before)
	if err != nil {
		return nil, err
	}
	return r.CommitsBetween(lastCommit, beforeCommit)
}

func (r *Repository) CommitsCountBetween(start, end string) (int64, error) {
	return commitsCount(r.path, start+"..."+end, "")
}

// The limit is depth, not total number of returned commits.
func (r *Repository) commitsBefore(l *list.List, parent *list.Element, id SHA1, current, limit int) error {
	// Reach the limit
	if limit > 0 && current > limit {
		return nil
	}

	commit, err := r.getCommit(id)
	if err != nil {
		return fmt.Errorf("getCommit: %v", err)
	}

	var e *list.Element
	if parent == nil {
		e = l.PushBack(commit)
	} else {
		var in = parent
		for {
			if in == nil {
				break
			} else if in.Value.(*Commit).id.Equal(commit.id) {
				return nil
			} else if in.Next() == nil {
				break
			}

			if in.Value.(*Commit).committer.When.Equal(commit.committer.When) {
				break
			}

			if in.Value.(*Commit).committer.When.After(commit.committer.When) &&
				in.Next().Value.(*Commit).committer.When.Before(commit.committer.When) {
				break
			}

			in = in.Next()
		}

		e = l.InsertAfter(commit, in)
	}

	pr := parent
	if commit.ParentsCount() > 1 {
		pr = e
	}

	for i := 0; i < commit.ParentsCount(); i++ {
		id, err := commit.ParentID(i)
		if err != nil {
			return err
		}
		err = r.commitsBefore(l, pr, id, current+1, limit)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) getCommitsBefore(id SHA1) (*list.List, error) {
	l := list.New()
	return l, r.commitsBefore(l, nil, id, 1, 0)
}

func (r *Repository) getCommitsBeforeLimit(id SHA1, num int) (*list.List, error) {
	l := list.New()
	return l, r.commitsBefore(l, nil, id, 1, num)
}

// CommitsAfterDate returns a list of commits which committed after given date.
// The format of date should be in RFC3339.
func (r *Repository) CommitsAfterDate(date string) (*list.List, error) {
	stdout, err := NewCommand("log", prettyLogFormat, "--since="+date).RunInDir(r.path)
	if err != nil {
		return nil, err
	}

	return r.parsePrettyFormatLogToList(stdout)
}

// GetLatestCommitDate returns the date of latest commit of repository.
// If branch is empty, it returns the latest commit across all branches.
func GetLatestCommitDate(repoPath, branch string) (time.Time, error) {
	cmd := NewCommand("for-each-ref", "--count=1", "--sort=-committerdate", "--format=%(committerdate:iso8601)")
	if len(branch) > 0 {
		cmd.AddArgs("refs/heads/" + branch)
	}
	stdout, err := cmd.RunInDir(repoPath)
	if err != nil {
		return time.Time{}, err
	}

	return time.Parse("2006-01-02 15:04:05 -0700", strings.TrimSpace(string(stdout)))
}
