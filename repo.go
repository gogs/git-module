// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Repository contains information of a Git repository.
type Repository struct {
	path string

	cachedCommits *objectCache
	cachedTags    *objectCache
	cachedTrees   *objectCache
}

// Path returns the path of the repository.
func (r *Repository) Path() string {
	return r.path
}

const LogFormatHashOnly = `format:%H`

// parsePrettyFormatLogToList returns a list of commits parsed from given logs
// that are formatted in LogFormatHashOnly.
func (r *Repository) parsePrettyFormatLogToList(timeout time.Duration, logs []byte) ([]*Commit, error) {
	if len(logs) == 0 {
		return []*Commit{}, nil
	}

	var err error
	ids := bytes.Split(logs, []byte{'\n'})
	commits := make([]*Commit, len(ids))
	for i, id := range ids {
		commits[i], err = r.CatFileCommit(string(id), CatFileCommitOptions{Timeout: timeout})
		if err != nil {
			return nil, err
		}
	}
	return commits, nil
}

// InitOptions contains optional arguments for initializing a repository.
//
// Docs: https://git-scm.com/docs/git-init
type InitOptions struct {
	// Indicates whether the repository should be initialized in bare format.
	Bare bool
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// Init initializes a new Git repository.
func Init(path string, opts ...InitOptions) error {
	var opt InitOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

	cmd := NewCommand("init").AddOptions(opt.CommandOptions)
	if opt.Bare {
		cmd.AddArgs("--bare")
	}
	_, err = cmd.RunInDirWithTimeout(opt.Timeout, path)
	return err
}

// Open opens the repository at the given path. It returns an os.ErrNotExist if
// the path does not exist.
func Open(repoPath string) (*Repository, error) {
	repoPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, err
	} else if !isDir(repoPath) {
		return nil, os.ErrNotExist
	}

	return &Repository{
		path:          repoPath,
		cachedCommits: newObjectCache(),
		cachedTags:    newObjectCache(),
		cachedTrees:   newObjectCache(),
	}, nil
}

// CloneOptions contains optional arguments for cloning a repository.
//
// Docs: https://git-scm.com/docs/git-clone
type CloneOptions struct {
	// Indicates whether the repository should be cloned as a mirror.
	Mirror bool
	// Indicates whether the repository should be cloned in bare format.
	Bare bool
	// Indicates whether to suppress the log output.
	Quiet bool
	// The branch to checkout for the working tree when Bare=false.
	Branch string
	// The number of revisions to clone.
	Depth uint64
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// Clone clones the repository from remote URL to the destination.
func Clone(url, dst string, opts ...CloneOptions) error {
	var opt CloneOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	err := os.MkdirAll(path.Dir(dst), os.ModePerm)
	if err != nil {
		return err
	}

	cmd := NewCommand("clone").AddOptions(opt.CommandOptions)
	if opt.Mirror {
		cmd.AddArgs("--mirror")
	}
	if opt.Bare {
		cmd.AddArgs("--bare")
	}
	if opt.Quiet {
		cmd.AddArgs("--quiet")
	}
	if !opt.Bare && opt.Branch != "" {
		cmd.AddArgs("-b", opt.Branch)
	}
	if opt.Depth > 0 {
		cmd.AddArgs("--depth", strconv.FormatUint(opt.Depth, 10))
	}

	_, err = cmd.AddArgs(url, dst).RunWithTimeout(opt.Timeout)
	return err
}

// FetchOptions contains optional arguments for fetching repository updates.
//
// Docs: https://git-scm.com/docs/git-fetch
type FetchOptions struct {
	// Indicates whether to prune during fetching.
	Prune bool
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// Fetch fetches updates for the repository.
func (r *Repository) Fetch(opts ...FetchOptions) error {
	var opt FetchOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("fetch").AddOptions(opt.CommandOptions)
	if opt.Prune {
		cmd.AddArgs("--prune")
	}

	_, err := cmd.RunInDirWithTimeout(opt.Timeout, r.path)
	return err
}

// PullOptions contains optional arguments for pulling repository updates.
//
// Docs: https://git-scm.com/docs/git-pull
type PullOptions struct {
	// Indicates whether to rebased during pulling.
	Rebase bool
	// Indicates whether to pull from all remotes.
	All bool
	// The remote to pull updates from when All=false.
	Remote string
	// The branch to pull updates from when All=false and Remote is supplied.
	Branch string
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// Pull pulls updates for the repository.
func (r *Repository) Pull(opts ...PullOptions) error {
	var opt PullOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("pull").AddOptions(opt.CommandOptions)
	if opt.Rebase {
		cmd.AddArgs("--rebase")
	}
	if opt.All {
		cmd.AddArgs("--all")
	}
	if !opt.All && opt.Remote != "" {
		cmd.AddArgs(opt.Remote)
		if opt.Branch != "" {
			cmd.AddArgs(opt.Branch)
		}
	}

	_, err := cmd.RunInDirWithTimeout(opt.Timeout, r.path)
	return err
}

// PushOptions contains optional arguments for pushing repository changes.
//
// Docs: https://git-scm.com/docs/git-push
type PushOptions struct {
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// Push pushes local changes to given remote and branch for the repository in
// given path.
func Push(repoPath, remote, branch string, opts ...PushOptions) error {
	var opt PushOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("push").AddOptions(opt.CommandOptions).AddArgs(remote, branch)
	_, err := cmd.RunInDirWithTimeout(opt.Timeout, repoPath)
	return err
}

// Deprecated: Use Push instead.
func RepoPush(repoPath, remote, branch string, opts ...PushOptions) error {
	return Push(repoPath, remote, branch, opts...)
}

// Push pushes local changes to given remote and branch for the repository.
func (r *Repository) Push(remote, branch string, opts ...PushOptions) error {
	return Push(r.path, remote, branch, opts...)
}

// CheckoutOptions contains optional arguments for checking out to a branch.
//
// Docs: https://git-scm.com/docs/git-checkout
type CheckoutOptions struct {
	// The base branch if checks out to a new branch.
	BaseBranch string
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// Checkout checks out to given branch for the repository in given path.
func Checkout(repoPath, branch string, opts ...CheckoutOptions) error {
	var opt CheckoutOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("checkout").AddOptions(opt.CommandOptions)
	if opt.BaseBranch != "" {
		cmd.AddArgs("-b")
	}
	cmd.AddArgs(branch)
	if opt.BaseBranch != "" {
		cmd.AddArgs(opt.BaseBranch)
	}

	_, err := cmd.RunInDirWithTimeout(opt.Timeout, repoPath)
	return err
}

// Deprecated: Use Checkout instead.
func RepoCheckout(repoPath, branch string, opts ...CheckoutOptions) error {
	return Checkout(repoPath, branch, opts...)
}

// Checkout checks out to given branch for the repository.
func (r *Repository) Checkout(branch string, opts ...CheckoutOptions) error {
	return Checkout(r.path, branch, opts...)
}

// ResetOptions contains optional arguments for resetting a branch.
//
// Docs: https://git-scm.com/docs/git-reset
type ResetOptions struct {
	// Indicates whether to perform a hard reset.
	Hard bool
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// Reset resets working tree to given revision for the repository in given path.
func Reset(repoPath, rev string, opts ...ResetOptions) error {
	var opt ResetOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("reset")
	if opt.Hard {
		cmd.AddArgs("--hard")
	}

	_, err := cmd.AddOptions(opt.CommandOptions).AddArgs(rev).RunInDir(repoPath)
	return err
}

// Deprecated: Use Reset instead.
func RepoReset(repoPath, rev string, opts ...ResetOptions) error {
	return Reset(repoPath, rev, opts...)
}

// Reset resets working tree to given revision for the repository.
func (r *Repository) Reset(rev string, opts ...ResetOptions) error {
	return Reset(r.path, rev, opts...)
}

// MoveOptions contains optional arguments for moving a file, a directory, or a
// symlink.
//
// Docs: https://git-scm.com/docs/git-mv
type MoveOptions struct {
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// Move moves a file, a directory, or a symlink file or directory from source to
// destination for the repository in given path.
func Move(repoPath, src, dst string, opts ...MoveOptions) error {
	var opt MoveOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	_, err := NewCommand("mv").AddOptions(opt.CommandOptions).AddArgs(src, dst).RunInDirWithTimeout(opt.Timeout, repoPath)
	return err
}

// Deprecated: Use Move instead.
func RepoMove(repoPath, src, dst string, opts ...MoveOptions) error {
	return Move(repoPath, src, dst, opts...)
}

// Move moves a file, a directory, or a symlink file or directory from source to
// destination for the repository.
func (r *Repository) Move(src, dst string, opts ...MoveOptions) error {
	return Move(r.path, src, dst, opts...)
}

// AddOptions contains optional arguments for adding local changes.
//
// Docs: https://git-scm.com/docs/git-add
type AddOptions struct {
	// Indicates whether to add all changes to index.
	All bool
	// The specific pathspecs to be added to index.
	Pathspecs []string
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// Add adds local changes to index for the repository in given path.
func Add(repoPath string, opts ...AddOptions) error {
	var opt AddOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("add").AddOptions(opt.CommandOptions)
	if opt.All {
		cmd.AddArgs("--all")
	}
	if len(opt.Pathspecs) > 0 {
		cmd.AddArgs("--")
		cmd.AddArgs(opt.Pathspecs...)
	}
	_, err := cmd.RunInDirWithTimeout(opt.Timeout, repoPath)
	return err
}

// Deprecated: Use Add instead.
func RepoAdd(repoPath string, opts ...AddOptions) error {
	return Add(repoPath, opts...)
}

// Add adds local changes to index for the repository.
func (r *Repository) Add(opts ...AddOptions) error {
	return Add(r.path, opts...)
}

// CommitOptions contains optional arguments to commit changes.
//
// Docs: https://git-scm.com/docs/git-commit
type CommitOptions struct {
	// Author is the author of the changes if that's not the same as committer.
	Author *Signature
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// CreateCommit commits local changes with given author, committer and message
// for the repository in given path.
func CreateCommit(repoPath string, committer *Signature, message string, opts ...CommitOptions) error {
	var opt CommitOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("commit")
	cmd.AddCommitter(committer)

	if opt.Author == nil {
		opt.Author = committer
	}
	cmd = cmd.AddArgs(fmt.Sprintf("--author='%s <%s>'", opt.Author.Name, opt.Author.Email)).
		AddArgs("-m", message).
		AddOptions(opt.CommandOptions)

	_, err := cmd.RunInDirWithTimeout(opt.Timeout, repoPath)
	// No stderr but exit status 1 means nothing to commit.
	if err != nil && err.Error() == "exit status 1" {
		return nil
	}
	return err
}

// Deprecated: Use CreateCommit instead.
func RepoCommit(repoPath string, committer *Signature, message string, opts ...CommitOptions) error {
	return CreateCommit(repoPath, committer, message, opts...)
}

// Commit commits local changes with given author, committer and message for the
// repository.
func (r *Repository) Commit(committer *Signature, message string, opts ...CommitOptions) error {
	return CreateCommit(r.path, committer, message, opts...)
}

// NameStatus contains name status of a commit.
type NameStatus struct {
	Added    []string
	Removed  []string
	Modified []string
}

// ShowNameStatusOptions contains optional arguments for showing name status.
//
// Docs: https://git-scm.com/docs/git-show#Documentation/git-show.txt---name-status
type ShowNameStatusOptions struct {
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// ShowNameStatus returns name status of given revision of the repository in
// given path.
func ShowNameStatus(repoPath, rev string, opts ...ShowNameStatusOptions) (*NameStatus, error) {
	var opt ShowNameStatusOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	fileStatus := &NameStatus{}
	stdout, w := io.Pipe()
	done := make(chan struct{})
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) < 2 {
				continue
			}

			switch fields[0][0] {
			case 'A':
				fileStatus.Added = append(fileStatus.Added, fields[1])
			case 'D':
				fileStatus.Removed = append(fileStatus.Removed, fields[1])
			case 'M':
				fileStatus.Modified = append(fileStatus.Modified, fields[1])
			}
		}
		done <- struct{}{}
	}()

	stderr := new(bytes.Buffer)
	cmd := NewCommand("show", "--name-status", "--pretty=format:''").
		AddOptions(opt.CommandOptions).
		AddArgs(rev)
	err := cmd.RunInDirPipelineWithTimeout(opt.Timeout, w, stderr, repoPath)
	_ = w.Close() // Close writer to exit parsing goroutine
	if err != nil {
		return nil, concatenateError(err, stderr.String())
	}

	<-done
	return fileStatus, nil
}

// Deprecated: Use ShowNameStatus instead.
func RepoShowNameStatus(repoPath, rev string, opts ...ShowNameStatusOptions) (*NameStatus, error) {
	return ShowNameStatus(repoPath, rev, opts...)
}

// ShowNameStatus returns name status of given revision of the repository.
func (r *Repository) ShowNameStatus(rev string, opts ...ShowNameStatusOptions) (*NameStatus, error) {
	return ShowNameStatus(r.path, rev, opts...)
}

// RevParseOptions contains optional arguments for parsing revision.
//
// Docs: https://git-scm.com/docs/git-rev-parse
type RevParseOptions struct {
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// RevParse returns full length (40) commit ID by given revision in the
// repository.
func (r *Repository) RevParse(rev string, opts ...RevParseOptions) (string, error) {
	var opt RevParseOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	commitID, err := NewCommand("rev-parse").
		AddOptions(opt.CommandOptions).
		AddArgs(rev).
		RunInDirWithTimeout(opt.Timeout, r.path)
	if err != nil {
		if strings.Contains(err.Error(), "exit status 128") {
			return "", ErrRevisionNotExist
		}
		return "", err
	}
	return strings.TrimSpace(string(commitID)), nil
}

// CountObject contains disk usage report of a repository.
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

// CountObjectsOptions contains optional arguments for counting objects.
//
// Docs: https://git-scm.com/docs/git-count-objects
type CountObjectsOptions struct {
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// CountObjects returns disk usage report of the repository in given path.
func CountObjects(repoPath string, opts ...CountObjectsOptions) (*CountObject, error) {
	var opt CountObjectsOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	stdout, err := NewCommand("count-objects", "-v").
		AddOptions(opt.CommandOptions).
		RunInDirWithTimeout(opt.Timeout, repoPath)
	if err != nil {
		return nil, err
	}

	toInt64 := func(b []byte) int64 {
		i, _ := strconv.ParseInt(string(b), 10, 64)
		return i
	}

	countObject := new(CountObject)
	for _, line := range bytes.Split(stdout, []byte("\n")) {
		switch {
		case bytes.HasPrefix(line, []byte("count: ")):
			countObject.Count = toInt64(line[7:])
		case bytes.HasPrefix(line, []byte("size: ")):
			countObject.Size = toInt64(line[6:]) * 1024
		case bytes.HasPrefix(line, []byte("in-pack: ")):
			countObject.InPack = toInt64(line[9:])
		case bytes.HasPrefix(line, []byte("packs: ")):
			countObject.Packs = toInt64(line[7:])
		case bytes.HasPrefix(line, []byte("size-pack: ")):
			countObject.SizePack = toInt64(line[11:]) * 1024
		case bytes.HasPrefix(line, []byte("prune-packable: ")):
			countObject.PrunePackable = toInt64(line[16:])
		case bytes.HasPrefix(line, []byte("garbage: ")):
			countObject.Garbage = toInt64(line[9:])
		case bytes.HasPrefix(line, []byte("size-garbage: ")):
			countObject.SizeGarbage = toInt64(line[14:]) * 1024
		}
	}

	return countObject, nil
}

// Deprecated: Use CountObjects instead.
func RepoCountObjects(repoPath string, opts ...CountObjectsOptions) (*CountObject, error) {
	return CountObjects(repoPath, opts...)
}

// CountObjects returns disk usage report of the repository.
func (r *Repository) CountObjects(opts ...CountObjectsOptions) (*CountObject, error) {
	return CountObjects(r.path, opts...)
}

// FsckOptions contains optional arguments for verifying the objects.
//
// Docs: https://git-scm.com/docs/git-fsck
type FsckOptions struct {
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// Fsck verifies the connectivity and validity of the objects in the database
// for the repository in given path.
func Fsck(repoPath string, opts ...FsckOptions) error {
	var opt FsckOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("fsck").AddOptions(opt.CommandOptions)
	_, err := cmd.RunInDirWithTimeout(opt.Timeout, repoPath)
	return err
}

// Deprecated: Use Fsck instead.
func RepoFsck(repoPath string, opts ...FsckOptions) error {
	return Fsck(repoPath, opts...)
}

// Fsck verifies the connectivity and validity of the objects in the database
// for the repository.
func (r *Repository) Fsck(opts ...FsckOptions) error {
	return Fsck(r.path, opts...)
}
