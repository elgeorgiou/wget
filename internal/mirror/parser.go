// Package mirror provides functionality for discovering website resources
// and supporting recursive website downloads.
//
// It extracts resource references from downloaded files so they can be
// processed during website mirroring.
package mirror

import (
	"os"
	"regexp"
)

// ExtractLinks extracts unique resource references from a local HTML file.
//
// It reads the file and uses regular expressions to find double-quoted
// href attributes, src attributes, and CSS url() references.
// The extracted values are returned in the order they are discovered,
// with duplicate values removed.
//
// This function does not resolve relative URLs, validate destinations,
// or parse HTML using a dedicated HTML parser.
//
// Parameters:
// - path: The filesystem path of the file to examine.
//
// Returns:
// - A slice containing the unique extracted references.
// - An error if the file cannot be read, or nil on success.
func ExtractLinks(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	html := string(content)

	// Match double-quoted href, src, and CSS url() references separately.
	hrefPattern := regexp.MustCompile(`href="([^"]+)"`)
	srcPattern := regexp.MustCompile(`src="([^"]+)"`)
	cssPattern := regexp.MustCompile(`url\("([^"]+)"\)`)
	matches := hrefPattern.FindAllStringSubmatch(html, -1)
	srcmatches := srcPattern.FindAllStringSubmatch(html, -1)
	cssmatches := cssPattern.FindAllStringSubmatch(html, -1)
	var links []string
	seen := make(map[string]bool)

	// Preserve discovery order while using seen to skip duplicate references.
	for _, link := range matches {
		if seen[link[1]] {
			continue
		}
		seen[link[1]] = true
		links = append(links, link[1])
	}
	for _, link := range srcmatches {
		if seen[link[1]] {
			continue
		}
		seen[link[1]] = true
		links = append(links, link[1])
	}
	for _, link := range cssmatches {
		if seen[link[1]] {
			continue
		}
		seen[link[1]] = true
		links = append(links, link[1])
	}
	return links, nil
}
