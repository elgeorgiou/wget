package downloader

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestDownloadFileStatus200(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)

			},
		),
	)
	defer server.Close()
	cfg := DownloadConfig{
		URL: server.URL,
	}
	_, err := DownloadFile(cfg)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)

	}
}
func TestDownloadFileStatus404(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)

			},
		),
	)
	defer server.Close()
	cfg := DownloadConfig{
		URL: server.URL,
	}
	_, err := DownloadFile(cfg)
	if err == nil {
		t.Fatalf("Expected error,got %v", err)
	}
	if errText := strings.Contains(err.Error(), "404"); !errText {
		t.Errorf("Error message should contain status code 404,got %q", err.Error())
	}

}
func TestDownloadFileCustomName(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	cfg := DownloadConfig{
		URL:      server.URL,
		FileName: "test",
	}
	savedPath, err := DownloadFile(cfg)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if savedPath != cfg.FileName {
		t.Errorf("Expected %v, got %v", cfg.FileName, savedPath)
	}
}
func TestDownloadFileCustomDir(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	cfg := DownloadConfig{
		URL:       server.URL + "/test.jpg",
		OutputDir: "test",
	}
	savedPath, err := DownloadFile(cfg)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if savedPath != "test/test.jpg" {
		t.Errorf("Expected test/test.jpg, got %v", savedPath)
	}
}
func TestDownloadFileCombinedFlags(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	cfg := DownloadConfig{
		URL:       server.URL + "/test.jpg",
		FileName:  "test",
		OutputDir: "test",
	}
	savedPath, err := DownloadFile(cfg)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if savedPath != "test/test" {
		t.Errorf("Expected test/test, got %v", savedPath)
	}
}
func TestDownloadFileSavesContent(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Hello World"))
			},
		),
	)
	defer server.Close()
	dir := t.TempDir()
	cfg := DownloadConfig{
		URL:       server.URL,
		OutputDir: dir,
		FileName:  "test.txt",
	}
	savedPath, err := DownloadFile(cfg)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if savedPath != dir+"/test.txt" {
		t.Errorf("Expected %v, got %v", dir+"/test.txt", savedPath)
	}
	content, err := os.ReadFile(savedPath)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if string(content) != "Hello World" {
		t.Errorf("Expected %v, got %v", "Hello World", string(content))

	}
}
