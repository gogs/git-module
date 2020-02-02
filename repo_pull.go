// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// PullRequestInfo represents needed information for a pull request.
type PullRequestInfo struct {
	MergeBase string
	Commits   []*Commit
	NumFiles  int
}

// GetMergeBase checks and returns merge base of two branches.
func (r *Repository) GetMergeBase(base, head string) (string, error) {
	stdout, err := NewCommand("merge-base", base, head).RunInDir(r.path)
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return "", ErrNoMergeBase{}
		}
		return "", err
	}
	return strings.TrimSpace(string(stdout)), nil
}

type PullRequestInfoOptions struct {
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// GetPullRequestInfo generates and returns pull request information
// between base and head branches of repositories.
func (r *Repository) GetPullRequestInfo(basePath, baseBranch, headBranch string, opts ...PullRequestInfoOptions) (_ *PullRequestInfo, err error) {
	var opt PullRequestInfoOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	var remoteBranch string

	// We don't need a temporary remote for same repository.
	if r.path != basePath {
		// Add a temporary remote
		tmpRemote := strconv.FormatInt(time.Now().UnixNano(), 10)
		if err = r.AddRemote(tmpRemote, basePath, true); err != nil {
			return nil, fmt.Errorf("AddRemote: %v", err)
		}
		defer func() {
			_ = r.RemoveRemote(tmpRemote)
		}()

		remoteBranch = "remotes/" + tmpRemote + "/" + baseBranch
	} else {
		remoteBranch = baseBranch
	}

	prInfo := new(PullRequestInfo)
	prInfo.MergeBase, err = r.GetMergeBase(remoteBranch, headBranch)
	if err != nil {
		return nil, err
	}

	logs, err := NewCommand("log", prInfo.MergeBase+"..."+headBranch, LogFormatHashOnly).RunInDir(r.path)
	if err != nil {
		return nil, err
	}
	prInfo.Commits, err = r.parsePrettyFormatLogToList(opt.Timeout, logs)
	if err != nil {
		return nil, fmt.Errorf("parsePrettyFormatLogToList: %v", err)
	}

	// Count number of changed files.
	stdout, err := NewCommand("diff", "--name-only", remoteBranch+"..."+headBranch).RunInDir(r.path)
	if err != nil {
		return nil, err
	}
	prInfo.NumFiles = bytes.Count(stdout, []byte("\n"))

	return prInfo, nil
}

// GetPatch generates and returns patch data between given revisions.
func (r *Repository) GetPatch(base, head string) ([]byte, error) {
	return NewCommand("diff", "-p", "--binary", base, head).RunInDir(r.path)
}
