// Copyright 2019 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

const (
	// DEPRECATED: use ArchiveZip instead
	ZIP = ArchiveZip
	// DEPRECATED: use ArchiveTarGz instead
	TARGZ = ArchiveTarGz
)

// DEPRECATED: use BranchPrefix instead
const BRANCH_PREFIX = BranchPrefix

// DEPRECATED: use RemotePrefix instead
const REMOTE_PREFIX = RemotePrefix

const (
	// DEPRECATED: use DiffLinePlain instead
	DIFF_LINE_PLAIN = DiffLinePlain
	// DEPRECATED: use DiffLineAdd instead
	DIFF_LINE_ADD = DiffLineAdd
	// DEPRECATED: use DiffLineDel instead
	DIFF_LINE_DEL = DiffLineDel
	// DEPRECATED: use DiffLineSection instead
	DIFF_LINE_SECTION = DiffLineSection
)

const (
	// DEPRECATED: use DiffFileAdd instead
	DIFF_FILE_ADD = DiffFileAdd
	// DEPRECATED: use DiffFileChange instead
	DIFF_FILE_CHANGE = DiffFileChange
	// DEPRECATED: use DiffFileDel instead
	DIFF_FILE_DEL = DiffFileDel
	// DEPRECATED: use DiffFileRename instead
	DIFF_FILE_RENAME = DiffFileRename
)

const (
	// DEPRECATED: use RawDiffNormal instead
	RAW_DIFF_NORMAL = RawDiffNormal
	// DEPRECATED: use RawDiffPatch instead
	RAW_DIFF_PATCH = RawDiffPatch
)

const (
	// DEPRECATED: use ObjectCommit instead
	OBJECT_COMMIT = ObjectCommit
	// DEPRECATED: use ObjectTree instead
	OBJECT_TREE = ObjectTree
	// DEPRECATED: use ObjectBlob instead
	OBJECT_BLOB = ObjectBlob
	// DEPRECATED: use ObjectTag instead
	OBJECT_TAG = ObjectTag
)

// DEPRECATED: use TagPrefix instead
const TAG_PREFIX = TagPrefix

// DEPRECATED: use EmptySHA instead
const EMPTY_SHA = EmptySHA

const (
	// DEPRECATED: use EntryBlob instead
	ENTRY_MODE_BLOB = EntryBlob
	// DEPRECATED: use EntryExec instead
	ENTRY_MODE_EXEC = EntryExec
	// DEPRECATED: use EntrySymlink instead
	ENTRY_MODE_SYMLINK = EntrySymlink
	// DEPRECATED: use EntryCommit instead
	ENTRY_MODE_COMMIT = EntryCommit
	// DEPRECATED: use EntryTree instead
	ENTRY_MODE_TREE = EntryTree
)
