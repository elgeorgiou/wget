package mirror

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"wgetclone/internal/downloader"
)

// MirrorOptions holds the filtering and link conversion options for mirroring.
type MirrorOptions struct {
	RejectSuffixes []string
	ExcludePaths   []string
	ConvertLinks   bool
	Out            io.Writer // NEW: log destination; nil discards output
}

// Mirror crawls rootURL and saves reachable same-domain resources to disk.
// The injected dl function downloads each resource and allows testing
// without making real HTTP requests.
func Mirror(
	rootURL string,
	opts MirrorOptions,
	dl func(downloader.DownloadConfig) (string, error),
) error {
	parsedRootURL, err := url.Parse(rootURL)
	if err != nil {
		return fmt.Errorf("failed to parse root URL %q: %w", rootURL, err)
	}

	if (parsedRootURL.Scheme != "http" && parsedRootURL.Scheme != "https") ||
		parsedRootURL.Hostname() == "" {
		return fmt.Errorf("invalid HTTP or HTTPS root URL %q", rootURL)
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

	// NEW: a nil writer must not crash the downloader.
	out := opts.Out
	if out == nil {
		out = io.Discard
	}

	pendingURLs := []string{rootURL}
	visitedURLs := make(map[string]bool)

	for len(pendingURLs) > 0 {
		currentURL := pendingURLs[0]
		pendingURLs = pendingURLs[1:]

		isRoot := len(visitedURLs) == 0

		parsedURL, err := url.Parse(currentURL)
		if err != nil {
			if isRoot {
				return fmt.Errorf("failed to parse URL %q: %w", currentURL, err)
			}
			continue
		}

		parsedURL.Fragment = ""
		parsedURL.Host = strings.ToLower(parsedURL.Host) // NEW
		currentURL = parsedURL.String()

		normalizedURL := *parsedURL
		normalizedURL.Scheme = "https" // NEW: http and https share one visited key
		normalizedURL.Path = strings.TrimRight(normalizedURL.Path, "/")

		if normalizedURL.Path == "" {
			normalizedURL.Path = "/"
		}
		normalizedURL.RawPath = ""
		visitedKey := normalizedURL.String()

		if visitedURLs[visitedKey] {
			continue
		}

		if IsRejected(currentURL, opts.RejectSuffixes) {
			continue
		}

		if IsExcluded(currentURL, opts.ExcludePaths) {
			continue
		}

		visitedURLs[visitedKey] = true

		localPath, err := ResolveLocalPath(domainDirectory, currentURL)
		if err != nil {
			if isRoot {
				return fmt.Errorf("failed to resolve local path for %q: %w", currentURL, err)
			}
			continue
		}

		// NEW: never write outside the host directory (e.g. /%2e%2e/x).
		rel, err := filepath.Rel(domainDirectory, localPath)
		if err != nil || rel == ".." ||
			strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}

		savedPath, err := dl(downloader.DownloadConfig{
			URL:       currentURL,
			OutputDir: filepath.Dir(localPath),
			FileName:  filepath.Base(localPath),
			Out:       out, // NEW: was io.Discard
		})

		if err != nil {
			if isRoot {
				return fmt.Errorf("failed to download URL %q: %w", currentURL, err)
			}
			fmt.Fprintf(out, "skipping %s: %v\n", currentURL, err) // NEW
			continue
		}

		// NEW: replaces the extension-only check.
		if !shouldParse(parsedURL.Path, savedPath) {
			continue
		}

		links, err := ExtractLinks(savedPath)
		if err != nil {
			if isRoot {
				return fmt.Errorf("failed to extract links from %q: %w", savedPath, err)
			}
			continue
		}

		for _, link := range links {
			linkURL, err := url.Parse(link)
			if err != nil {
				continue
			}

			absoluteURL := parsedURL.ResolveReference(linkURL)

			if absoluteURL.Scheme != "http" && absoluteURL.Scheme != "https" {
				continue
			}

			// NEW: EqualFold instead of !=
			if !strings.EqualFold(absoluteURL.Host, parsedRootURL.Host) {
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

// shouldParse reports whether the saved file may contain links. Known HTML/CSS
// extensions and extensionless paths are parsed; any other extension is parsed
// only if the file content looks like HTML (e.g. page.php).
func shouldParse(urlPath, savedPath string) bool {
	switch strings.ToLower(filepath.Ext(urlPath)) {
	case "", ".html", ".htm", ".css":
		return true
	}

	file, err := os.Open(savedPath)
	if err != nil {
		return false
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, _ := io.ReadFull(file, buf)

	return strings.HasPrefix(http.DetectContentType(buf[:n]), "text/html")
}
