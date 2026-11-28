// Package progress provides the progress bar line shown while a file downloads.
//
// Render builds one line with the downloaded and total size, a 48-character
// bar, the percentage, the current speed and the remaining time, for example
// " 55.05 KiB / 55.05 KiB [====...] 100.00% 1.24 MiB/s 0s". Sizes and speed
// are shown in KiB below 1 MiB and in MiB otherwise, always with two decimals.
//
// Render only returns the text. The caller writes it with a leading \r so the
// line is rewritten in place. Unknown sizes (total <= 0), zero elapsed time
// and downloaded > total are handled without panics, Inf or NaN.
package progress

import (
	"fmt"
	"strings"
	"time"
)

const (
	barWidth = 48
	kib      = 1024
	mib      = 1024 * 1024
)

// Render returns a formatted progress bar line for the given download state.
//
// The line looks like " 55.05 KiB / 55.05 KiB [====...] 100.00% 1.24 MiB/s 0s".
// Sizes are shown in KiB when total is below 1 MiB and in MiB otherwise, and
// the speed in KiB/s below 1 MiB/s and in MiB/s otherwise. When total is
// unknown, only the downloaded size and the speed are shown. The remaining
// time is "--" while the speed is still unknown. The caller writes the line
// with a leading \r so it is rewritten in place.
//
// Parameters:
//   - downloaded: The number of bytes downloaded so far.
//   - total: The total size in bytes, or 0 or less when it is unknown.
//   - elapsed: The time since the download started.
//
// Returns:
//   - The formatted progress line, without a trailing newline.
func Render(downloaded, total int64, elapsed time.Duration) string {
	if downloaded < 0 {
		downloaded = 0
	}

	var speed float64
	if elapsed > 0 {
		speed = float64(downloaded) / elapsed.Seconds()
	}
	speedText := formatSpeed(speed)

	if total <= 0 {
		size, unit := scale(downloaded, downloaded)
		return fmt.Sprintf(" %.2f %s %s", size, unit, speedText)
	}

	if downloaded > total {
		downloaded = total
	}
	percentage := float64(downloaded) / float64(total) * 100
	filled := int(percentage / 100 * barWidth)
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", barWidth-filled)

	eta := "--"
	if speed > 0 {
		eta = formatETA(float64(total-downloaded) / speed)
	}

	done, unit := scale(downloaded, total)
	size, _ := scale(total, total)
	return fmt.Sprintf(" %.2f %s / %.2f %s [%s] %.2f%% %s %s",
		done, unit, size, unit, bar, percentage, speedText, eta)
}

// scale converts n to KiB or MiB, choosing the unit from ref.
func scale(n, ref int64) (float64, string) {
	if ref < mib {
		return float64(n) / kib, "KiB"
	}
	return float64(n) / mib, "MiB"
}

// formatSpeed shows speeds below 1 MiB/s in KiB/s and faster ones in MiB/s.
func formatSpeed(bytesPerSec float64) string {
	if bytesPerSec < mib {
		return fmt.Sprintf("%.2f KiB/s", bytesPerSec/kib)
	}
	return fmt.Sprintf("%.2f MiB/s", bytesPerSec/mib)
}

// formatETA shows the remaining time as "5s" or "1m 5s".
func formatETA(seconds float64) string {
	s := int64(seconds)
	if s < 60 {
		return fmt.Sprintf("%ds", s)
	}
	return fmt.Sprintf("%dm %ds", s/60, s%60)
}
