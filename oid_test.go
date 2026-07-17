package git

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOid_Equal(t *testing.T) {
	tests := []struct {
		s1     Oid
		s2     interface{}
		expVal bool
	}{
		{
			s1:     MustIDFromString("fcf7087e732bfe3c25328248a9bf8c3ccd85bed4"),
			s2:     "fcf7087e732bfe3c25328248a9bf8c3ccd85bed4",
			expVal: true,
		}, {
			s1:     MustIDFromString("fcf7087e732bfe3c25328248a9bf8c3ccd85bed4"),
			s2:     EmptySha1ID,
			expVal: false,
		},

		{
			s1:     MustIDFromString("fcf7087e732bfe3c25328248a9bf8c3ccd85bed4"),
			s2:     MustIDFromString("fcf7087e732bfe3c25328248a9bf8c3ccd85bed4").(*SHA1).bytes,
			expVal: true,
		},
		{
			s1:     MustIDFromString("fcf7087e732bfe3c25328248a9bf8c3ccd85bed4"),
			s2:     MustIDFromString(EmptySha1ID).(*SHA1).bytes,
			expVal: false,
		},
		{
			s1:     MustIDFromString("fcf7087e732bfe3c25328248a9bf8c3ccd85bed4"),
			s2:     MustIDFromString("fcf7087e732bfe3c25328248a9bf8c3ccd85bed4"),
			expVal: true,
		},
		{
			s1:     MustIDFromString("fcf7087e732bfe3c25328248a9bf8c3ccd85bed4"),
			s2:     MustIDFromString(EmptySha1ID),
			expVal: false,
		},

		{
			s1:     MustIDFromString("fcf7087e732bfe3c25328248a9bf8c3ccd85bed4"),
			s2:     []byte(EmptySha1ID),
			expVal: false,
		},

		{
			s1:     MustIDFromString("ba2cc18e595b5f70a47cea3ed71f3568addd55e7bb5581a8918b7885abb7888e"),
			s2:     "ba2cc18e595b5f70a47cea3ed71f3568addd55e7bb5581a8918b7885abb7888e",
			expVal: true,
		},
		{
			s1:     MustIDFromString("ba2cc18e595b5f70a47cea3ed71f3568addd55e7bb5581a8918b7885abb7888e"),
			s2:     EmptySha256ID,
			expVal: false,
		},
		{
			s1:     MustIDFromString("ba2cc18e595b5f70a47cea3ed71f3568addd55e7bb5581a8918b7885abb7888e"),
			s2:     MustIDFromString("ba2cc18e595b5f70a47cea3ed71f3568addd55e7bb5581a8918b7885abb7888e").(*SHA256).bytes,
			expVal: true,
		},
		{
			s1:     MustIDFromString("ba2cc18e595b5f70a47cea3ed71f3568addd55e7bb5581a8918b7885abb7888e"),
			s2:     MustIDFromString(EmptySha256ID).(*SHA256).bytes,
			expVal: false,
		},
		{
			s1:     MustIDFromString("ba2cc18e595b5f70a47cea3ed71f3568addd55e7bb5581a8918b7885abb7888e"),
			s2:     MustIDFromString("ba2cc18e595b5f70a47cea3ed71f3568addd55e7bb5581a8918b7885abb7888e"),
			expVal: true,
		},
		{
			s1:     MustIDFromString("ba2cc18e595b5f70a47cea3ed71f3568addd55e7bb5581a8918b7885abb7888e"),
			s2:     []byte(EmptySha256ID),
			expVal: false,
		},

		{
			s1:     MustIDFromString("ba2cc18e595b5f70a47cea3ed71f3568addd55e7bb5581a8918b7885abb7888e"),
			s2:     MustIDFromString("ba2cc18e595b5f70a47cea3ed71f3568addd55e7"),
			expVal: false,
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, test.expVal, test.s1.Equal(test.s2))
		})
	}
}

func TestNewID(t *testing.T) {
	sha, err := NewID([]byte("000000"))
	assert.Equal(t, errors.New("length must be 20 or 32"), err)
	assert.Nil(t, sha)
}

func TestNewIDFromString(t *testing.T) {
	sha, err := NewIDFromString("000000")
	assert.Equal(t, errors.New("length must be 40 or 64"), err)
	assert.Nil(t, sha)
}
