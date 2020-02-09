// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"fmt"
	"path"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// EntryMode is the unix file mode of a tree entry.
type EntryMode int

// There are only a few file modes in Git. They look like unix file modes, but they can only be
// one of these.
const (
	EntryTree    EntryMode = 0040000
	EntryBlob    EntryMode = 0100644
	EntryExec    EntryMode = 0100755
	EntrySymlink EntryMode = 0120000
	EntryCommit  EntryMode = 0160000
)

type TreeEntry struct {
	mode EntryMode
	typ  ObjectType
	id   *SHA1
	name string

	parent *Tree

	size     int64
	sizeOnce sync.Once
}

// IsTree returns tree if the entry itself is another tree (i.e. a directory).
func (e *TreeEntry) IsTree() bool {
	return e.mode == EntryTree
}

// IsBlob returns true if the entry is a blob.
func (e *TreeEntry) IsBlob() bool {
	return e.mode == EntryBlob
}

// IsExec returns tree if the entry is an executable.
func (e *TreeEntry) IsExec() bool {
	return e.mode == EntryExec
}

// IsSymlink returns true if the entry is a symbolic link.
func (e *TreeEntry) IsSymlink() bool {
	return e.mode == EntrySymlink
}

// IsCommit returns true if the entry is a commit (i.e. a submodule).
func (e *TreeEntry) IsCommit() bool {
	return e.mode == EntryCommit
}

// Type returns the object type of the entry.
func (e *TreeEntry) Type() ObjectType {
	return e.typ
}

// ID returns the SHA-1 hash of the entry.
func (e *TreeEntry) ID() *SHA1 {
	return e.id
}

// Name returns name of the entry.
func (e *TreeEntry) Name() string {
	return e.name
}

// Size returns the size of thr entry.
func (e *TreeEntry) Size() int64 {
	e.sizeOnce.Do(func() {
		if e.IsTree() {
			return
		}

		stdout, err := NewCommand("cat-file", "-s", e.id.String()).RunInDir(e.parent.repo.path)
		if err != nil {
			return
		}
		e.size, _ = strconv.ParseInt(strings.TrimSpace(string(stdout)), 10, 64)
	})

	return e.size
}

// Blob returns a blob object from the entry.
func (e *TreeEntry) Blob() *Blob {
	return &Blob{
		TreeEntry: e,
	}
}

type Entries []*TreeEntry

var sorter = []func(t1, t2 *TreeEntry) bool{
	func(t1, t2 *TreeEntry) bool {
		return (t1.IsTree() || t1.IsCommit()) && !t2.IsTree() && !t2.IsCommit()
	},
	func(t1, t2 *TreeEntry) bool {
		return t1.name < t2.name
	},
}

func (es Entries) Len() int      { return len(es) }
func (es Entries) Swap(i, j int) { es[i], es[j] = es[j], es[i] }
func (es Entries) Less(i, j int) bool {
	t1, t2 := es[i], es[j]
	var k int
	for k = 0; k < len(sorter)-1; k++ {
		sort := sorter[k]
		switch {
		case sort(t1, t2):
			return true
		case sort(t2, t1):
			return false
		}
	}
	return sorter[k](t1, t2)
}

func (es Entries) Sort() {
	sort.Sort(es)
}

var defaultConcurrency = runtime.NumCPU()

type commitInfo struct {
	entryName string
	infos     []interface{}
	err       error
}

// CommitsInfo takes advantages of concurrency to speed up getting information
// of all commits that are corresponding to these entries. This method will automatically
// choose the right number of goroutine (concurrency) to use related of the host CPU.
func (es Entries) CommitsInfo(timeout time.Duration, commit *Commit, treePath string) ([][]interface{}, error) {
	return es.CommitsInfoWithCustomConcurrency(timeout, commit, treePath, 0)
}

// CommitsInfoWithCustomConcurrency takes advantages of concurrency to speed up getting information
// of all commits that are corresponding to these entries. If the given maxConcurrency is negative or
// equal to zero: the right number of goroutine (concurrency) to use will be choosen related of the
// host CPU.
func (es Entries) CommitsInfoWithCustomConcurrency(timeout time.Duration, commit *Commit, treePath string, maxConcurrency int) ([][]interface{}, error) {
	if len(es) == 0 {
		return nil, nil
	}

	if maxConcurrency <= 0 {
		maxConcurrency = defaultConcurrency
	}

	// Length of taskChan determines how many goroutines (subprocesses) can run at the same time.
	// The length of revChan should be same as taskChan so goroutines whoever finished job can
	// exit as early as possible, only store data inside channel.
	taskChan := make(chan bool, maxConcurrency)
	revChan := make(chan commitInfo, maxConcurrency)
	doneChan := make(chan error)

	// Receive loop will exit when it collects same number of data pieces as tree entries.
	// It notifies doneChan before exits or notify early with possible error.
	infoMap := make(map[string][]interface{}, len(es))
	go func() {
		i := 0
		for info := range revChan {
			if info.err != nil {
				doneChan <- info.err
				return
			}

			infoMap[info.entryName] = info.infos
			i++
			if i == len(es) {
				break
			}
		}
		doneChan <- nil
	}()

	for i := range es {
		// When taskChan is idle (or has empty slots), put operation will not block.
		// However when taskChan is full, code will block and wait any running goroutines to finish.
		taskChan <- true

		if es[i].typ != ObjectCommit {
			go func(i int) {
				cinfo := commitInfo{entryName: es[i].Name()}
				c, err := commit.CommitByPath(CommitByRevisionOptions{
					Path:    path.Join(treePath, es[i].Name()),
					Timeout: timeout,
				})
				if err != nil {
					cinfo.err = fmt.Errorf("get commit by path (%s/%s): %v", treePath, es[i].Name(), err)
				} else {
					cinfo.infos = []interface{}{es[i], c}
				}
				revChan <- cinfo
				<-taskChan // Clear one slot from taskChan to allow new goroutines to start.
			}(i)
			continue
		}

		// Handle submodule
		go func(i int) {
			cinfo := commitInfo{entryName: es[i].Name()}
			sm, err := commit.Submodule(path.Join(treePath, es[i].Name()))
			if err != nil && err != ErrSubmoduleNotExist {
				cinfo.err = fmt.Errorf("get submodule (%s/%s): %v", treePath, es[i].Name(), err)
				revChan <- cinfo
				return
			}

			smURL := ""
			if sm != nil {
				smURL = sm.url
			}

			c, err := commit.CommitByPath(CommitByRevisionOptions{
				Path:    path.Join(treePath, es[i].Name()),
				Timeout: timeout,
			})
			if err != nil {
				cinfo.err = fmt.Errorf("get commit by path (%s/%s): %v", treePath, es[i].Name(), err)
			} else {
				cinfo.infos = []interface{}{
					es[i],
					&SubmoduleFile{
						Commit: c,
						refURL: smURL,
						refID:  es[i].id.String(),
					},
				}
			}
			revChan <- cinfo
			<-taskChan
		}(i)
	}

	if err := <-doneChan; err != nil {
		return nil, err
	}

	commitsInfo := make([][]interface{}, len(es))
	for i := 0; i < len(es); i++ {
		commitsInfo[i] = infoMap[es[i].Name()]
	}
	return commitsInfo, nil
}
