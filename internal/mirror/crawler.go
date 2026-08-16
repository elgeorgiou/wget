package mirror

import (
	"fmt"
	"net/url"
	"strings"

	downloader "wgetclone/internal/downloader"
)

type MirrorOptions struct {
	RejectSuffixes []string
	ExcludePaths   []string
	ConvertLinks   bool
}

func Mirror(
	rootURL string,
	opts MirrorOptions,
	dl func(downloader.DownloadConfig) (string, error),
) error {
	pendingURLs := []string{rootURL}
	visitedURLs := make(map[string]bool)

	for len(pendingURLs) > 0 {
		currentURL := pendingURLs[0]
		pendingURLs = pendingURLs[1:]

		parsedURL, err := url.Parse(currentURL)
		if err != nil {
			return fmt.Errorf("failed to parse URL %q: %w", currentURL, err)
		}

		parsedURL.Fragment = ""

		if parsedURL.Path != "/" {
			parsedURL.Path = strings.TrimRight(parsedURL.Path, "/")
		}

		currentURL = parsedURL.String()

		if visitedURLs[currentURL] {
			continue
		}

		visitedURLs[currentURL] = true

		_, err = dl(downloader.DownloadConfig{
			URL: currentURL,
		})
		if err != nil {
			return fmt.Errorf("failed to download URL %q: %w", currentURL, err)
		}
	}

	return nil
}
