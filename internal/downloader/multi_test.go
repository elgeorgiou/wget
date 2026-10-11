package downloader

import (
	"errors"
	"testing"
	"time"
)

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

func TestDownloadAllOneError(t *testing.T) {
	urls := []string{"http://a.com/1", "http://a.com/2", "http://a.com/3"}

	failing := func(link string) error {
		if link == "http://a.com/2" {
			return errors.New("simulated download failure")
		}
		return nil
	}
	errs := DownloadAll(urls, failing)
	if len(errs) != 3 {
		t.Fatalf("expected 3 results, got %d", len(errs))
	}
	if errs[0] != nil {
		t.Errorf("expected nil at index 0, got %v", errs[0])
	}
	if errs[1] == nil {
		t.Errorf("expected an error at index 1")
	}
	if errs[2] != nil {
		t.Errorf("expected nil at index 2, got %v", errs[2])
	}
}

func TestDownloadAllConcurrent(t *testing.T) {
	urls := []string{"http://a.com/1", "http://a.com/2", "http://a.com/3"}

	slow := func(link string) error {
		time.Sleep(500 * time.Millisecond)
		return nil
	}
	start := time.Now()
	DownloadAll(urls, slow)
	elapsed := time.Since(start)

	if elapsed >= 1000*time.Millisecond {
		t.Errorf("expected concurrent downloads under 1s, took %v", elapsed)
	}
}
