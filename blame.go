// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"
	"time"
)

type (
	Blame struct {
		commits map[int]*Commit
	}
	BlameOptions struct {
	}
)

// BlameFile returns map of line number and the Commit changed that line.
func (r *Repository) BlameFile(rev, file string, opts ...BlameOptions) (*Blame, error) {
	cmd := NewCommand("blame", "-p", rev, "--", file)
	stdout, err := cmd.RunInDir(r.path)
	if err != nil {
		return nil, err
	}
	return BlameContent(stdout)
}

// BlameContent parse content of `git blame` in porcelain format
func BlameContent(content []byte) (*Blame, error) {
	var commits = make(map[[20]byte]*Commit)
	var commit = &Commit{}
	var details = make(map[string]string)
	var result = createBlame()
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if string(line[0]) != "\t" {
			words := strings.Fields(line)
			sha, err := NewIDFromString(words[0])
			if err == nil {
				// SHA and rows numbers line
				commit = getCommit(sha, commits)
				commit.fill(details)
				details = make(map[string]string) // empty all details
				i, err := strconv.Atoi(words[2])
				if err != nil {
					return nil, err
				}
				result.commits[i] = commit
			} else {
				// commit details line
				switch words[0] {
				case "summary":
					commit.Message = line[len(words[0])+1:]
				case "previous":
					commit.parents = []*SHA1{MustIDFromString(words[1])}
				default:
					if len(words) > 1 {
						details[words[0]] = line[len(words[0])+1:]
					}
				}
			}
		} else {
			// needed for last line in blame
			commit.fill(details)
		}
	}

	return result, nil
}

func createBlame() *Blame {
	var blame = Blame{}
	blame.commits = make(map[int]*Commit)
	return &blame
}

// Return commit from map or creates a new one
func getCommit(sha *SHA1, commits map[[20]byte]*Commit) *Commit {
	commit, ok := commits[sha.bytes]
	if !ok {
		commit = &Commit{
			ID: sha,
		}
		commits[sha.bytes] = commit
	}

	return commit
}

func (c *Commit) fill(data map[string]string) {
	author, ok := data["author"]
	if ok && c.Author == nil {
		t, err := parseBlameTime(data, "author")
		if err != nil {
			c.Author = &Signature{
				Name:  author,
				Email: data["author-mail"],
			}
		} else {
			c.Author = &Signature{
				Name:  author,
				Email: data["author-mail"],
				When:  t,
			}
		}
	}
	committer, ok := data["committer"]
	if ok && c.Committer == nil {
		t, err := parseBlameTime(data, "committer")
		if err != nil {
			c.Committer = &Signature{
				Name:  committer,
				Email: data["committer-mail"],
			}
		} else {
			c.Committer = &Signature{
				Name:  committer,
				Email: data["committer-mail"],
				When:  t,
			}
		}
	}
}

func parseBlameTime(data map[string]string, prefix string) (time.Time, error) {
	atoi, err := strconv.ParseInt(data[prefix+"-time"], 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	t := time.Unix(atoi, 0)

	if len(data["author-tz"]) == 5 {
		hours, ok1 := strconv.ParseInt(data[prefix+"-tz"][:3], 10, 0)
		mins, ok2 := strconv.ParseInt(data[prefix+"-tz"][3:5], 10, 0)
		if ok1 == nil && ok2 == nil {
			t = t.In(time.FixedZone("Fixed", int((hours*60+mins)*60)))
		}
	}
	return t, nil
}
