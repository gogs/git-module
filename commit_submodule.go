// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bufio"
	"bytes"
	"strings"
)

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
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "[submodule") {
				inSection = true
				continue
			}
			if inSection {
				fields := strings.Split(scanner.Text(), "=")
				k := strings.TrimSpace(fields[0])
				if k == "path" {
					path = strings.TrimSpace(fields[1])
				} else if k == "url" {
					c.submodules.Set(path, &Submodule{
						name: path,
						url:  strings.TrimSpace(fields[1])},
					)
					inSection = false
				}
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
