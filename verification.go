// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

// Verification represents the PGP payload information of a signed commit.
type Verification struct {
	Verified  bool
	Reason    string
	Signature string
	Payload   string
}

