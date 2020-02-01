package git

import (
	"bufio"
	"io"
	"strings"
)

// Submodules contains information of submodules.
type Submodules = *objectCache

// Submodules returns submodules found in this commit.
func (c *Commit) Submodules() (Submodules, error) {
	c.submodulesOnce.Do(func() {
		var e *TreeEntry
		e, c.submodulesErr = c.GetTreeEntryByPath(".gitmodules")
		if c.submodulesErr != nil {
			return
		}

		var r io.Reader
		r, c.submodulesErr = e.Blob().Data()
		if c.submodulesErr != nil {
			return
		}

		scanner := bufio.NewScanner(r)
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
						Name: path,
						URL:  strings.TrimSpace(fields[1])},
					)
					inSection = false
				}
			}
		}
	})

	return c.submodules, c.submodulesErr
}

// Submodule returns submodule by given name.
func (c *Commit) Submodule(name string) (*Submodule, error) {
	mods, err := c.Submodules()
	if err != nil {
		return nil, err
	}

	m, has := mods.Get(name)
	if has {
		return m.(*Submodule), nil
	}
	return nil, nil
}
