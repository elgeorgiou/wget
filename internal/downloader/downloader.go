package downloader

import (
	"errors"
	"io"
)

type DownloadConfig struct {
	URL       string
	OutputDir string
	FileName  string
	RateLimit int64
	Out       io.Writer
}

func DownloadFile(cfg DownloadConfig) (savedPath string, err error) {
	return "", errors.New("not implemented yet")
}
