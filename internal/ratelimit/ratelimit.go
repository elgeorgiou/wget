package ratelimit

import (
	"io"
	"time"
)

type reader struct {
	r           io.Reader
	bytesPerSec int64
}

func NewReader(r io.Reader, bytesPerSec int64) io.Reader {

	return &reader{
		r:           r,
		bytesPerSec: bytesPerSec,
	}
}

func (r *reader) Read(p []byte) (int, error) {
	start := time.Now()

	n, err := r.r.Read(p)

	if n > 0 && r.bytesPerSec > 0 {
		expected := time.Duration(int64(n)) * time.Second / time.Duration(r.bytesPerSec)
		elapsed := time.Since(start)

		if elapsed < expected {
			time.Sleep(expected - elapsed)
		}
	}

	return n, err
}
