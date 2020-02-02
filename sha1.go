// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
)

// EmptyID is an ID with empty SHA-1 hash.
const EmptyID = "0000000000000000000000000000000000000000"

// SHA1 is the SHA-1 hash of a Git object.
type SHA1 struct {
	bytes [20]byte

	str     string
	strOnce sync.Once
}

// Equal returns true if s2 has the same SHA1 as s. It supports 40-length-string, []byte, and SHA1.
func (s SHA1) Equal(s2 interface{}) bool {
	switch v := s2.(type) {
	case string:
		if len(v) != 40 {
			return false
		}
		return v == s.String()
	case []byte:
		if len(v) != 20 {
			return false
		}
		for i := range v {
			if s.bytes[i] != v[i] {
				return false
			}
		}
		return true
	case SHA1:
		return v.bytes == s.bytes
	}
	return false
}

// String returns string (hex) representation of the SHA1.
func (s SHA1) String() string {
	s.strOnce.Do(func() {
		result := make([]byte, 0, 40)
		hexvalues := []byte("0123456789abcdef")
		for i := 0; i < 20; i++ {
			result = append(result, hexvalues[s.bytes[i]>>4])
			result = append(result, hexvalues[s.bytes[i]&0xf])
		}
		s.str = string(result)
	})
	return s.str
}

// MustID always returns a new SHA1 from a [20]byte array with no validation of input.
func MustID(b []byte) SHA1 {
	var id SHA1
	for i := 0; i < 20; i++ {
		id.bytes[i] = b[i]
	}
	return id
}

// NewID returns a new SHA1 from a [20]byte array.
func NewID(b []byte) (SHA1, error) {
	if len(b) != 20 {
		return SHA1{}, fmt.Errorf("length must be 20: %v", b)
	}
	return MustID(b), nil
}

// MustIDFromString always returns a new sha from a ID with no validation of input.
func MustIDFromString(s string) SHA1 {
	b, _ := hex.DecodeString(s)
	return MustID(b)
}

// NewIDFromString returns a new SHA1 from a ID string of length 40.
func NewIDFromString(s string) (SHA1, error) {
	var id SHA1
	s = strings.TrimSpace(s)
	if len(s) != 40 {
		return id, fmt.Errorf("length must be 40: %s", s)
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return id, err
	}
	return NewID(b)
}
