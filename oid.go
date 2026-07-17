package git

import (
	"encoding/hex"
	"errors"
	"strings"
	"sync"
)

const (
	// EmptySha1ID is an ID with empty SHA-1 hash.
	EmptySha1ID = "0000000000000000000000000000000000000000"
	// EmptySha256ID is an ID with empty SHA-256 hash.
	EmptySha256ID = "0000000000000000000000000000000000000000000000000000000000000000"
)

// Oid is the id of a Git object
type Oid interface {
	Equal(interface{}) bool
	String() string
	Bytes() []byte
}

// SHA1 is the SHA-1 hash of a Git object.
type SHA1 struct {
	bytes [20]byte

	str     string
	strOnce sync.Once
}

// Equal returns true if s2 has the same SHA1 as s. It supports
// 40-length-string, []byte, and SHA1.
func (s *SHA1) Equal(s2 interface{}) bool {
	switch v := s2.(type) {
	case string:
		return v == s.String()
	case [20]byte:
		return v == s.bytes
	case *SHA1:
		return v.bytes == s.bytes
	}
	return false
}

// String returns string (hex) representation of the SHA1.
func (s *SHA1) String() string {
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

// Bytes returns bytes (length 20) representation of the SHA1.
func (s *SHA1) Bytes() []byte {
	return s.bytes[:]
}

// SHA256 is the SHA-256 hash of a Git object.
type SHA256 struct {
	bytes [32]byte

	str     string
	strOnce sync.Once
}

// Equal returns true if s2 has the same SHA256 as s. It supports
// 64-length-string, []byte, and SHA256.
func (s *SHA256) Equal(s2 interface{}) bool {
	switch v := s2.(type) {
	case string:
		return v == s.String()
	case [32]byte:
		return v == s.bytes
	case *SHA256:
		return v.bytes == s.bytes
	}
	return false

}

// String returns string (hex) representation of the SHA256.
func (s *SHA256) String() string {
	s.strOnce.Do(func() {
		result := make([]byte, 0, 64)
		hexvalues := []byte("0123456789abcdef")
		for i := 0; i < 32; i++ {
			result = append(result, hexvalues[s.bytes[i]>>4])
			result = append(result, hexvalues[s.bytes[i]&0xf])
		}
		s.str = string(result)
	})
	return s.str
}

// Bytes returns bytes (length 32) representation of the SHA256.
func (s *SHA256) Bytes() []byte {
	return s.bytes[:]
}

// MustID always returns a new Oid from a [20]byte or [32]byte array with no validation of input.
func MustID(b []byte) Oid {
	if len(b) < 32 {
		var id SHA1
		for i := 0; i < 20; i++ {
			id.bytes[i] = b[i]
		}
		return &id
	} else {
		var id SHA256
		for i := 0; i < 32; i++ {
			id.bytes[i] = b[i]
		}
		return &id
	}
}

// NewID returns a new Oid from a [20]byte or [32]byte array.
func NewID(b []byte) (Oid, error) {
	if len(b) != 20 && len(b) != 32 {
		return nil, errors.New("length must be 20 or 32")
	}
	return MustID(b), nil
}

// MustIDFromString always returns a new sha from a ID with no validation of
// input.
func MustIDFromString(s string) Oid {
	b, _ := hex.DecodeString(s)
	return MustID(b)
}

// NewIDFromString returns a new Oid from a ID string of length 40 or 64.
func NewIDFromString(s string) (Oid, error) {
	s = strings.TrimSpace(s)
	if len(s) != 40 && len(s) != 64 {
		return nil, errors.New("length must be 40 or 64")
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return NewID(b)
}
