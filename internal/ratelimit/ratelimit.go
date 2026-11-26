// Package ratelimit provides functionality for limiting the reading rate
// of an io.Reader.
//
// The package controls the reading rate by tracking the total number of
// bytes read and introducing delays when necessary.
//
// A non-positive rate disables throttling.
package ratelimit

import (
	"io"
	"time"
)

// reader wraps an io.Reader and tracks the cumulative number of bytes read
// and the starting time to enforce the configured reading rate.
type reader struct {
	r           io.Reader
	bytesPerSec int64
	totalBytes  int64
	start       time.Time
}

// NewReader wraps an io.Reader and limits its reading rate.
//
// A non-positive rate disables throttling.
//
// Parameters:
//   - r: The underlying reader from which data is read.
//   - bytesPerSec: The maximum reading rate in bytes per second.
//
// Returns:
//   - An io.Reader that applies the configured rate limit.
func NewReader(r io.Reader, bytesPerSec int64) io.Reader {
	return &reader{
		r:           r,
		bytesPerSec: bytesPerSec,
	}
}

// Read reads data from the underlying reader while enforcing the configured
// reading rate.
//
// Each read is limited to a maximum chunk size of 32 KiB. When throttling
// is enabled, the method calculates the expected elapsed time based on
// the cumulative number of bytes read and waits when necessary.
//
// Parameters:
//   - p: The buffer into which the data is read.
//
// Returns:
//   - The number of bytes read into the buffer.
//   - The error returned by the underlying reader, if any.
func (r *reader) Read(p []byte) (int, error) {

	const maxChunkSize = 32 * 1024

	if len(p) > maxChunkSize {
		p = p[:maxChunkSize]
	}

	if r.bytesPerSec > 0 && r.start.IsZero() {
		r.start = time.Now()
	}

	n, err := r.r.Read(p)

	if n > 0 && r.bytesPerSec > 0 {

		r.totalBytes += int64(n)

		// Calculate the expected elapsed time from the total bytes read,
		// preventing small timing errors from accumulating across reads.
		expectedSeconds := float64(r.totalBytes) / float64(r.bytesPerSec)

		// Cap the expected duration to prevent time.Duration overflow.
		const maxDuration = time.Duration(1<<63 - 1)
		maxSeconds := float64(maxDuration) / float64(time.Second)

		var expected time.Duration

		if expectedSeconds >= maxSeconds {
			expected = maxDuration
		} else {
			expected = time.Duration(expectedSeconds * float64(time.Second))
		}

		elapsed := time.Since(r.start)

		if elapsed < expected {
			time.Sleep(expected - elapsed)
		}
	}

	return n, err
}
