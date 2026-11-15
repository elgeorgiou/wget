package mirror

import (
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// RewriteAll walks hostDir and rewrites every link in .html, .htm and .css
// files that points to the mirrored host (absolute or root-relative) into a
// path relative to the file that contains it, so the site works offline.
// Targets are resolved with ResolveLocalPath, the same mapping the crawler
// uses to save files. Links to other hosts, mailto:, data:, fragment-only
// and already-relative links are left unchanged. hostDir's base name must be
// the mirrored host, e.g. "www.example.com". Files without changes are not
// rewritten.
func RewriteAll(hostDir string) error {
	host := filepath.Base(hostDir)
	attrLinkPattern := regexp.MustCompile(`(?i)(\s(?:href|src)\s*=\s*)(["'])([^"']*)(["'])`)
	cssURLPattern := regexp.MustCompile(`(?i)(url\(\s*)(["']?)([^"')\s]+)(["']?)(\s*\))`)

	return filepath.Walk(hostDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(filePath)) {
		case ".html", ".htm", ".css":
		default:
			return nil
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		original := string(data)
		rewrite := func(link string) string {
			return localLink(link, hostDir, host, filePath)
		}
		text := replaceLinks(attrLinkPattern, original, rewrite)
		text = replaceLinks(cssURLPattern, text, rewrite)

		if text == original {
			return nil
		}
		return os.WriteFile(filePath, []byte(text), info.Mode())
	})
}

// replaceLinks rewrites group 3 (the link) of every match of re, keeping the
// surrounding groups such as quotes and attribute names as they were.
func replaceLinks(re *regexp.Regexp, text string, rewrite func(string) string) string {
	return re.ReplaceAllStringFunc(text, func(match string) string {
		groups := re.FindStringSubmatch(match)
		groups[3] = rewrite(groups[3])
		return strings.Join(groups[1:], "")
	})
}

// localLink returns link as a path relative to the directory of fromFile when
// it points to host, or link unchanged otherwise.
func localLink(link, hostDir, host, fromFile string) string {
	parsed, err := url.Parse(link)
	if err != nil {
		return link
	}
	switch {
	case parsed.Scheme != "" && parsed.Scheme != "http" && parsed.Scheme != "https":
		return link
	case parsed.Host != "" && parsed.Host != host:
		return link
	case parsed.Host == "" && !strings.HasPrefix(parsed.Path, "/"):
		// fragment-only or already relative
		return link
	}

	target, err := ResolveLocalPath(hostDir, (&url.URL{Scheme: "http", Host: host, Path: parsed.Path}).String())
	if err != nil {
		return link
	}
	rel, err := filepath.Rel(filepath.Dir(fromFile), target)
	if err != nil {
		return link
	}
	rel = filepath.ToSlash(rel)
	if parsed.Fragment != "" {
		rel += "#" + parsed.Fragment
	}
	return rel
}
