package downloader

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	filename := ""
	if cfg.FileName != "" {
		filename = cfg.FileName
	}
	if filename == "" {
		parsedURL, err := url.Parse(cfg.URL)
		if err != nil {
			return "", err
		}
		filename = filepath.Base(parsedURL.Path)
	}
	if cfg.OutputDir != "" {
		err = os.MkdirAll(cfg.OutputDir, 0755)
		if err != nil {
			return "", err
		}

	}
	savedPath = filepath.Join(cfg.OutputDir, filename)
	file, err := os.Create(savedPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}
	return savedPath, nil
}
