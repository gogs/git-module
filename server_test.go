package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateServerInfo(t *testing.T) {
	ctx := context.Background()
	err := os.RemoveAll(filepath.Join(repoPath, "info"))
	require.NoError(t, err)
	err = UpdateServerInfo(ctx, repoPath, UpdateServerInfoOptions{Force: true})
	require.NoError(t, err)
	assert.True(t, isFile(filepath.Join(repoPath, "info", "refs")))
}

func TestReceivePack(t *testing.T) {
	ctx := context.Background()
	got, err := ReceivePack(ctx, repoPath, ReceivePackOptions{HTTPBackendInfoRefs: true})
	require.NoError(t, err)
	var regPattern = fmt.Sprintf(
		"report-status report-status-v2 delete-refs side-band-64k quiet atomic ofs-delta (push-options |)object-format=%s agent=git/",
		testrepoObjectFormat,
	)
	assert.Regexp(t, regPattern, string(got))
}

func TestUploadPack(t *testing.T) {
	ctx := context.Background()
	got, err := UploadPack(ctx, repoPath,
		UploadPackOptions{
			StatelessRPC:        true,
			Strict:              true,
			HTTPBackendInfoRefs: true,
		},
	)
	require.NoError(t, err)
	var regPattern = fmt.Sprintf("multi_ack thin-pack side-band side-band-64k ofs-delta shallow deepen-since deepen-not deepen-relative no-progress include-tag multi_ack_detailed (allow-tip-sha1-in-want |)(allow-reachable-sha1-in-want |)no-done symref=HEAD:refs/heads/master (filter |)object-format=%s agent=git/", testrepoObjectFormat)
	assert.Regexp(t, regPattern, string(got))
}
