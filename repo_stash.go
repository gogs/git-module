package git

import (
	"bytes"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// Stash represents a stash in the repository.
type Stash struct {
	// Index is the index of the stash.
	Index int

	// Message is the message of the stash.
	Message string

	// Files is the list of files in the stash.
	Files []string
}

// StashListOptions describes the options for the StashList function.
type StashListOptions struct {
	// CommandOptions describes the options for the command.
	CommandOptions
}

var stashLineRegexp = regexp.MustCompile(`^stash@\{(\d+)\}: (.*)$`)

// StashList returns a list of stashes in the repository.
// This must be run in a work tree.
func (r *Repository) StashList(opts ...StashListOptions) ([]*Stash, error) {
	var opt StashListOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	stash := make([]*Stash, 0)
	cmd := NewCommand("stash", "list", "--name-only").AddOptions(opt.CommandOptions)
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	if err := cmd.RunInDirPipeline(stdout, stderr, r.path); err != nil {
		return nil, concatenateError(err, stderr.String())
	}

	var entry *Stash
	lines := strings.Split(stdout.String(), "\n")
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		// Init entry
		if match := stashLineRegexp.FindStringSubmatch(line); len(match) == 3 {
			if entry != nil {
				stash = append(stash, entry)
			}

			idx, err := strconv.Atoi(match[1])
			if err != nil {
				idx = -1
			}
			entry = &Stash{
				Index:   idx,
				Message: match[2],
				Files:   make([]string, 0),
			}
		} else if entry != nil && line != "" {
			entry.Files = append(entry.Files, line)
		} else {
			continue
		}
	}

	if entry != nil {
		stash = append(stash, entry)
	}

	return stash, nil
}

// StashDiff returns a parsed diff object for the given stash index.
// This must be run in a work tree.
func (r *Repository) StashDiff(index int, maxFiles, maxFileLines, maxLineChars int, opts ...DiffOptions) (*Diff, error) {
	var opt DiffOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("stash", "show", "-p", "--full-index", "-M", strconv.Itoa(index)).AddOptions(opt.CommandOptions)
	stdout, w := io.Pipe()
	done := make(chan SteamParseDiffResult)
	go StreamParseDiff(stdout, done, maxFiles, maxFileLines, maxLineChars)

	stderr := new(bytes.Buffer)
	err := cmd.RunInDirPipelineWithTimeout(opt.Timeout, w, stderr, r.path)
	_ = w.Close() // Close writer to exit parsing goroutine
	if err != nil {
		return nil, concatenateError(err, stderr.String())
	}

	result := <-done
	return result.Diff, result.Err
}

// StashPushOptions describes the options for the StashPush function.
type StashPushOptions struct {
	// CommandOptions describes the options for the command.
	CommandOptions
}

// StashPush pushes the current worktree to the stash.
// This must be run in a work tree.
func (r *Repository) StashPush(msg string, opts ...StashPushOptions) error {
	var opt StashPushOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	cmd := NewCommand("stash", "push")
	if msg != "" {
		cmd.AddArgs("-m", msg)
	}
	cmd.AddOptions(opt.CommandOptions)

	_, err := cmd.RunInDir(r.path)
	return err
}
