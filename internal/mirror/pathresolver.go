package mirror

import (
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

// ResolveLocalPath maps a URL to a local file path under hostDir. URLs whose
// path has a file extension keep their path; the site root, URLs ending in
// "/" and URLs without an extension map to an index.html inside the matching
// folder. ".." segments are
// cleaned so the result never leaves hostDir. It returns an error if rawURL
// cannot be parsed.
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

// IsRejected returns true if the URL path ends with any of the given file
// suffixes (case-insensitive). A suffix may be given with or without the
// leading dot ("jpg" or ".jpg"); empty suffixes are ignored.
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

// IsExcluded returns true if the URL path is inside any of the given
// directories. A prefix matches whole path segments only, so "/css" excludes
// "/css/a.css" but not "/csslib/". Leading and trailing slashes are optional;
// empty prefixes are ignored.
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
