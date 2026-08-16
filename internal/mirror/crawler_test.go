package mirror

import (
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
