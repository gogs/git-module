package git

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"
	"time"
)

// BlameFile return map of line number and Commit that change that line
func (r *Repository) BlameFile(file string) (map[int]*Commit, error) {
	cmd := NewCommand("blame", "-p", file)
	stdout, err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return Blame(stdout)
}

// Blame parse content of `git blame` in porcelain format
func Blame(content []byte) (map[int]*Commit, error) {
	var commits = make(map[[20]byte]*Commit)
	var commit = &Commit{}
	var details = make(map[string]string)
	var result = make(map[int]*Commit)
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
				result[i] = commit
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
