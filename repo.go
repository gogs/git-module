// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bufio"
	"bytes"
	"container/list"
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

	commitCache *objectCache
	tagCache    *objectCache
}

// Path returns the path of the repository.
func (r *Repository) Path() string {
	return r.path
}

const prettyLogFormat = `--pretty=format:%H`

// parsePrettyFormatLogToList returns a list of commits parsed from given logs that are
// formatted in prettyLogFormat.
func (r *Repository) parsePrettyFormatLogToList(logs []byte) (*list.List, error) {
	l := list.New()
	if len(logs) == 0 {
		return l, nil
	}

	ids := bytes.Split(logs, []byte{'\n'})
	for _, id := range ids {
		c, err := r.CommitByID(string(id))
		if err != nil {
			return nil, err
		}
		l.PushBack(c)
	}
	return l, nil
}

// NetworkOptions contains arguments for a network operation.
type NetworkOptions struct {
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// IsURLAccessible returns true if given remote URL is accessible via Git.
func IsURLAccessible(url string, opts ...NetworkOptions) bool {
	var opt NetworkOptions
	if len(opts) > 1 {
		opt = opts[0]
	}

	_, err := NewCommand("ls-remote", "-q", "-h", url, "HEAD").RunWithTimeout(opt.Timeout)
	return err == nil
}

// InitOptions contains optional arguments for initializing a repository.
type InitOptions struct {
	// Indicates whether the repository should be initialized in bare format.
	Bare bool
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// Init initializes a new Git repository.
func Init(path string, opts ...InitOptions) error {
	var opt InitOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	_ = os.MkdirAll(path, os.ModePerm)

	cmd := NewCommand("init")
	if opt.Bare {
		cmd.AddArgs("--bare")
	}
	_, err := cmd.RunInDirWithTimeout(opt.Timeout, path)
	return err
}

// Open opens the repository at the given path. It returns an os.ErrNotExist
// if the path does not exist.
func Open(path string) (*Repository, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	} else if !isDir(path) {
		return nil, os.ErrNotExist
	}

	return &Repository{
		path:        path,
		commitCache: newObjectCache(),
		tagCache:    newObjectCache(),
	}, nil
}

// CloneOptions contains optional arguments for cloning a repository.
type CloneOptions struct {
	// Indicates whether the repository should be cloned as a mirror.
	Mirror bool
	// Indicates whether the repository should be cloned in bare format.
	Bare bool
	// Indicates whether to suppress the log output.
	Quiet bool
	// The branch to checkout for the working tree when Bare=false.
	Branch string
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
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

	cmd := NewCommand("clone")
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

	_, err = cmd.AddArgs(url, dst).RunWithTimeout(opt.Timeout)
	return err
}

// FetchOptions contains optional arguments for fetching repository updates.
type FetchOptions struct {
	// Indicates whether to prune during fetching.
	Prune bool
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// Fetch fetches updates for the repository.
func (r *Repository) Fetch(opts ...FetchOptions) error {
	var opt FetchOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("fetch")
	if opt.Prune {
		cmd.AddArgs("--prune")
	}

	_, err := cmd.RunInDirWithTimeout(opt.Timeout, r.path)
	return err
}

// PullOptions contains optional arguments for pulling repository updates.
type PullOptions struct {
	// Indicates whether to rebased during pulling.
	Rebase bool
	// Indicates whether to pull from all remotes.
	All bool
	// The remote to pull updates from when All=false.
	Remote string
	// The branch to pull updates from when All=false and Remote is supplied.
	Branch string
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// Pull pulls updates for the repository.
func (r *Repository) Pull(opts ...PullOptions) error {
	var opt PullOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("pull")
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
type PushOptions struct {
	// The environment variables set for the push.
	Envs []string
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// Push pushs local changes to given remote and branch for the repository.
func (r *Repository) Push(remote, branch string, opts ...PushOptions) error {
	var opt PushOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	_, err := NewCommand("push", remote, branch).AddEnvs(opt.Envs...).RunInDirWithTimeout(opt.Timeout, r.path)
	return err
}

// CheckoutOptions contains optional arguments for checking out to a branch.
type CheckoutOptions struct {
	// The base branch if checks out to a new branch.
	BaseBranch string
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// Checkout checks out to given branch for the repository.
func (r *Repository) Checkout(branch string, opts ...CheckoutOptions) error {
	var opt CheckoutOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("checkout")
	if opt.BaseBranch != "" {
		cmd.AddArgs("-b")
	}
	cmd.AddArgs(branch)
	if opt.BaseBranch != "" {
		cmd.AddArgs(opt.BaseBranch)
	}

	_, err := cmd.RunInDirWithTimeout(opt.Timeout, r.path)
	return err
}

// ResetOptions contains optional arguments for resetting a branch.
type ResetOptions struct {
	// Indicates whether to perform a hard reset.
	Hard bool
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// Reset resets working tree to given revision for the repository.
func (r *Repository) Reset(rev string, opts ...ResetOptions) error {
	var opt ResetOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("reset")
	if opt.Hard {
		cmd.AddArgs("--hard")
	}

	_, err := cmd.AddArgs(rev).RunInDir(r.path)
	return err
}

// MoveOptions contains optional arguments for moving a file, a directory, or a symlink.
type MoveOptions struct {
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// Move moves a file, a directory, or a symlink file or directory from source to destination.
func (r *Repository) Move(src, dst string, opts ...MoveOptions) error {
	var opt MoveOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	_, err := NewCommand("mv", src, dst).RunInDirWithTimeout(opt.Timeout, r.path)
	return err
}

// AddOptions contains optional arguments for adding local changes.
type AddOptions struct {
	// Indicates whether to add all changes to index.
	All bool
	// The specific pathspecs to be added to index.
	Pathsepcs []string
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// Add adds local changes to index for the repository.
func (r *Repository) Add(opts ...AddOptions) error {
	var opt AddOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("add")
	if opt.All {
		cmd.AddArgs("--all")
	}
	if len(opt.Pathsepcs) > 0 {
		cmd.AddArgs("--")
		cmd.AddArgs(opt.Pathsepcs...)
	}
	_, err := cmd.RunInDirWithTimeout(opt.Timeout, r.path)
	return err
}

// CommitOptions contains optional arguments to commit changes.
type CommitOptions struct {
	// Author is the author of the changes if that's not the same as committer.
	Author *Signature
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// Commit commits local changes with given author, committer and message for the repository.
func (r *Repository) Commit(committer *Signature, message string, opts ...CommitOptions) error {
	var opt CommitOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("commit")
	cmd.AddEnvs("GIT_COMMITTER_NAME="+committer.Name, "GIT_COMMITTER_EMAIL="+committer.Email)

	if opt.Author != nil {
		cmd.AddArgs(fmt.Sprintf("--author='%s <%s>'", opt.Author.Name, opt.Author.Email))
	}
	cmd.AddArgs("-m", message)

	_, err := cmd.RunInDirWithTimeout(opt.Timeout, r.path)
	// No stderr but exit status 1 means nothing to commit.
	if err != nil && err.Error() == "exit status 1" {
		return nil
	}
	return err
}

// NameStatus contains name status of a commit.
type NameStatus struct {
	Added    []string
	Removed  []string
	Modified []string
}

// ShowNameStatusOptions contains optional arguments for showing name status.
type ShowNameStatusOptions struct {
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// ShowNameStatus returns name status of given commit of the repository.
func (r *Repository) ShowNameStatus(commitID string, opts ...ShowNameStatusOptions) (*NameStatus, error) {
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
	err := NewCommand("show", "--name-status", "--pretty=format:''", commitID).RunInDirPipelineWithTimeout(opt.Timeout, w, stderr, r.path)
	_ = w.Close() // Close writer to exit parsing goroutine
	if err != nil {
		return nil, concatenateError(err, stderr.String())
	}

	<-done
	return fileStatus, nil
}

// RevParseOptions contains optional arguments for parsing revision.
type RevParseOptions struct {
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// RevParse returns full length (40) commit ID by given revision in the repository.
func (r *Repository) RevParse(rev string, opts ...RevParseOptions) (string, error) {
	var opt RevParseOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	commitID, err := NewCommand("rev-parse", rev).RunInDirWithTimeout(opt.Timeout, r.path)
	if err != nil {
		if strings.Contains(err.Error(), "exit status 128") {
			return "", ErrNotExist{rev, ""}
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
type CountObjectsOptions struct {
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// CountObjects returns disk usage report of the repository.
func (r *Repository) CountObjects(opts ...CountObjectsOptions) (*CountObject, error) {
	var opt CountObjectsOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	stdout, err := NewCommand("count-objects", "-v").RunInDirWithTimeout(opt.Timeout, r.path)
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

// FsckOptions contains optional arguments for verifying the objects.
type FsckOptions struct {
	// The timeout duration before giving up. The default timeout duration will be used when not supplied.
	Timeout time.Duration
}

// Fsck verifies the connectivity and validity of the objects in the database for the repository.
func (r *Repository) Fsck(opts ...FsckOptions) error {
	var opt FsckOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	_, err := NewCommand("fsck").RunInDirWithTimeout(opt.Timeout, r.path)
	return err
}
