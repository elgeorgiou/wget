package mirror

// ResolveLocalPath maps a URL to a local file path under hostDir.
func ResolveLocalPath(hostDir, rawURL string) (string, error) {
	return "", nil
}

// IsRejected returns true if the URL ends with any of the given suffixes (case-insensitive).
func IsRejected(rawURL string, suffixes []string) bool {
	return false
}

// IsExcluded returns true if the URL path starts with any of the given prefixes.
func IsExcluded(rawURL string, prefixes []string) bool {
	return false
}
