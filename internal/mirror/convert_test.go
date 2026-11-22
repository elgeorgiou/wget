package mirror

import (
	"os"
	"path/filepath"
	"testing"
)

// siteFiles are the files every test site starts with, as the crawler would
// have saved them. Links to anything else point to files that were not
// downloaded.
var siteFiles = []string{
	"index.html",
	"about.html",
	"about/index.html",
	"blog/index.html",
	"blog/post.html",
	"css/a.css",
	"a.css",
	"img/a.png",
	"img/bg.png",
	"fonts/a.woff2",
	"a b.html",
	"q?.html",
	"c#.html",
	"a:b.html",
}

// newSite creates siteFiles under a fresh example.com hostDir and returns it.
func newSite(t *testing.T) string {
	t.Helper()
	hostDir := filepath.Join(t.TempDir(), "example.com")
	for _, rel := range siteFiles {
		writeFile(t, filepath.Join(hostDir, filepath.FromSlash(rel)), "")
	}
	return hostDir
}

func writeFile(t *testing.T, filePath, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("could not create dir: %v", err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("could not write file: %v", err)
	}
}

// rewriteFile writes content to rel inside a new site, runs RewriteAll and
// returns the rewritten content.
func rewriteFile(t *testing.T, rel, content string) string {
	t.Helper()
	hostDir := newSite(t)
	filePath := filepath.Join(hostDir, filepath.FromSlash(rel))
	writeFile(t, filePath, content)
	if err := RewriteAll(hostDir); err != nil {
		t.Fatalf("RewriteAll returned an error: %v", err)
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	return string(data)
}

func TestRewriteAllAbsoluteHref(t *testing.T) {
	got := rewriteFile(t, "page.html", `<a href="http://example.com/about.html">about</a>`)
	expected := `<a href="about.html">about</a>`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestRewriteAllNonHTMLSkipped(t *testing.T) {
	content := `<a href="http://example.com/about.html">about</a>`
	got := rewriteFile(t, "notes.txt", content)
	if got != content {
		t.Errorf("expected file to stay unchanged %q, got %q", content, got)
	}
}

func TestRewriteAllRootRelativeHref(t *testing.T) {
	got := rewriteFile(t, "page.html", `<a href="/about.html">about</a>`)
	expected := `<a href="about.html">about</a>`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestRewriteAllLinks(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		content string
		want    string
	}{
		{"page in subfolder", "blog/post/index.html",
			`<link href="/css/a.css">`, `<link href="../../css/a.css">`},
		{"absolute url in subfolder", "blog/post/index.html",
			`<img src="https://example.com/img/a.png">`, `<img src="../../img/a.png">`},
		{"root link", "index.html",
			`<a href="/">home</a>`, `<a href="index.html">home</a>`},
		{"root link from subfolder", "blog/index.html",
			`<a href="/">home</a>`, `<a href="../index.html">home</a>`},
		{"host without trailing slash", "index.html",
			`<a href="http://example.com">home</a>`, `<a href="index.html">home</a>`},
		{"directory link", "index.html",
			`<a href="/about">about</a>`, `<a href="about/index.html">about</a>`},
		{"directory link with slash", "index.html",
			`<a href="/about/">about</a>`, `<a href="about/index.html">about</a>`},
		{"fragment is kept", "index.html",
			`<a href="/about.html#team">team</a>`, `<a href="about.html#team">team</a>`},
		{"query is dropped", "index.html",
			`<link href="/a.css?v=2">`, `<link href="a.css">`},
		{"single quotes", "index.html",
			`<a href='/about.html'>about</a>`, `<a href='about.html'>about</a>`},
		{"spaces around equals", "index.html",
			`<img src = "/img/a.png">`, `<img src = "img/a.png">`},
		{"unquoted value", "index.html",
			`<a href=/about>about</a>`, `<a href=about/index.html>about</a>`},
		{"unquoted value before closing bracket", "blog/index.html",
			`<img src=/img/a.png>`, `<img src=../img/a.png>`},
		{"other host unchanged", "index.html",
			`<a href="https://other.com/x.html">x</a>`, `<a href="https://other.com/x.html">x</a>`},
		{"protocol-relative other host unchanged", "index.html",
			`<img src="//cdn.com/x.png">`, `<img src="//cdn.com/x.png">`},
		{"mailto unchanged", "index.html",
			`<a href="mailto:a@b.com">mail</a>`, `<a href="mailto:a@b.com">mail</a>`},
		{"data uri unchanged", "index.html",
			`<img src="data:image/png;base64,AAAA">`, `<img src="data:image/png;base64,AAAA">`},
		{"fragment only unchanged", "index.html",
			`<a href="#top">top</a>`, `<a href="#top">top</a>`},
		{"empty link unchanged", "index.html",
			`<a href="">self</a>`, `<a href="">self</a>`},
		{"relative file link unchanged", "blog/index.html",
			`<a href="post.html">post</a>`, `<a href="post.html">post</a>`},
		{"relative directory link", "index.html",
			`<a href="about/">about</a>`, `<a href="about/index.html">about</a>`},
		{"relative directory link without slash", "index.html",
			`<a href="about">about</a>`, `<a href="about/index.html">about</a>`},
		{"dot link", "blog/index.html",
			`<a href="./">blog</a>`, `<a href="index.html">blog</a>`},
		{"parent link", "blog/index.html",
			`<a href="../">home</a>`, `<a href="../index.html">home</a>`},
		{"relative link with query", "index.html",
			`<link href="a.css?v=2">`, `<link href="a.css">`},
		{"missing absolute target unchanged", "index.html",
			`<img src="https://example.com/img/skipped.gif">`, `<img src="https://example.com/img/skipped.gif">`},
		{"missing root-relative target unchanged", "index.html",
			`<script src="/js/app.js"></script>`, `<script src="/js/app.js"></script>`},
		{"missing directory target unchanged", "index.html",
			`<a href="/contact">contact</a>`, `<a href="/contact">contact</a>`},
		{"escaped space stays escaped", "index.html",
			`<a href="/a%20b.html">a</a>`, `<a href="a%20b.html">a</a>`},
		{"escaped question mark stays escaped", "index.html",
			`<a href="/q%3F.html">q</a>`, `<a href="q%3F.html">q</a>`},
		{"escaped hash stays escaped", "index.html",
			`<a href="/c%23.html#top">c</a>`, `<a href="c%23.html#top">c</a>`},
		{"colon in first segment", "index.html",
			`<a href="/a:b.html">a</a>`, `<a href="./a:b.html">a</a>`},
		{"htm uppercase extension", "PAGE.HTM",
			`<a href="/about.html">about</a>`, `<a href="about.html">about</a>`},
		{"css url unquoted", "css/style.css",
			`body{background:url(/img/bg.png)}`, `body{background:url(../img/bg.png)}`},
		{"css url quoted", "css/style.css",
			`@font-face{src:url('/fonts/a.woff2')}`, `@font-face{src:url('../fonts/a.woff2')}`},
		{"css url relative", "css/style.css",
			`body{background:url(../img/bg.png)}`, `body{background:url(../img/bg.png)}`},
		{"css url in html style", "index.html",
			`<div style="background:url(https://example.com/img/bg.png)">`, `<div style="background:url(img/bg.png)">`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rewriteFile(t, tt.file, tt.content); got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
