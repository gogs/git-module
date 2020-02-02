package git

import (
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
