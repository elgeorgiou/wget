package downloader

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// newTestServer serves "content of <path>" for every path except /missing,
// which returns 404. Each response waits delay before being sent.
func newTestServer(t *testing.T, delay time.Duration) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		if r.URL.Path == "/missing" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte("content of " + r.URL.Path))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestDownloadAllTwoURLs(t *testing.T) {
	srv := newTestServer(t, 0)
	dir := t.TempDir()
	urls := []string{srv.URL + "/a.txt", srv.URL + "/b.txt"}

	errs := DownloadAll(urls, DownloadConfig{OutputDir: dir})
	if len(errs) != 2 {
		t.Fatalf("expected 2 results, got %d", len(errs))
	}
	for i, err := range errs {
		if err != nil {
			t.Errorf("url %d: expected nil error, got %v", i, err)
		}
	}
	for _, name := range []string{"a.txt", "b.txt"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("expected %s to be saved: %v", name, err)
		}
		if string(data) != "content of /"+name {
			t.Errorf("%s: unexpected content %q", name, data)
		}
	}
}

func TestDownloadAllEmptyList(t *testing.T) {
	errs := DownloadAll([]string{}, DownloadConfig{OutputDir: t.TempDir()})
	if len(errs) != 0 {
		t.Errorf("expected 0 results, got %d", len(errs))
	}
}

func TestDownloadAllOneError(t *testing.T) {
	srv := newTestServer(t, 0)
	urls := []string{srv.URL + "/1.txt", srv.URL + "/missing", srv.URL + "/3.txt"}

	errs := DownloadAll(urls, DownloadConfig{OutputDir: t.TempDir()})
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
	srv := newTestServer(t, 500*time.Millisecond)
	urls := []string{srv.URL + "/1.txt", srv.URL + "/2.txt", srv.URL + "/3.txt"}

	start := time.Now()
	DownloadAll(urls, DownloadConfig{OutputDir: t.TempDir()})
	elapsed := time.Since(start)

	if elapsed >= 1000*time.Millisecond {
		t.Errorf("expected concurrent downloads under 1s, took %v", elapsed)
	}
}

// All downloads print to the same Out. The lines of one download must stay
// together and not get mixed with the lines of another download.
func TestDownloadAllConcurrentOutput(t *testing.T) {
	srv := newTestServer(t, 50*time.Millisecond)
	urls := []string{
		srv.URL + "/1.txt",
		srv.URL + "/2.txt",
		srv.URL + "/3.txt",
		srv.URL + "/4.txt",
		srv.URL + "/5.txt",
	}

	var out bytes.Buffer
	DownloadAll(urls, DownloadConfig{OutputDir: t.TempDir(), Out: &out})

	// Every download starts with "start at ", so cutting there gives one
	// block per download. The text before the first cut is empty.
	blocks := strings.Split(out.String(), "start at ")[1:]
	if len(blocks) != len(urls) {
		t.Fatalf("expected %d blocks, got %d:\n%s", len(urls), len(blocks), out.String())
	}

	// A block with two "saving file to: " lines contains two downloads.
	for _, block := range blocks {
		if strings.Count(block, "saving file to: ") != 1 {
			t.Errorf("block is mixed with another download:\n%s", block)
		}
	}
}

// With -O every download would write to the same file, so DownloadAll must
// ignore FileName and let each file keep the name from its url.
func TestDownloadAllIgnoresFileName(t *testing.T) {
	srv := newTestServer(t, 0)
	dir := t.TempDir()
	urls := []string{srv.URL + "/a.txt", srv.URL + "/b.txt"}

	DownloadAll(urls, DownloadConfig{OutputDir: dir, FileName: "same.txt"})

	if _, err := os.Stat(filepath.Join(dir, "a.txt")); err != nil {
		t.Errorf("expected a.txt to be saved")
	}
	if _, err := os.Stat(filepath.Join(dir, "b.txt")); err != nil {
		t.Errorf("expected b.txt to be saved")
	}
	if _, err := os.Stat(filepath.Join(dir, "same.txt")); err == nil {
		t.Errorf("expected same.txt not to exist")
	}
}
