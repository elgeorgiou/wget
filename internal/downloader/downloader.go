package downloader

import (
	"errors"
	"io"
	"net/http"
)

type DownloadConfig struct {
	URL       string
	OutputDir string
	FileName  string
	RateLimit int64
	Out       io.Writer
}

func DownloadFile(cfg DownloadConfig) (savedPath string, err error) {
	resp, err := http.Get(cfg.URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("bad status code " + resp.Status)
	}
	return "", nil
}
