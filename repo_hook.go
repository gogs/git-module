// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"os"
	"path"
)

func (repo *Repository) Hook(name HookName) (*Hook, error) {
	return GetHook(repo.Path, name)
}

// ListHooks returns a list of Git hooks found in given repository. It may return an empty slice
// when no hooks found.
func ListHooks(repoPath string) ([]*Hook, error) {
	if !isDir(path.Join(repoPath, DefaultHooksDir)) {
		return []*Hook{}, nil
	}

	hooks := make([]*Hook, 0, len(ServerSideHooks))
	for _, name := range ServerSideHooks {
		h, err := GetHook(repoPath, name)
		if err != nil {
			if err == os.ErrNotExist {
				continue
			}
			return nil, err
		}
		hooks = append(hooks, h)
	}
	return hooks, nil
}

func (repo *Repository) Hooks() ([]*Hook, error) {
	return ListHooks(repo.Path)
}
