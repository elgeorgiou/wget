package mirror

import (
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

// ResolveLocalPath maps a URL to the local file path where it is saved.
//
// URLs whose path has a file extension keep their path. The site root, URLs
// ending in "/" and URLs without an extension map to an index.html inside the
// matching folder. ".." segments are cleaned so the result never leaves
// hostDir. For example:
//
//	https://www.example.com/js/app.js  ->  www.example.com/js/app.js
//	https://www.example.com/           ->  www.example.com/index.html
//	https://www.example.com/about      ->  www.example.com/about/index.html
//
// Parameters:
//   - hostDir: The folder of the mirrored site (e.g., "www.example.com").
//   - rawURL: The URL to map.
//
// Returns:
//   - The local file path under hostDir.
//   - An error if rawURL cannot be parsed.
func ResolveLocalPath(hostDir, rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	// Cleaning a rooted path drops any ".." that would climb above "/".
	cleanPath := path.Clean("/" + parsedURL.Path)
	if cleanPath == "/" {
		return filepath.Join(hostDir, "index.html"), nil
	}
	// path.Clean drops the trailing slash, so check it on the original path.
	isDir := strings.HasSuffix(parsedURL.Path, "/")
	if !isDir && path.Ext(cleanPath) != "" {
		return filepath.Join(hostDir, filepath.FromSlash(cleanPath)), nil
	}
	return filepath.Join(hostDir, filepath.FromSlash(cleanPath), "index.html"), nil
}

// IsRejected reports whether the URL path ends with one of the given file
// suffixes, for the -R flag.
//
// The match is case-insensitive and ignores the query string. A suffix may be
// given with or without the leading dot ("jpg" or ".jpg"). Empty suffixes are
// ignored, so a value like "-R=jpg," does not reject every URL.
//
// Parameters:
//   - rawURL: The URL to check.
//   - suffixes: The file suffixes to reject (e.g., []string{"jpg", "gif"}).
//
// Returns:
//   - true if the URL must be skipped, false otherwise or if rawURL cannot be
//     parsed.
func IsRejected(rawURL string, suffixes []string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	lowerPath := strings.ToLower(parsedURL.Path)

	for _, suffix := range suffixes {
		suffix = strings.ToLower(strings.TrimSpace(suffix))
		if suffix == "" {
			continue
		}
		if !strings.HasPrefix(suffix, ".") {
			suffix = "." + suffix
		}
		if strings.HasSuffix(lowerPath, suffix) {
			return true
		}
	}
	return false
}

// IsExcluded reports whether the URL path is inside one of the given
// directories, for the -X flag.
//
// A prefix matches whole path segments only, so "/css" excludes "/css/a.css"
// but not "/csslib/". Leading and trailing slashes are optional ("/js", "js"
// and "/js/" are the same). Empty prefixes are ignored.
//
// Parameters:
//   - rawURL: The URL to check.
//   - prefixes: The directories to exclude (e.g., []string{"/js", "/assets"}).
//
// Returns:
//   - true if the URL must be skipped, false otherwise or if rawURL cannot be
//     parsed.
func IsExcluded(rawURL string, prefixes []string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	for _, prefix := range prefixes {
		dir := strings.Trim(strings.TrimSpace(prefix), "/")
		if dir == "" {
			continue
		}
		dir = "/" + dir
		if parsedURL.Path == dir || strings.HasPrefix(parsedURL.Path, dir+"/") {
			return true
		}
	}
	return false
}
