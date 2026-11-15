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

// Render returns a formatted progress bar line for the given state, for
// example " 55.05 KiB / 55.05 KiB [====...] 100.00% 1.24 MiB/s 0s".
// The caller writes it using \r to rewrite in place. Sizes are shown in KiB
// when total is below 1 MiB and in MiB otherwise. When total is unknown
// (total <= 0) only the downloaded size and speed are shown. The remaining
// time is "--" while the speed is still unknown.
func Render(downloaded, total int64, elapsed time.Duration) string {
	if downloaded < 0 {
		downloaded = 0
	}

	var speed float64
	if elapsed > 0 {
		speed = float64(downloaded) / elapsed.Seconds()
	}
	speedText := fmt.Sprintf("%.2f MiB/s", speed/mib)

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

func formatETA(seconds float64) string {
	s := int64(seconds)
	if s < 60 {
		return fmt.Sprintf("%ds", s)
	}
	return fmt.Sprintf("%dm %ds", s/60, s%60)
}
