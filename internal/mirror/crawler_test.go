package mirror

import (
	"os"
	"path/filepath"
	"testing"

	downloader "wgetclone/internal/downloader"
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

func TestMirrorFollowsLinks(t *testing.T) {
	downloadedURLs := make(map[string]int)

	tempDir := t.TempDir()

	rootHTMLPath := filepath.Join(tempDir, "index.html")
	aboutHTMLPath := filepath.Join(tempDir, "about.html")

	rootHTML := `<a href="/about">About</a>`
	aboutHTML := `<html><body>About page</body></html>`

	err := os.WriteFile(rootHTMLPath, []byte(rootHTML), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(aboutHTMLPath, []byte(aboutHTML), 0644)
	if err != nil {
		t.Fatal(err)
	}

	mockDownload := func(cfg downloader.DownloadConfig) (string, error) {
		downloadedURLs[cfg.URL]++

		if cfg.URL == "https://example.com/about" {
			return aboutHTMLPath, nil
		}

		return rootHTMLPath, nil
	}

	err = Mirror(
		"https://example.com",
		MirrorOptions{},
		mockDownload,
	)
	if err != nil {
		t.Fatal(err)
	}

	if downloadedURLs["https://example.com"] != 1 {
		t.Fatalf(
			"expected root URL to be downloaded once, got %d",
			downloadedURLs["https://example.com"],
		)
	}

	if downloadedURLs["https://example.com/about"] != 1 {
		t.Fatalf(
			"expected about URL to be downloaded once, got %d",
			downloadedURLs["https://example.com/about"],
		)
	}
}

func TestMirrorDoesNotFollowExternal(t *testing.T) {
	downloadedURLs := make(map[string]int)

	tempDir := t.TempDir()

	rootHTMLPath := filepath.Join(tempDir, "index.html")

	rootHTML := `<a href="https://external.com/page">External</a>`

	err := os.WriteFile(rootHTMLPath, []byte(rootHTML), 0644)
	if err != nil {
		t.Fatal(err)
	}

	mockDownload := func(cfg downloader.DownloadConfig) (string, error) {
		downloadedURLs[cfg.URL]++
		return rootHTMLPath, nil
	}

	err = Mirror(
		"https://example.com",
		MirrorOptions{},
		mockDownload,
	)
	if err != nil {
		t.Fatal(err)
	}

	if downloadedURLs["https://example.com"] != 1 {
		t.Fatalf(
			"expected root URL to be downloaded once, got %d",
			downloadedURLs["https://example.com"],
		)
	}

	if downloadedURLs["https://external.com/page"] != 0 {
		t.Fatalf(
			"expected external URL not to be downloaded, got %d downloads",
			downloadedURLs["https://external.com/page"],
		)
	}
}

func TestMirrorRejectSuffix(t *testing.T) {
	downloadedURLs := make(map[string]int)

	tempDir := t.TempDir()

	rootHTMLPath := filepath.Join(tempDir, "index.html")
	aboutHTMLPath := filepath.Join(tempDir, "about.html")

	rootHTML := `
		<a href="/about">About</a>
		<a href="/image.jpg">Image</a>
	`
	aboutHTML := `<html><body>About page</body></html>`

	err := os.WriteFile(rootHTMLPath, []byte(rootHTML), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(aboutHTMLPath, []byte(aboutHTML), 0644)
	if err != nil {
		t.Fatal(err)
	}

	mockDownload := func(cfg downloader.DownloadConfig) (string, error) {
		downloadedURLs[cfg.URL]++

		if cfg.URL == "https://example.com/about" {
			return aboutHTMLPath, nil
		}

		return rootHTMLPath, nil
	}

	err = Mirror(
		"https://example.com",
		MirrorOptions{
			RejectSuffixes: []string{".jpg"},
		},
		mockDownload,
	)
	if err != nil {
		t.Fatal(err)
	}

	if downloadedURLs["https://example.com/about"] != 1 {
		t.Fatalf(
			"expected about URL to be downloaded once, got %d",
			downloadedURLs["https://example.com/about"],
		)
	}

	if downloadedURLs["https://example.com/image.jpg"] != 0 {
		t.Fatalf(
			"expected rejected URL not to be downloaded, got %d downloads",
			downloadedURLs["https://example.com/image.jpg"],
		)
	}
}
