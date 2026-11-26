// Package mirror provides functionality for downloading and mirroring websites.
//
// The package crawls reachable resources within the same domain, applies URL
// filtering rules, and saves downloaded files to a local directory structure.
//
// It also supports converting downloaded links for offline navigation.
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

// MirrorOptions defines the filtering, link conversion, and logging options
// for website mirroring.
type MirrorOptions struct {
	// RejectSuffixes specifies the file suffixes to exclude from mirroring.
	RejectSuffixes []string

	// ExcludePaths specifies the URL paths to exclude from mirroring.
	ExcludePaths []string

	// ConvertLinks enables link conversion for offline navigation.
	ConvertLinks bool

	// Out specifies the logging destination. A nil writer discards output.
	Out io.Writer
}

// Mirror crawls a website and saves reachable same-domain resources to disk.
//
// The function processes URLs using a queue, normalizes visited URLs to avoid
// duplicate downloads, and applies suffix and path filtering rules.
//
// Resources are saved within the domain directory. Errors encountered while
// processing subsequent resources are skipped where applicable, allowing the
// crawl to continue.
//
// The injected download function allows testing without making real HTTP
// requests. Downloaded links are converted for offline navigation when enabled.
//
// Parameters:
//   - rootURL: The starting HTTP or HTTPS URL of the website.
//   - opts: The filtering, link conversion, and logging options.
//   - dl: The download function used to retrieve and save each resource.
//
// Returns:
//   - An error if root URL validation, directory creation, root resource
//     processing, or link conversion fails.
//   - Nil if the mirroring process completes without a fatal error.
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

	// Provide a valid writer even when logging is disabled.
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
		parsedURL.Host = strings.ToLower(parsedURL.Host)
		currentURL = parsedURL.String()

		normalizedURL := *parsedURL
		normalizedURL.Scheme = "https" // HTTP and HTTPS share the same visited key.
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

		// Prevent path traversal outside the domain directory.
		rel, err := filepath.Rel(domainDirectory, localPath)
		if err != nil || rel == ".." ||
			strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}

		savedPath, err := dl(downloader.DownloadConfig{
			URL:       currentURL,
			OutputDir: filepath.Dir(localPath),
			FileName:  filepath.Base(localPath),
			Out:       out,
		})

		if err != nil {
			if isRoot {
				return fmt.Errorf("failed to download URL %q: %w", currentURL, err)
			}
			fmt.Fprintf(out, "skipping %s: %v\n", currentURL, err)
			continue
		}

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

// shouldParse reports whether a downloaded resource may contain links.
//
// Resources with HTML or CSS extensions and extensionless URL paths are
// selected for parsing directly. For other extensions, the first 512 bytes
// of the saved file are inspected to determine whether it contains HTML.
//
// Parameters:
//   - urlPath: The path component of the resource URL.
//   - savedPath: The local path of the downloaded resource.
//
// Returns:
//   - True if the resource should be parsed for links.
//   - False if the resource is not recognized as HTML or cannot be opened
//     when content inspection is required.
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
