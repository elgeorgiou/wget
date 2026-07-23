package progress

import "time"

// / Render returns a formatted progress bar line for the given state.
// The caller writes it using \r to rewrite in place.
func Render(downloaded, total int64, elapsed time.Duration) string {
	return ""
}
