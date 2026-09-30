package mirror

import (
	"os"
	"regexp"
)

func ExtractLinks(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	html := string(content)
	hrefPattern := regexp.MustCompile(`href="([^"]+)"`)
	srcPattern := regexp.MustCompile(`src="([^"]+)"`)
	cssPattern := regexp.MustCompile(`url\("([^"]+)"\)`)
	matches := hrefPattern.FindAllStringSubmatch(html, -1)
	srcmatches := srcPattern.FindAllStringSubmatch(html, -1)
	cssmatches := cssPattern.FindAllStringSubmatch(html, -1)
	var links []string
	seen := make(map[string]bool)
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
