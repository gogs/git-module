// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

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
