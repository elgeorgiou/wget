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
	matches := hrefPattern.FindAllStringSubmatch(html, -1)
	srcmatches := srcPattern.FindAllStringSubmatch(html, -1)
	var links []string
	for _, link := range matches {
		links = append(links, link[1])
	}
	for _, link := range srcmatches {
		links = append(links, link[1])
	}
	return links, nil
}
