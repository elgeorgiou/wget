package progress

import (
	"fmt"
	"strings"
	"time"
)

// Render returns a formatted progress bar line for the given state.
// The caller writes it using \r to rewrite in place.
func Render(downloaded, total int64, elapsed time.Duration) string {
	downloadedKiB := float64(downloaded) / 1024
	totalKiB := float64(total) / 1024
	percentage := float64(downloaded) / float64(total) * 100
	filled := int(percentage / 100 * 20)
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", 20-filled)
	bytesPerSecond := float64(downloaded) / elapsed.Seconds()
	mibPerSecond := bytesPerSecond / 1024 / 1024
	remainingSeconds := float64(total-downloaded) / bytesPerSecond
	totalSecondsInt := int(remainingSeconds)
	minutes := totalSecondsInt / 60
	seconds := totalSecondsInt % 60
	var eta string
	if minutes == 0 {
		eta = fmt.Sprintf("%ds", seconds)
	} else {
		eta = fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%.2f KiB / %.2f KiB [%s] %.2f%% %.2f MiB/s %s",
		downloadedKiB, totalKiB, bar, percentage, mibPerSecond, eta)
}
