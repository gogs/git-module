// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bufio"
	"bytes"
	"strings"
)

// Submodule contains information of a Git submodule.
type Submodule struct {
	// The name of the submodule.
	Name string
	// The URL of the submodule.
	URL string
	// The commit ID of the subproject.
	Commit string
}

// Submodules contains information of submodules.
type Submodules = *objectCache

// Submodules returns submodules found in this commit.
func (c *Commit) Submodules() (Submodules, error) {
	c.submodulesOnce.Do(func() {
		var e *TreeEntry
		e, c.submodulesErr = c.TreeEntry(".gitmodules")
		if c.submodulesErr != nil {
			return
		}

		var p []byte
		p, c.submodulesErr = e.Blob().Bytes()
		if c.submodulesErr != nil {
			return
		}

		scanner := bufio.NewScanner(bytes.NewReader(p))
		c.submodules = newObjectCache()
		var inSection bool
		var path string
		var url string
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "[submodule") {
				inSection = true
				path = ""
				url = ""
				continue
			} else if !inSection {
				continue
			}

			fields := strings.Split(scanner.Text(), "=")
			switch strings.TrimSpace(fields[0]) {
			case "path":
				path = strings.TrimSpace(fields[1])
			case "url":
				url = strings.TrimSpace(fields[1])
			}

			if len(path) > 0 && len(url) > 0 {
				mod := &Submodule{
					Name: path,
					URL:  url,
				}

				mod.Commit, c.submodulesErr = c.repo.RevParse(c.id.String() + ":" + mod.Name)
				if c.submodulesErr != nil {
					return
				}

				c.submodules.Set(path, mod)
				inSection = false
			}
		}
	})

	return c.submodules, c.submodulesErr
}

// Submodule returns submodule by given name. It returns an ErrSubmoduleNotExist
// if the path does not exist as a submodule.
func (c *Commit) Submodule(path string) (*Submodule, error) {
	mods, err := c.Submodules()
	if err != nil {
		return nil, err
	}

	m, has := mods.Get(path)
	if has {
		return m.(*Submodule), nil
	}
	return nil, ErrSubmoduleNotExist
}
