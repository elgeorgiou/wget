package ratelimit

import (
	"io"
	"time"
)

type reader struct {
	r           io.Reader
	bytesPerSec int64
	totalBytes  int64
	start       time.Time
}

// NewReader wraps r and throttles reads to bytesPerSec.
func NewReader(r io.Reader, bytesPerSec int64) io.Reader {
	return &reader{
		r:           r,
		bytesPerSec: bytesPerSec,
	}
}

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

		expectedSeconds := float64(r.totalBytes) / float64(r.bytesPerSec)

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
