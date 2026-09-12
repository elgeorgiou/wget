package progress

import (
	"fmt"
	"time"
)

// / Render returns a formatted progress bar line for the given state.
// The caller writes it using \r to rewrite in place.
func Render(downloaded, total int64, elapsed time.Duration) string {
	downloadedKiB := float64(downloaded) / 1024
	totalKiB := float64(total) / 1024
	return fmt.Sprintf("%.2f KiB / %.2f KiB", downloadedKiB, totalKiB)
}
