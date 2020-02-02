package git

import (
	"time"
)

// IsURLAccessible returns true if given remote URL is accessible via Git.
func IsURLAccessible(url string, timeout time.Duration) bool {
	_, err := LsRemote(LsRemoteOptions{
		URL:     url,
		Timeout: timeout,
	})
	return err == nil
}
