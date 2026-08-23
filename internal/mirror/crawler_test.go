package mirror

import (
	"os"
	"path/filepath"
	"testing"

	"wgetclone/internal/downloader"
)

func TestMirrorSinglePage(t *testing.T) {
	calls := 0

	mockDownload := func(cfg downloader.DownloadConfig) (string, error) {
		calls++
		return "index.html", nil
	}

	err := Mirror(
		"https://example.com",
		MirrorOptions{},
		mockDownload,
	)

	if err != nil {
		t.Fatal(err)
	}

	if calls != 1 {
		t.Fatalf("expected 1 download, got %d", calls)
	}
}

func TestMirrorVisitedOnce(t *testing.T) {
	calls := make(map[string]int)

	tempDir := t.TempDir()
	htmlPath := filepath.Join(tempDir, "index.html")
	htmlContent := `<a href="https://example.com">Home</a>`

	err := os.WriteFile(htmlPath, []byte(htmlContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	mockDownload := func(cfg downloader.DownloadConfig) (string, error) {
		calls[cfg.URL]++
		return htmlPath, nil
	}

	err = Mirror(
		"https://example.com",
		MirrorOptions{},
		mockDownload,
	)

	if err != nil {
		t.Fatal(err)
	}

	if calls["https://example.com"] != 1 {
		t.Fatalf(
			"expected https://example.com to be downloaded once, got %d",
			calls["https://example.com"],
		)
	}
}
