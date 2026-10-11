package downloader

import "testing"

func fakeDownload(link string) error {
	return nil
}

func TestDownloadAllTwoURLs(t *testing.T) {
	urls := []string{"http://a.com/1", "http://a.com/2"}
	errs := DownloadAll(urls, fakeDownload)
	if len(errs) != 2 {
		t.Fatalf("expected 2 results, got %d", len(errs))
	}
	for i, err := range errs {
		if err != nil {
			t.Errorf("url %d: expected nil error, got %v", i, err)
		}
	}
}

func TestDownloadAllEmptyList(t *testing.T) {
	errs := DownloadAll([]string{}, fakeDownload)
	if len(errs) != 0 {
		t.Errorf("expected 0 results, got %d", len(errs))
	}
}
