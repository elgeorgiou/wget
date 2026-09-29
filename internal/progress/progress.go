package progress

import (
	"fmt"
	"strings"
	"time"
)

// / Render returns a formatted progress bar line for the given state.
// The caller writes it using \r to rewrite in place.
func Render(downloaded, total int64, elapsed time.Duration) string {
	downloadedKiB := float64(downloaded) / 1024
	totalKiB := float64(total) / 1024
	percentage := float64(downloaded) / float64(total) * 100
	filled := int(percentage / 100 * 20)
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", 20-filled)
	return fmt.Sprintf("%.2f KiB / %.2f KiB [%s] %.2f%%", downloadedKiB, totalKiB, bar, percentage)
}
