// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"strings"
	"sync"
)

// Submodule contains information of a Git submodule.
type Submodule struct {
	name string
	url  string
}

// Name returns the name of the submodule.
func (s *Submodule) Name() string {
	return s.name
}

// URL returns the URL of the submodule.
func (s *Submodule) URL() string {
	return s.url
}

// SubmoduleFile is a file with submodule type.
type SubmoduleFile struct {
	*Commit

	refID      string
	refURL     string
	refURLOnce sync.Once
}

// RefURL guesses and returns the reference URL.
func (f *SubmoduleFile) RefURL(urlPrefix, parentPath string) string {
	f.refURLOnce.Do(func() {
		f.refURL = strings.TrimSuffix(f.refURL, ".git")

		// git://xxx/user/repo
		if strings.HasPrefix(f.refURL, "git://") {
			f.refURL = "http://" + strings.TrimPrefix(f.refURL, "git://")
			return
		}

		// http[s]://xxx/user/repo
		if strings.HasPrefix(f.refURL, "http://") || strings.HasPrefix(f.refURL, "https://") {
			return
		}

		// Relative URL prefix check (according to Git submodule documentation)
		if strings.HasPrefix(f.refURL, "./") || strings.HasPrefix(f.refURL, "../") {
			// ...construct and return correct submodule URL here.
			idx := strings.Index(parentPath, "/src/")
			if idx == -1 {
				return
			}
			f.refURL = strings.TrimSuffix(urlPrefix, "/") + parentPath[:idx] + "/" + f.refURL
			return
		}

		// sysuser@xxx:user/repo
		i := strings.Index(f.refURL, "@")
		j := strings.LastIndex(f.refURL, ":")

		// Only process when i < j because git+ssh://git@git.forwardbias.in/npploader.git
		if i > -1 && j > -1 && i < j {
			// Fix problem with reverse proxy works only with local server
			if strings.Contains(urlPrefix, f.refURL[i+1:j]) {
				f.refURL = urlPrefix + f.refURL[j+1:]
				return
			}

			f.refURL = "http://" + f.refURL[i+1:j] + "/" + f.refURL[j+1:]
			return
		}
	})

	return f.refURL
}

// RefID returns the reference ID.
func (f *SubmoduleFile) RefID() string {
	return f.refID
}
