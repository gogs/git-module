// Copyright 2019 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"strings"
	"time"
)

// LsRemoteOptions contains arguments for listing references in a remote
// repository.
//
// Docs: https://git-scm.com/docs/git-ls-remote
type LsRemoteOptions struct {
	// Indicates whether include heads.
	Heads bool
	// Indicates whether include tags.
	Tags bool
	// Indicates whether to not show peeled tags or pseudo refs.
	Refs bool
	// The list of patterns to filter results.
	Patterns []string
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// LsRemote returns a list references in the remote repository.
func LsRemote(url string, opts ...LsRemoteOptions) ([]*Reference, error) {
	var opt LsRemoteOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("ls-remote", "--quiet").AddOptions(opt.CommandOptions)
	if opt.Heads {
		cmd.AddArgs("--heads")
	}
	if opt.Tags {
		cmd.AddArgs("--tags")
	}
	if opt.Refs {
		cmd.AddArgs("--refs")
	}
	cmd.AddArgs(url)
	if len(opt.Patterns) > 0 {
		cmd.AddArgs(opt.Patterns...)
	}

	stdout, err := cmd.RunWithTimeout(opt.Timeout)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(stdout, []byte("\n"))
	refs := make([]*Reference, 0, len(lines))
	for i := range lines {
		fields := bytes.Fields(lines[i])
		if len(fields) < 2 {
			continue
		}

		refs = append(refs, &Reference{
			ID:      string(fields[0]),
			Refspec: string(fields[1]),
		})
	}
	return refs, nil
}

// IsURLAccessible returns true if given remote URL is accessible via Git within
// given timeout.
func IsURLAccessible(timeout time.Duration, url string) bool {
	_, err := LsRemote(url, LsRemoteOptions{
		Patterns: []string{"HEAD"},
		Timeout:  timeout,
	})
	return err == nil
}

// RemoteAddOptions contains options to add a remote address.
//
// Docs: https://git-scm.com/docs/git-remote#Documentation/git-remote.txt-emaddem
type RemoteAddOptions struct {
	// Indicates whether to execute git fetch after the remote information is set
	// up.
	Fetch bool
	// Indicates whether to add remote as mirror with --mirror=fetch.
	MirrorFetch bool
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// Deprecated: Use RemoteAddOptions instead.
type AddRemoteOptions = RemoteAddOptions

// RemoteAdd adds a new remote to the repository in given path.
func RemoteAdd(repoPath, name, url string, opts ...RemoteAddOptions) error {
	var opt RemoteAddOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("remote", "add").AddOptions(opt.CommandOptions)
	if opt.Fetch {
		cmd.AddArgs("-f")
	}
	if opt.MirrorFetch {
		cmd.AddArgs("--mirror=fetch")
	}

	_, err := cmd.AddArgs(name, url).RunInDirWithTimeout(opt.Timeout, repoPath)
	return err
}

// Deprecated: Use RemoteAdd instead.
func RepoAddRemote(repoPath, name, url string, opts ...RemoteAddOptions) error {
	return RemoteAdd(repoPath, name, url, opts...)
}

// RemoteAdd adds a new remote to the repository.
func (r *Repository) RemoteAdd(name, url string, opts ...RemoteAddOptions) error {
	return RemoteAdd(r.path, name, url, opts...)
}

// Deprecated: Use RemoteAdd instead.
func (r *Repository) AddRemote(name, url string, opts ...RemoteAddOptions) error {
	return RemoteAdd(r.path, name, url, opts...)
}

// RemoteRemoveOptions contains arguments for removing a remote from the
// repository.
//
// Docs: https://git-scm.com/docs/git-remote#Documentation/git-remote.txt-emremoveem
type RemoteRemoveOptions struct {
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// Deprecated: Use RemoteRemoveOptions instead.
type RemoveRemoteOptions = RemoteRemoveOptions

// RemoteRemove removes a remote from the repository in given path.
func RemoteRemove(repoPath, name string, opts ...RemoteRemoveOptions) error {
	var opt RemoteRemoveOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	_, err := NewCommand("remote", "remove").
		AddOptions(opt.CommandOptions).
		AddArgs(name).
		RunInDirWithTimeout(opt.Timeout, repoPath)
	if err != nil {
		// the error status may differ from git clients
		if strings.Contains(err.Error(), "error: No such remote") ||
			strings.Contains(err.Error(), "fatal: No such remote") {
			return ErrRemoteNotExist
		}
		return err
	}
	return nil
}

// Deprecated: Use RemoteRemove instead.
func RepoRemoveRemote(repoPath, name string, opts ...RemoteRemoveOptions) error {
	return RemoteRemove(repoPath, name, opts...)
}

// RemoteRemove removes a remote from the repository.
func (r *Repository) RemoteRemove(name string, opts ...RemoteRemoveOptions) error {
	return RemoteRemove(r.path, name, opts...)
}

// Deprecated: Use RemoteRemove instead.
func (r *Repository) RemoveRemote(name string, opts ...RemoteRemoveOptions) error {
	return RemoteRemove(r.path, name, opts...)
}

// RemotesOptions contains arguments for listing remotes of the repository.
// /
// Docs: https://git-scm.com/docs/git-remote#_commands
type RemotesOptions struct {
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// Remotes lists remotes of the repository in given path.
func Remotes(repoPath string, opts ...RemotesOptions) ([]string, error) {
	var opt RemotesOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	stdout, err := NewCommand("remote").
		AddOptions(opt.CommandOptions).
		RunInDirWithTimeout(opt.Timeout, repoPath)
	if err != nil {
		return nil, err
	}

	return bytesToStrings(stdout), nil
}

// Remotes lists remotes of the repository.
func (r *Repository) Remotes(opts ...RemotesOptions) ([]string, error) {
	return Remotes(r.path, opts...)
}

// RemoteGetURLOptions contains arguments for retrieving URL(s) of a remote of
// the repository.
//
// Docs: https://git-scm.com/docs/git-remote#Documentation/git-remote.txt-emget-urlem
type RemoteGetURLOptions struct {
	// Indicates whether to get push URLs instead of fetch URLs.
	Push bool
	// Indicates whether to get all URLs, including lists that are not part of main
	// URLs. This option is independent of the Push option.
	All bool
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// RemoteGetURL retrieves URL(s) of a remote of the repository in given path.
func RemoteGetURL(repoPath, name string, opts ...RemoteGetURLOptions) ([]string, error) {
	var opt RemoteGetURLOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("remote", "get-url").AddOptions(opt.CommandOptions)
	if opt.Push {
		cmd.AddArgs("--push")
	}
	if opt.All {
		cmd.AddArgs("--all")
	}

	stdout, err := cmd.AddArgs(name).RunInDirWithTimeout(opt.Timeout, repoPath)
	if err != nil {
		return nil, err
	}
	return bytesToStrings(stdout), nil
}

// RemoteGetURL retrieves URL(s) of a remote of the repository in given path.
func (r *Repository) RemoteGetURL(name string, opts ...RemoteGetURLOptions) ([]string, error) {
	return RemoteGetURL(r.path, name, opts...)
}

// RemoteSetURLOptions contains arguments for setting an URL of a remote of the
// repository.
//
// Docs: https://git-scm.com/docs/git-remote#Documentation/git-remote.txt-emset-urlem
type RemoteSetURLOptions struct {
	// Indicates whether to get push URLs instead of fetch URLs.
	Push bool
	// The regex to match existing URLs to replace (instead of first).
	Regex string
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// RemoteSetURL sets first URL of the remote with given name of the repository
// in given path.
func RemoteSetURL(repoPath, name, newurl string, opts ...RemoteSetURLOptions) error {
	var opt RemoteSetURLOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("remote", "set-url").AddOptions(opt.CommandOptions)
	if opt.Push {
		cmd.AddArgs("--push")
	}

	cmd.AddArgs(name, newurl)

	if opt.Regex != "" {
		cmd.AddArgs(opt.Regex)
	}

	_, err := cmd.RunInDirWithTimeout(opt.Timeout, repoPath)
	if err != nil {
		if strings.Contains(err.Error(), "No such URL found") {
			return ErrURLNotExist
		} else if strings.Contains(err.Error(), "No such remote") {
			return ErrRemoteNotExist
		}
		return err
	}
	return nil
}

// RemoteSetURL sets the first URL of the remote with given name of the
// repository.
func (r *Repository) RemoteSetURL(name, newurl string, opts ...RemoteSetURLOptions) error {
	return RemoteSetURL(r.path, name, newurl, opts...)
}

// RemoteSetURLAddOptions contains arguments for appending an URL to a remote
// of the repository.
//
// Docs: https://git-scm.com/docs/git-remote#Documentation/git-remote.txt-emset-urlem
type RemoteSetURLAddOptions struct {
	// Indicates whether to get push URLs instead of fetch URLs.
	Push bool
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// RemoteSetURLAdd appends an URL to the remote with given name of the
// repository in given path. Use RemoteSetURL to overwrite the URL(s) instead.
func RemoteSetURLAdd(repoPath, name, newurl string, opts ...RemoteSetURLAddOptions) error {
	var opt RemoteSetURLAddOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("remote", "set-url").
		AddOptions(opt.CommandOptions).
		AddArgs("--add")
	if opt.Push {
		cmd.AddArgs("--push")
	}

	cmd.AddArgs(name, newurl)

	_, err := cmd.RunInDirWithTimeout(opt.Timeout, repoPath)
	if err != nil && strings.Contains(err.Error(), "Will not delete all non-push URLs") {
		return ErrNotDeleteNonPushURLs
	}
	return err
}

// RemoteSetURLAdd appends an URL to the remote with given name of the
// repository. Use RemoteSetURL to overwrite the URL(s) instead.
func (r *Repository) RemoteSetURLAdd(name, newurl string, opts ...RemoteSetURLAddOptions) error {
	return RemoteSetURLAdd(r.path, name, newurl, opts...)
}

// RemoteSetURLDeleteOptions contains arguments for deleting an URL of a remote
// of the repository.
//
// Docs: https://git-scm.com/docs/git-remote#Documentation/git-remote.txt-emset-urlem
type RemoteSetURLDeleteOptions struct {
	// Indicates whether to get push URLs instead of fetch URLs.
	Push bool
	// The timeout duration before giving up for each shell command execution. The
	// default timeout duration will be used when not supplied.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// RemoteSetURLDelete deletes the remote with given name of the repository in
// given path.
func RemoteSetURLDelete(repoPath, name, regex string, opts ...RemoteSetURLDeleteOptions) error {
	var opt RemoteSetURLDeleteOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("remote", "set-url").
		AddOptions(opt.CommandOptions).
		AddArgs("--delete")
	if opt.Push {
		cmd.AddArgs("--push")
	}

	cmd.AddArgs(name, regex)

	_, err := cmd.RunInDirWithTimeout(opt.Timeout, repoPath)
	if err != nil && strings.Contains(err.Error(), "Will not delete all non-push URLs") {
		return ErrNotDeleteNonPushURLs
	}
	return err
}

// RemoteSetURLDelete deletes all URLs matching regex of the remote with given
// name of the repository.
func (r *Repository) RemoteSetURLDelete(name, regex string, opts ...RemoteSetURLDeleteOptions) error {
	return RemoteSetURLDelete(r.path, name, regex, opts...)
}
