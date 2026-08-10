package mirror

import (
	"net/url"
	"strings"
)

// ResolveLocalPath maps a URL to a local file path under hostDir.
func ResolveLocalPath(hostDir, rawURL string) (string, error) {
	return "", nil
}

// IsRejected returns true if the URL ends with any of the given suffixes (case-insensitive).
func IsRejected(rawURL string, suffixes []string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	lowerPath := strings.ToLower(parsedURL.Path)

	for _, suffix := range suffixes {
		if strings.HasSuffix(lowerPath, suffix) {
			return true
		}
	}

	return false
}

// IsExcluded returns true if the URL path starts with any of the given prefixes.
func IsExcluded(rawURL string, prefixes []string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(parsedURL.Path, prefix) {
			if len(parsedURL.Path) == len(prefix) || parsedURL.Path[len(prefix)] == '/' {
				return true
			}
		}
	}
	return false
}
