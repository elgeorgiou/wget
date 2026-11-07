package mirror

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"wgetclone/internal/downloader"
)

func chdirTemp(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(old); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})

	return dir
}

func saveMockFile(
	t *testing.T,
	cfg downloader.DownloadConfig,
	content string,
) (string, error) {
	t.Helper()

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return "", err
	}

	path := filepath.Join(cfg.OutputDir, cfg.FileName)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}

	return path, nil
}

func TestMirrorSinglePage(t *testing.T) {
	chdirTemp(t)

	calls := 0

	mockDownload := func(cfg downloader.DownloadConfig) (string, error) {
		calls++
		return saveMockFile(t, cfg, "<html><body>Home page</body></html>")
	}

	err := Mirror("https://example.com", MirrorOptions{}, mockDownload)
	if err != nil {
		t.Fatal(err)
	}

	if calls != 1 {
		t.Fatalf("expected 1 download, got %d", calls)
	}

	if _, err := os.Stat("example.com/index.html"); err != nil {
		t.Fatalf("expected saved file to exist: %v", err)
	}
}

func TestMirrorVisitedOnce(t *testing.T) {
	chdirTemp(t)

	calls := make(map[string]int)

	mockDownload := func(cfg downloader.DownloadConfig) (string, error) {
		calls[cfg.URL]++

		return saveMockFile(
			t,
			cfg,
			`<a href="https://example.com">Home</a>`,
		)
	}

	err := Mirror("https://example.com", MirrorOptions{}, mockDownload)
	if err != nil {
		t.Fatal(err)
	}

	if calls["https://example.com"] != 1 {
		t.Fatalf(
			"expected root to be downloaded once, got %d",
			calls["https://example.com"],
		)
	}
}

func TestMirrorFollowsLinks(t *testing.T) {
	chdirTemp(t)

	downloadedURLs := make(map[string]int)

	mockDownload := func(cfg downloader.DownloadConfig) (string, error) {
		downloadedURLs[cfg.URL]++

		if cfg.URL == "https://example.com/about" {
			return saveMockFile(t, cfg, "<html><body>About page</body></html>")
		}

		return saveMockFile(t, cfg, `<a href="/about">About</a>`)
	}

	err := Mirror("https://example.com", MirrorOptions{}, mockDownload)
	if err != nil {
		t.Fatal(err)
	}

	if downloadedURLs["https://example.com"] != 1 {
		t.Fatalf("expected root to be downloaded once")
	}

	if downloadedURLs["https://example.com/about"] != 1 {
		t.Fatalf("expected about page to be downloaded once")
	}
}

func TestMirrorDoesNotFollowExternal(t *testing.T) {
	chdirTemp(t)

	downloadedURLs := make(map[string]int)

	mockDownload := func(cfg downloader.DownloadConfig) (string, error) {
		downloadedURLs[cfg.URL]++

		return saveMockFile(
			t,
			cfg,
			`<a href="https://external.com/page">External</a>`,
		)
	}

	err := Mirror("https://example.com", MirrorOptions{}, mockDownload)
	if err != nil {
		t.Fatal(err)
	}

	if downloadedURLs["https://example.com"] != 1 {
		t.Fatalf("expected root to be downloaded once")
	}

	if downloadedURLs["https://external.com/page"] != 0 {
		t.Fatal("expected external URL not to be downloaded")
	}
}

func TestMirrorRejectSuffix(t *testing.T) {
	chdirTemp(t)

	downloadedURLs := make(map[string]int)

	mockDownload := func(cfg downloader.DownloadConfig) (string, error) {
		downloadedURLs[cfg.URL]++

		if cfg.URL == "https://example.com/about" {
			return saveMockFile(t, cfg, "<html>About</html>")
		}

		return saveMockFile(t, cfg, `
			<a href="/about">About</a>
			<a href="/image.jpg">Image</a>
		`)
	}

	err := Mirror(
		"https://example.com",
		MirrorOptions{RejectSuffixes: []string{".jpg"}},
		mockDownload,
	)
	if err != nil {
		t.Fatal(err)
	}

	if downloadedURLs["https://example.com/about"] != 1 {
		t.Fatal("expected about page to be downloaded")
	}

	if downloadedURLs["https://example.com/image.jpg"] != 0 {
		t.Fatal("expected rejected image not to be downloaded")
	}
}

func TestMirrorExcludePath(t *testing.T) {
	chdirTemp(t)

	downloadedURLs := make(map[string]int)

	mockDownload := func(cfg downloader.DownloadConfig) (string, error) {
		downloadedURLs[cfg.URL]++

		if cfg.URL == "https://example.com/about" {
			return saveMockFile(t, cfg, "<html>About</html>")
		}

		return saveMockFile(t, cfg, `
			<a href="/about">About</a>
			<a href="/private/secret">Private</a>
		`)
	}

	err := Mirror(
		"https://example.com",
		MirrorOptions{ExcludePaths: []string{"/private"}},
		mockDownload,
	)
	if err != nil {
		t.Fatal(err)
	}

	if downloadedURLs["https://example.com/about"] != 1 {
		t.Fatal("expected about page to be downloaded")
	}

	if downloadedURLs["https://example.com/private/secret"] != 0 {
		t.Fatal("expected private URL not to be downloaded")
	}
}

func TestMirrorDomainFolder(t *testing.T) {
	chdirTemp(t)

	var receivedOutputDir string
	var receivedFileName string

	mockDownload := func(cfg downloader.DownloadConfig) (string, error) {
		receivedOutputDir = cfg.OutputDir
		receivedFileName = cfg.FileName

		return saveMockFile(t, cfg, "<html>Home</html>")
	}

	err := Mirror("https://example.com", MirrorOptions{}, mockDownload)
	if err != nil {
		t.Fatal(err)
	}

	if receivedOutputDir != "example.com" {
		t.Fatalf(
			"expected output directory example.com, got %q",
			receivedOutputDir,
		)
	}

	if receivedFileName != "index.html" {
		t.Fatalf(
			"expected filename index.html, got %q",
			receivedFileName,
		)
	}
}

func TestMirrorConvertLinks(t *testing.T) {
	chdirTemp(t)

	htmlContent := `<a href="https://example.com/about.html">About</a>`

	mockDownload := func(cfg downloader.DownloadConfig) (string, error) {
		if cfg.URL == "https://example.com/about.html" {
			return saveMockFile(t, cfg, "<html>About page</html>")
		}

		return saveMockFile(t, cfg, htmlContent)
	}

	err := Mirror(
		"https://example.com",
		MirrorOptions{ConvertLinks: true},
		mockDownload,
	)
	if err != nil {
		t.Fatal(err)
	}

	convertedContent, err := os.ReadFile("example.com/index.html")
	if err != nil {
		t.Fatal(err)
	}

	if string(convertedContent) == htmlContent {
		t.Fatal("expected links to be rewritten")
	}
}

func TestMirrorNormalizesTrailingSlash(t *testing.T) {
	chdirTemp(t)

	downloadedURLs := make(map[string]int)

	mockDownload := func(cfg downloader.DownloadConfig) (string, error) {
		downloadedURLs[cfg.URL]++

		if cfg.URL == "https://example.com" {
			return saveMockFile(t, cfg, `
				<a href="/page">Page</a>
				<a href="/page/">Page with slash</a>
			`)
		}

		return saveMockFile(t, cfg, "<html>Page</html>")
	}

	err := Mirror("https://example.com", MirrorOptions{}, mockDownload)
	if err != nil {
		t.Fatal(err)
	}

	pageDownloads := downloadedURLs["https://example.com/page"] +
		downloadedURLs["https://example.com/page/"]

	if pageDownloads != 1 {
		t.Fatalf(
			"expected page to be downloaded once, got %d",
			pageDownloads,
		)
	}
}

func TestMirrorRootDownloadFailure(t *testing.T) {
	chdirTemp(t)

	mockDownload := func(cfg downloader.DownloadConfig) (string, error) {
		return "", errors.New("download failed")
	}

	err := Mirror("https://example.com", MirrorOptions{}, mockDownload)

	if err == nil {
		t.Fatal("expected root download failure to return an error")
	}
}

func TestMirrorContinuesAfterDownloadFailure(t *testing.T) {
	chdirTemp(t)

	downloadedURLs := make(map[string]int)

	mockDownload := func(cfg downloader.DownloadConfig) (string, error) {
		downloadedURLs[cfg.URL]++

		switch cfg.URL {
		case "https://example.com":
			return saveMockFile(t, cfg, `
				<a href="/broken.html">Broken</a>
				<a href="/working.html">Working</a>
			`)
		case "https://example.com/broken.html":
			return "", errors.New("404 not found")
		case "https://example.com/working.html":
			return saveMockFile(t, cfg, "<html>Working</html>")
		default:
			return "", errors.New("unexpected URL")
		}
	}

	err := Mirror("https://example.com", MirrorOptions{}, mockDownload)
	if err != nil {
		t.Fatal(err)
	}

	if downloadedURLs["https://example.com/working.html"] != 1 {
		t.Fatal("expected crawler to continue after failed download")
	}
}
