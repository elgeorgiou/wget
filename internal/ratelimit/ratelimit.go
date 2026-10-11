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

		expected := time.Duration(r.totalBytes) * time.Second /
			time.Duration(r.bytesPerSec)

		elapsed := time.Since(r.start)

		if elapsed < expected {
			time.Sleep(expected - elapsed)
		}
	}

	return n, err
}
