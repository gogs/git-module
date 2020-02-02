// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// HookName is the name of a Git hook.
type HookName string

// A list of Git server hooks' name that are supported.
const (
	HookPreReceive  HookName = "pre-receive"
	HookUpdate      HookName = "update"
	HookPostReceive HookName = "post-receive"
)

// ServerSideHooks contains a list of Git hooks are supported on the server side.
var ServerSideHooks = []HookName{HookPreReceive, HookUpdate, HookPostReceive}

// Hook contains information of a Git hook.
type Hook struct {
	name     HookName
	path     string // The absolute file path of the hook.
	isSample bool   // Indicates whether this hook is read from the sample.
	content  string // The content of the hook.
}

// DefaultHooksDir is the default directory for Git hooks.
const DefaultHooksDir = "hooks"

// GetHook returns a Git hook by given name in the repository. It returns an os.ErrNotExist
// if both active and sample hook do not exist.
func GetHook(repoPath string, name HookName) (*Hook, error) {
	// 1. Check if there is an active hook.
	fpath := path.Join(repoPath, DefaultHooksDir)
	if isFile(fpath) {
		p, err := ioutil.ReadFile(fpath)
		if err != nil {
			return nil, err
		}
		return &Hook{
			name:    name,
			path:    fpath,
			content: string(p),
		}, nil
	}

	// 2. Check if a sample file exists.
	fpath = path.Join(repoPath, DefaultHooksDir, string(name)) + ".sample"
	if isFile(fpath) {
		p, err := ioutil.ReadFile(fpath)
		if err != nil {
			return nil, err
		}
		return &Hook{
			name:     name,
			path:     fpath,
			isSample: true,
			content:  string(p),
		}, nil
	}

	return nil, os.ErrNotExist
}

// Name returns the name of the Git hook.
func (h *Hook) Name() HookName {
	return h.name
}

// Path returns the absolute file path of the Git hook.
func (h *Hook) Path() string {
	return h.path
}

// IsSample returns true if the content is read from the sample hook.
func (h *Hook) IsSample() bool {
	return h.isSample
}

// Content returns the content of the Git hook.
func (h *Hook) Content() string {
	return h.content
}

// Update writes the content of the Git hook on filesystem. It updates the memory copy of
// the content as well.
func (h *Hook) Update(content string) error {
	h.content = strings.TrimSpace(content)
	h.content = strings.Replace(h.content, "\r", "", -1)
	os.MkdirAll(path.Dir(h.path), os.ModePerm)
	return ioutil.WriteFile(h.path, []byte(h.content), os.ModePerm)
}

