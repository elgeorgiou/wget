package mirror

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractLinksAnchor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.html")
	if err := os.WriteFile(path, []byte(`<a href="/about">About</a>`), 0644); err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	links, err := ExtractLinks(path)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected one link, got %d", len(links))
	}
	if links[0] != "/about" {
		t.Errorf("expected link '/about', got '%s'", links[0])
	}
}
func TestExtractLinksImg(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.html")
	if err := os.WriteFile(path, []byte(`<img src="/image.jpg" />`), 0644); err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	links, err := ExtractLinks(path)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected one link, got %d", len(links))
	}
	if links[0] != "/image.jpg" {
		t.Errorf("expected link '/image.jpg',got '%s'", links[0])
	}
}
func TestExtractLinksLink(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.html")
	if err := os.WriteFile(path, []byte(`<link rel="stylesheet" href="/style.css" />`), 0644); err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	links, err := ExtractLinks(path)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected one link, got %d", len(links))
	}
	if links[0] != "/style.css" {
		t.Errorf("expected link '/style.css',got '%s'", links[0])
	}
}
