// Package mirror provides functionality for discovering website resources
// and supporting recursive website downloads.
//
// It extracts resource references from downloaded files so they can be
// processed during website mirroring.
package mirror

import (
	"html"
	"os"
	"regexp"
	"strings"
)

// ExtractLinks extracts unique resource references from a local HTML or CSS file.
//
// It reads the file and uses regular expressions to find quoted and unquoted
// href attributes, src attributes, and CSS url() references.
// HTML entities are decoded in the extracted values.
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
	text := string(content)

	// Match href, src, and CSS url() references with or without quotes.
	attrPattern := regexp.MustCompile(`(?i)\b(?:href|src)\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'>]+))`)
	cssPattern := regexp.MustCompile(`(?i)url\(\s*["']?([^"')\s]+)["']?\s*\)`)

	matches := attrPattern.FindAllStringSubmatch(text, -1)
	cssmatches := cssPattern.FindAllStringSubmatch(text, -1)

	var links []string
	seen := make(map[string]bool)

	// Preserve discovery order while using seen to skip duplicate references.
	for _, link := range matches {
		var value string

		for _, group := range link[1:] {
			if group != "" {
				value = group
				break
			}
		}

		value = strings.TrimSpace(html.UnescapeString(value))

		if value == "" || seen[value] {
			continue
		}

		seen[value] = true
		links = append(links, value)
	}

	for _, link := range cssmatches {
		value := strings.TrimSpace(html.UnescapeString(link[1]))

		if value == "" || seen[value] {
			continue
		}

		seen[value] = true
		links = append(links, value)
	}

	return links, nil
}
