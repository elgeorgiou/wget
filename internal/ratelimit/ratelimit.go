package ratelimit

import (
	"io"
)

func NewReader(r io.Reader, bytesPerSec int64) io.Reader {

	return r
}
