// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

var (
	// logOutput is the writer to write logs. When not set, no log will be produced.
	logOutput io.Writer
	// logPrefix is the prefix prepend to each log entry.
	logPrefix = "[git-module] "
)

// SetOutput sets the output writer for logs.
func SetOutput(output io.Writer) {
	logOutput = output
}

// SetPrefix sets the prefix to be prepended to each log entry.
func SetPrefix(prefix string) {
	logPrefix = prefix
}

func log(format string, args ...interface{}) {
	if logOutput == nil {
		return
	}

	fmt.Fprint(logOutput, logPrefix)
	fmt.Fprintf(logOutput, format, args...)
	fmt.Fprintln(logOutput)
}

var (
	// gitVersion stores the Git binary version.
	gitVersion     string
	gitVersionOnce sync.Once
	gitVersionErr  error
)

// BinVersion returns current Git binary version that is used by this module.
func BinVersion() (string, error) {
	gitVersionOnce.Do(func() {
		var stdout string
		stdout, gitVersionErr = NewCommand("version").Run()
		if gitVersionErr != nil {
			return
		}

		fields := strings.Fields(stdout)
		if len(fields) < 3 {
			gitVersionErr = fmt.Errorf("not enough output: %s", stdout)
			return
		}

		// Handle special case on Windows.
		i := strings.Index(fields[2], "windows")
		if i >= 1 {
			gitVersion = fields[2][:i-1]
			return
		}

		gitVersion = fields[2]
	})

	return gitVersion, gitVersionErr
}

// FsckWithTimeout verifies the connectivity and validity of the objects in the database
// with given timeout duration.
func FsckWithTimeout(timeout time.Duration, repoPath string, args ...string) error {
	_, err := NewCommand("fsck").AddArgs(args...).RunInDirWithTimeout(timeout, repoPath)
	return err
}
