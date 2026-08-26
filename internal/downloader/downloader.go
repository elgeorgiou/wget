package downloader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
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
	buffer := make([]byte, 64*1024)
	var downloaded int64

	total := resp.ContentLength
	start := time.Now()

	for {
		n, err := resp.Body.Read(buffer)

		if n > 0 {
			w, err := file.Write(buffer[:n])
			if err != nil {
				return "", err
			}
			downloaded += int64(w)
			if cfg.Out != nil {
				elapsed := time.Since(start)
				line := progress.Render(downloaded, total, elapsed)
				fmt.Fprintf(cfg.Out, "\r%s", line)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}
	return savedPath, nil
}
