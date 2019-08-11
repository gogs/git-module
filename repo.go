// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"container/list"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Repository represents a Git repository.
type Repository struct {
	Path string

	commitCache *objectCache
	tagCache    *objectCache
}

const prettyLogFormat = `--pretty=format:%H`

func (repo *Repository) parsePrettyFormatLogToList(logs []byte) (*list.List, error) {
	l := list.New()
	if len(logs) == 0 {
		return l, nil
	}

	parts := bytes.Split(logs, []byte{'\n'})

	for _, commitId := range parts {
		commit, err := repo.GetCommit(string(commitId))
		if err != nil {
			return nil, err
		}
		l.PushBack(commit)
	}

	return l, nil
}

type NetworkOptions struct {
	URL     string
	Timeout time.Duration
}

// IsRepoURLAccessible checks if given repository URL is accessible.
func IsRepoURLAccessible(opts NetworkOptions) bool {
	cmd := NewCommand("ls-remote", "-q", "-h", opts.URL, "HEAD")
	if opts.Timeout <= 0 {
		opts.Timeout = -1
	}
	_, err := cmd.RunTimeout(opts.Timeout)
	if err != nil {
		return false
	}
	return true
}

// InitRepository initializes a new Git repository.
func InitRepository(repoPath string, bare bool) error {
	os.MkdirAll(repoPath, os.ModePerm)

	cmd := NewCommand("init")
	if bare {
		cmd.AddArguments("--bare")
	}
	_, err := cmd.RunInDir(repoPath)
	return err
}

// OpenRepository opens the repository at the given path.
func OpenRepository(repoPath string) (*Repository, error) {
	repoPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, err
	} else if !isDir(repoPath) {
		return nil, errors.New("no such file or directory")
	}

	return &Repository{
		Path:        repoPath,
		commitCache: newObjectCache(),
		tagCache:    newObjectCache(),
	}, nil
}

type CloneRepoOptions struct {
	Mirror  bool
	Bare    bool
	Quiet   bool
	Branch  string
	Timeout time.Duration
}

// Clone clones original repository to target path.
func Clone(from, to string, opts CloneRepoOptions) (err error) {
	toDir := path.Dir(to)
	if err = os.MkdirAll(toDir, os.ModePerm); err != nil {
		return err
	}

	cmd := NewCommand("clone")
	if opts.Mirror {
		cmd.AddArguments("--mirror")
	}
	if opts.Bare {
		cmd.AddArguments("--bare")
	}
	if opts.Quiet {
		cmd.AddArguments("--quiet")
	}
	if len(opts.Branch) > 0 {
		cmd.AddArguments("-b", opts.Branch)
	}
	cmd.AddArguments(from, to)

	if opts.Timeout <= 0 {
		opts.Timeout = -1
	}
	_, err = cmd.RunTimeout(opts.Timeout)
	return err
}

type FetchRemoteOptions struct {
	Prune   bool
	Timeout time.Duration
}

// Fetch fetches changes from remotes without merging.
func Fetch(repoPath string, opts FetchRemoteOptions) error {
	cmd := NewCommand("fetch")
	if opts.Prune {
		cmd.AddArguments("--prune")
	}

	if opts.Timeout <= 0 {
		opts.Timeout = -1
	}
	_, err := cmd.RunInDirTimeout(opts.Timeout, repoPath)
	return err
}

type PullRemoteOptions struct {
	All     bool
	Rebase  bool
	Remote  string
	Branch  string
	Timeout time.Duration
}

// Pull pulls changes from remotes.
func Pull(repoPath string, opts PullRemoteOptions) error {
	cmd := NewCommand("pull")
	if opts.Rebase {
		cmd.AddArguments("--rebase")
	}
	if opts.All {
		cmd.AddArguments("--all")
	} else {
		cmd.AddArguments(opts.Remote)
		cmd.AddArguments(opts.Branch)
	}

	if opts.Timeout <= 0 {
		opts.Timeout = -1
	}
	_, err := cmd.RunInDirTimeout(opts.Timeout, repoPath)
	return err
}

// PushWithEnvs pushs local commits to given remote branch with given environment variables.
func PushWithEnvs(repoPath, remote, branch string, envs []string) error {
	_, err := NewCommand("push", remote, branch).AddEnvs(envs...).RunInDir(repoPath)
	return err
}

// Push pushs local commits to given remote branch.
func Push(repoPath, remote, branch string) error {
	return PushWithEnvs(repoPath, remote, branch, nil)
}

type CheckoutOptions struct {
	Branch    string
	OldBranch string
	Timeout   time.Duration
}

// Checkout checkouts a branch
func Checkout(repoPath string, opts CheckoutOptions) error {
	cmd := NewCommand("checkout")
	if len(opts.OldBranch) > 0 {
		cmd.AddArguments("-b")
	}

	cmd.AddArguments(opts.Branch)

	if len(opts.OldBranch) > 0 {
		cmd.AddArguments(opts.OldBranch)
	}
	if opts.Timeout <= 0 {
		opts.Timeout = -1
	}
	_, err := cmd.RunInDirTimeout(opts.Timeout, repoPath)
	return err
}

// ResetHEAD resets HEAD to given revision or head of branch.
func ResetHEAD(repoPath string, hard bool, revision string) error {
	cmd := NewCommand("reset")
	if hard {
		cmd.AddArguments("--hard")
	}
	_, err := cmd.AddArguments(revision).RunInDir(repoPath)
	return err
}

// MoveFile moves a file to another file or directory.
func MoveFile(repoPath, oldTreeName, newTreeName string) error {
	_, err := NewCommand("mv").AddArguments(oldTreeName, newTreeName).RunInDir(repoPath)
	return err
}

// CountObject represents disk usage report of Git repository.
type CountObject struct {
	Count         int64
	Size          int64
	InPack        int64
	Packs         int64
	SizePack      int64
	PrunePackable int64
	Garbage       int64
	SizeGarbage   int64
}

const (
	statCount         = "count: "
	statSize          = "size: "
	statInPack        = "in-pack: "
	statPacks         = "packs: "
	statSizePack      = "size-pack: "
	statPrunePackable = "prune-packable: "
	statGarbage       = "garbage: "
	statSizeGarbage   = "size-garbage: "
)

func strToInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

// GetRepoSize returns disk usage report of repository in given path.
func GetRepoSize(repoPath string) (*CountObject, error) {
	cmd := NewCommand("count-objects", "-v")
	stdout, err := cmd.RunInDir(repoPath)
	if err != nil {
		return nil, err
	}

	countObject := new(CountObject)
	for _, line := range strings.Split(stdout, "\n") {
		switch {
		case strings.HasPrefix(line, statCount):
			countObject.Count = strToInt64(line[7:])
		case strings.HasPrefix(line, statSize):
			countObject.Size = strToInt64(line[6:]) * 1024
		case strings.HasPrefix(line, statInPack):
			countObject.InPack = strToInt64(line[9:])
		case strings.HasPrefix(line, statPacks):
			countObject.Packs = strToInt64(line[7:])
		case strings.HasPrefix(line, statSizePack):
			countObject.SizePack = strToInt64(line[11:]) * 1024
		case strings.HasPrefix(line, statPrunePackable):
			countObject.PrunePackable = strToInt64(line[16:])
		case strings.HasPrefix(line, statGarbage):
			countObject.Garbage = strToInt64(line[9:])
		case strings.HasPrefix(line, statSizeGarbage):
			countObject.SizeGarbage = strToInt64(line[14:]) * 1024
		}
	}

	return countObject, nil
}
