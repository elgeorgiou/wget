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

func TestRewriteAllNonHTMLSkipped(t *testing.T) { // ← νέο όνομα
	hostDir := filepath.Join(t.TempDir(), "example.com")
	err := os.MkdirAll(hostDir, 0o755)
	if err != nil {
		t.Fatalf("could not create hostDir: %v", err)
	}

	filePath := filepath.Join(hostDir, "notes.txt") // ← .txt, όχι .html
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

	if string(data) != content { // ← συγκρίνουμε με το ΑΡΧΙΚΟ
		t.Errorf("expected file to stay unchanged %q, got %q", content, string(data))
	}
}

func TestRewriteAllRootRelativeHref(t *testing.T) { // ← νέο όνομα
	hostDir := filepath.Join(t.TempDir(), "example.com")
	err := os.MkdirAll(hostDir, 0o755)
	if err != nil {
		t.Fatalf("could not create hostDir: %v", err)
	}

	pagePath := filepath.Join(hostDir, "page.html")
	content := `<a href="/about.html">about</a>` // ← νέο περιεχόμενο
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
