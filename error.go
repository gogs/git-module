// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"fmt"
	"time"
)

type ErrExecTimeout struct {
	Duration time.Duration
}

func IsErrExecTimeout(err error) bool {
	_, ok := err.(ErrExecTimeout)
	return ok
}

func (err ErrExecTimeout) Error() string {
	return fmt.Sprintf("execution timed out [duration: %v]", err.Duration)
}

type ErrRevisionNotExist struct {
	Rev  string
	Path string
}

func IsErrRevesionNotExist(err error) bool {
	_, ok := err.(ErrRevisionNotExist)
	return ok
}

func (err ErrRevisionNotExist) Error() string {
	return fmt.Sprintf("revision does not exist [rev: %s, path: %s]", err.Rev, err.Path)
}

type ErrNoMergeBase struct{}

func IsErrNoMergeBase(err error) bool {
	_, ok := err.(ErrNoMergeBase)
	return ok
}

func (err ErrNoMergeBase) Error() string {
	return "no merge based found"
}
