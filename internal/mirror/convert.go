package mirror

import (
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// RewriteAll walks hostDir and rewrites every link in .html, .htm and .css
// files that points to a downloaded file of the mirrored host (absolute,
// root-relative or relative) into a path relative to the file that contains
// it, so the site works offline. Directory links become their index.html.
// Targets are resolved with ResolveLocalPath, the same mapping the crawler
// uses to save files. Links to other hosts, mailto:, data:, fragment-only
// links and links to files that were not downloaded are left unchanged.
// hostDir's base name must be the mirrored host, e.g. "www.example.com".
// Files without changes are not rewritten.
func RewriteAll(hostDir string) error {
	host := filepath.Base(hostDir)
	attrLinkPattern := regexp.MustCompile(`(?i)(\s(?:href|src)\s*=\s*)(["'])([^"']*)(["'])`)
	unquotedAttrPattern := regexp.MustCompile(`(?i)(\s(?:href|src)\s*=\s*)()([^\s"'>]+)()`)
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
		text = replaceLinks(unquotedAttrPattern, text, rewrite)
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
// it points to a file of host that exists under hostDir, or link unchanged
// otherwise. Relative links are resolved against the location of fromFile,
// the same way the crawler resolved them against the page URL.
func localLink(link, hostDir, host, fromFile string) string {
	if link == "" || strings.HasPrefix(link, "#") {
		return link
	}
	ref, err := url.Parse(link)
	if err != nil {
		return link
	}
	fromRel, err := filepath.Rel(hostDir, fromFile)
	if err != nil {
		return link
	}
	base := &url.URL{Scheme: "http", Host: host, Path: "/" + filepath.ToSlash(fromRel)}
	abs := base.ResolveReference(ref)
	if (abs.Scheme != "http" && abs.Scheme != "https") || abs.Host != host {
		return link
	}

	target, err := ResolveLocalPath(hostDir, abs.String())
	if err != nil {
		return link
	}
	// Files that were not downloaded (rejected, excluded or failed) keep
	// their original link, so absolute ones still load online.
	if _, err := os.Stat(target); err != nil {
		return link
	}
	rel, err := filepath.Rel(filepath.Dir(fromFile), target)
	if err != nil {
		return link
	}
	// Escape the path so names containing ' ', '?', '#' or ':' still point to
	// the saved file instead of being read as a query, fragment or scheme.
	relURL := &url.URL{Path: filepath.ToSlash(rel), RawFragment: ref.RawFragment, Fragment: ref.Fragment}
	return relURL.String()
}
