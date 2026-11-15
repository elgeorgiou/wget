package mirror

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRewriteAllAbsoluteHref(t *testing.T) {
	hostDir := filepath.Join(t.TempDir(), "example.com")
	err := os.MkdirAll(hostDir, 0o755)
	if err != nil {
		t.Fatalf("could not create hostDir: %v", err)
	}

	pagePath := filepath.Join(hostDir, "page.html")
	content := `<a href="http://example.com/about.html">about</a>`
	err = os.WriteFile(pagePath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("could not write page: %v", err)
	}

	err = RewriteAll(hostDir)
	if err != nil {
		t.Fatalf("RewriteAll returned an error: %v", err)
	}

	data, err := os.ReadFile(pagePath)
	if err != nil {
		t.Fatalf("could not read page: %v", err)
	}

	expected := `<a href="about.html">about</a>`
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestRewriteAllNonHTMLSkipped(t *testing.T) {
	hostDir := filepath.Join(t.TempDir(), "example.com")
	err := os.MkdirAll(hostDir, 0o755)
	if err != nil {
		t.Fatalf("could not create hostDir: %v", err)
	}

	filePath := filepath.Join(hostDir, "notes.txt")
	content := `<a href="http://example.com/about.html">about</a>`
	err = os.WriteFile(filePath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("could not write file: %v", err)
	}

	err = RewriteAll(hostDir)
	if err != nil {
		t.Fatalf("RewriteAll returned an error: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}

	if string(data) != content {
		t.Errorf("expected file to stay unchanged %q, got %q", content, string(data))
	}
}

func TestRewriteAllRootRelativeHref(t *testing.T) {
	hostDir := filepath.Join(t.TempDir(), "example.com")
	err := os.MkdirAll(hostDir, 0o755)
	if err != nil {
		t.Fatalf("could not create hostDir: %v", err)
	}

	pagePath := filepath.Join(hostDir, "page.html")
	content := `<a href="/about.html">about</a>`
	err = os.WriteFile(pagePath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("could not write page: %v", err)
	}

	err = RewriteAll(hostDir)
	if err != nil {
		t.Fatalf("RewriteAll returned an error: %v", err)
	}

	data, err := os.ReadFile(pagePath)
	if err != nil {
		t.Fatalf("could not read page: %v", err)
	}

	expected := `<a href="about.html">about</a>`
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

// rewriteFile writes content to rel under a fresh example.com hostDir,
// runs RewriteAll and returns the rewritten content.
func rewriteFile(t *testing.T, rel, content string) string {
	t.Helper()
	hostDir := filepath.Join(t.TempDir(), "example.com")
	filePath := filepath.Join(hostDir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("could not create dir: %v", err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("could not write file: %v", err)
	}
	if err := RewriteAll(hostDir); err != nil {
		t.Fatalf("RewriteAll returned an error: %v", err)
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	return string(data)
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
		{"relative link unchanged", "blog/index.html",
			`<a href="post.html">post</a>`, `<a href="post.html">post</a>`},
		{"htm uppercase extension", "PAGE.HTM",
			`<a href="/about.html">about</a>`, `<a href="about.html">about</a>`},
		{"css url unquoted", "css/style.css",
			`body{background:url(/img/bg.png)}`, `body{background:url(../img/bg.png)}`},
		{"css url quoted", "css/style.css",
			`@font-face{src:url('/fonts/a.woff2')}`, `@font-face{src:url('../fonts/a.woff2')}`},
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
