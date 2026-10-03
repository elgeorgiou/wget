package mirror

import (
	"fmt"
	"net/url"
	"os"
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
	parsedRootURL, err := url.Parse(rootURL)
	if err != nil {
		return fmt.Errorf("failed to parse root URL %q: %w", rootURL, err)
	}

	domainDirectory := parsedRootURL.Host

	err = os.MkdirAll(domainDirectory, 0755)
	if err != nil {
		return fmt.Errorf(
			"failed to create domain directory %q: %w",
			domainDirectory,
			err,
		)
	}

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

		if IsRejected(currentURL, opts.RejectSuffixes) {
			continue
		}

		if IsExcluded(currentURL, opts.ExcludePaths) {
			continue
		}

		visitedURLs[currentURL] = true

		savedPath, err := dl(downloader.DownloadConfig{
			URL:       currentURL,
			OutputDir: domainDirectory,
		})
		if err != nil {
			return fmt.Errorf("failed to download URL %q: %w", currentURL, err)
		}

		links, err := ExtractLinks(savedPath)
		if err != nil {
			return fmt.Errorf("failed to extract links from %q: %w", savedPath, err)
		}

		for _, link := range links {
			linkURL, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse link %q: %w", link, err)
			}

			absoluteURL := parsedURL.ResolveReference(linkURL)

			if absoluteURL.Host != parsedURL.Host {
				continue
			}

			pendingURLs = append(pendingURLs, absoluteURL.String())
		}
	}

	if opts.ConvertLinks {
		err := RewriteAll(domainDirectory)
		if err != nil {
			return fmt.Errorf(
				"failed to rewrite links in %q: %w",
				domainDirectory,
				err,
			)
		}
	}

	return nil
}
