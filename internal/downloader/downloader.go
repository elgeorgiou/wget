package downloader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"wgetclone/internal/progress"
	"wgetclone/internal/ratelimit"
)

type DownloadConfig struct {
	URL        string
	OutputDir  string
	FileName   string
	RateLimit  int64
	Out        io.Writer
	NoProgress bool
}

func DownloadFile(cfg DownloadConfig) (savedPath string, err error) {
	if cfg.Out != nil {
		start := time.Now().Format("2006-01-02 15:04:05")
		fmt.Fprintf(cfg.Out, "start at %s\n", start)
	}
	resp, err := http.Get(cfg.URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if cfg.Out != nil {
		fmt.Fprintf(cfg.Out, "sending request, awaiting response... status %s\n", resp.Status)
		if resp.ContentLength >= 0 {
			decimal := float64(resp.ContentLength)
			mb := decimal / 1000000
			fmt.Fprintf(cfg.Out, "content size: %d [~%.2fMB]\n", resp.ContentLength, mb)
		} else {
			fmt.Fprintf(cfg.Out, "content size: unknown\n")
		}
	}
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
		name := filepath.Base(parsedURL.Path)
		if name == "." || name == "/" || name == "" {
			name = "index.html"
		}
		filename = name
	}
	originalDir := cfg.OutputDir
	if strings.HasPrefix(cfg.OutputDir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		cfg.OutputDir = filepath.Join(home, cfg.OutputDir[2:])
	}
	if cfg.OutputDir != "" {
		err = os.MkdirAll(cfg.OutputDir, 0755)
		if err != nil {
			return "", err
		}
	}
	savedPath = filepath.Join(cfg.OutputDir, filename)

	display := filepath.Join(originalDir, filename)
	if originalDir == "" {
		display = "./" + filename
	}
	if cfg.Out != nil {
		fmt.Fprintf(cfg.Out, "saving file to: %s\n", display)
	}
	file, err := os.Create(savedPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	buffer := make([]byte, 64*1024)
	var downloaded int64

	total := resp.ContentLength
	start := time.Now()
	var reader io.Reader = resp.Body
	if cfg.RateLimit > 0 {
		reader = ratelimit.NewReader(reader, cfg.RateLimit)
	}
	for {
		n, err := reader.Read(buffer)

		if n > 0 {
			w, err := file.Write(buffer[:n])
			if err != nil {
				return "", err
			}
			downloaded += int64(w)
			if cfg.Out != nil && !cfg.NoProgress && total > 0 {
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
	if cfg.Out != nil {
		if !cfg.NoProgress && total > 0 && downloaded > 0 {
			fmt.Fprint(cfg.Out, "\n")
		}
		finished := time.Now().Format("2006-01-02 15:04:05")
		fmt.Fprintf(cfg.Out, "Downloaded [%s]\n", cfg.URL)
		fmt.Fprintf(cfg.Out, "finished at %s\n", finished)
	}
	return savedPath, nil
}
