package git

import (
	"strings"
	"time"
)

// IsURLAccessible returns true if given remote URL is accessible via Git.
func IsURLAccessible(timeout time.Duration, url string) bool {
	_, err := LsRemote(LsRemoteOptions{
		URL:     url,
		Timeout: timeout,
	})
	return err == nil
}

// RefEndName returns short name of heads or tags. Other references will retrun original string.
func RefEndName(ref string) string {
	if strings.HasPrefix(ref, RefsHeads) {
		return ref[len(RefsHeads):]
	}

	if strings.HasPrefix(ref, RefsTags) {
		return ref[len(RefsTags):]
	}

	return ref
}
