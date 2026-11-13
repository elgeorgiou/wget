// Package downloader provides functionality for downloading files over HTTP.
//
// It handles HTTP requests, output file creation, download progress reporting,
// and optional transfer rate limiting.
//
// Download behavior is configured through DownloadConfig, allowing callers
// to control file destinations, logging, and progress display.
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

// DownloadConfig defines the configuration options for a single file download.
//
// Fields:
// - URL: The HTTP or HTTPS URL of the file to download.
// - OutputDir: The destination directory; an empty value uses the current directory.
// - FileName: The output filename; an empty value derives the name from the URL.
// - RateLimit: The maximum download rate in bytes per second; zero disables limiting.
// - Out: The writer for download logs and progress output; nil disables output.
// - NoProgress: Disables progress bar rendering when true.
type DownloadConfig struct {
	URL        string
	OutputDir  string
	FileName   string
	RateLimit  int64
	Out        io.Writer
	NoProgress bool
}

// DownloadFile downloads a file from the specified URL and saves it locally.
//
// The function sends an HTTP GET request, validates the response status,
// determines the output filename, and creates the destination file.
// It reads the response body in chunks, optionally limits the transfer rate,
// and reports download progress when the content length is known.
//
// If no filename is specified, the function derives one from the URL path.
// URLs without a usable filename are saved as "index.html". Output directories
// beginning with "~/" are expanded to the user's home directory.
//
// Parameters:
//   - cfg: The download configuration containing the URL, output destination,
//     rate limit, and output settings.
//
// Returns:
// - savedPath: The local path of the downloaded file.
// - err: An error if the request, path preparation, file creation, or transfer fails.
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
		return "", errors.New("status " + resp.Status)
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
	// Preserve the original directory for logging before expanding "~/".
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

	// Display the user-provided path while saving to the resolved filesystem path.
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
	// Wrap the response reader to throttle transfers when a rate limit is set.
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
			// Render progress only when enabled and the total size is known.
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
		// Separate the completed progress bar from the final download summary.
		if !cfg.NoProgress && total > 0 && downloaded > 0 {
			fmt.Fprint(cfg.Out, "\n\n")
		}
		finished := time.Now().Format("2006-01-02 15:04:05")
		fmt.Fprintf(cfg.Out, "Downloaded [%s]\n", cfg.URL)
		fmt.Fprintf(cfg.Out, "finished at %s\n", finished)
	}
	return savedPath, nil
}
