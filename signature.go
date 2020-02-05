// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"strconv"
	"time"
)

// Signature represents a author or committer.
type Signature struct {
	// The name of the person.
	Name string
	// The email address.
	Email string
	// The time of the signurate.
	When time.Time
}

// parseSignature parses signature information from the (uncompressed) commit line,
// which looks like the following but without the "author " at the beginning:
//     author Patrick Gundlach <gundlach@speedata.de> 1378823654 +0200
//     author Patrick Gundlach <gundlach@speedata.de> Thu, 07 Apr 2005 22:13:13 +0200
// This method should only be used for parsing author and committer.
func parseSignature(line []byte) (*Signature, error) {
	emailStart := bytes.IndexByte(line, '<')
	emailEnd := bytes.IndexByte(line, '>')
	sig := &Signature{
		Name:  string(line[:emailStart-1]),
		Email: string(line[emailStart+1 : emailEnd]),
	}

	// Check the date format
	firstChar := line[emailEnd+2]
	if firstChar >= 48 && firstChar <= 57 { // ASCII code for 0-9
		timestop := bytes.IndexByte(line[emailEnd+2:], ' ')
		timestamp := line[emailEnd+2 : emailEnd+2+timestop]
		seconds, err := strconv.ParseInt(string(timestamp), 10, 64)
		if err != nil {
			return nil, err
		}
		sig.When = time.Unix(seconds, 0)

		// Handle the timezone shift
		timezone := line[timestop+1:]

		// Discard malformed timezone
		if len(timezone) != 5 {
			return sig, err
		}

		// Add or minus, default to add
		var direction int64 = 1
		if timezone[0] == '-' {
			direction = -1
		}

		hours, _ := strconv.ParseInt(string(timezone[1:3]), 10, 64)
		minutes, _ := strconv.ParseInt(string(timezone[3:]), 10, 64)
		sig.When = sig.When.Add(time.Duration(direction*hours) * time.Hour)
		sig.When = sig.When.Add(time.Duration(direction*minutes) * time.Minute)
		return sig, nil
	}

	var err error
	sig.When, err = time.Parse("Mon Jan _2 15:04:05 2006 -0700", string(line[emailEnd+2:]))
	if err != nil {
		return nil, err
	}
	return sig, nil
}
