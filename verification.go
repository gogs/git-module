// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
)

// Verification represents the PGP payload information of a signed commit.
type Verification struct {
	Verified  bool
	Reason    string
	Signature string
	Payload   string
}

// Helper to get verification data from the commit inforamtion, which looks like these:
//gpgsig -----BEGIN PGP SIGNATURE-----
// Version: GnuPG v2
//
// iQIcBAABCAAGBQJXD4OAAAoJEFYV4RsDRH59URsP/1on/dZKWKQQeogZVe1F1Yi/
// vvmvhEkOIaGhFREi7GA5LLyOonKbTmYoH5/xCuZvOJIp5/KbR5qpdahhfT1J/9fh
// iJAIm6MDSXAAiRMASLQVcwBmJTweOwm5LaKZxdY70s8WWqnN4hQt1irodzxpikLl
// EQ2rfbvfOP4/MDYkQUI1Yvb3e+cNK2o0R1DjFbfSE5xX9X+miqnOjIvmBZ7vL3Hp
// GhxJ9dtGyhM7vsGiWk42dCbOnJshCeJnCZIeXKH6Xlo6EJnwiGAvFUy4UQP7bhzO
// ZgE+leWrUiyPs7P1OYIMV6sXPpMZmKh/UVOjEmxzbC8P6/ye5pURYZpkB70P7d2w
// bbxnLmVDK+pIedAdY3VWOhrAg26Jmq/i51un+OsYet3rpPOPC9Q9WzRg/s9aMg+S
// hLle77kjzAqK2m38qIJjVRZFFRM00WW4GnbmSu1xJw125jEfNnqjS5CfioQ+MyYN
// 9ARfLk4hTe5gZ/jgJ8AFQWygEruQxzUAkZLgeFt6TbOm5HSmTh2OpSJCupwJjwNu
// iMXQ0gLF99rUs5vtEXqDs5xfEYxdb1H/dDe++Of+NDcXcoJE4LtdK9kP8/ilYiBu
// MlShuryaeNtdNB6javCBA1mXwI7WIOhYlFzaNQ3KW2+vTA3VjiGJLB5jjYGmgrpz
// 0SuOoRPfFT3QY4xrOXIR
// =aEJU
// -----END PGP SIGNATURE-----
// but without the "gpgsig " at the beginning
//
func newVerificationFromCommitline(line []byte) (_ *Verification, err error) {
	verif := new(Verification)

	signatureEnd := bytes.LastIndex(line, []byte("-----END PGP SIGNATURE-----"))
	verif.Signature = string(line[:signatureEnd+27])

	return verif, nil
}
