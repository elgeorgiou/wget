package downloader

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	dir := t.TempDir()
	cfg := DownloadConfig{
		URL:       server.URL,
		OutputDir: dir,
		FileName:  "test",
	}
	_, err := DownloadFile(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)

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
	dir := t.TempDir()
	cfg := DownloadConfig{
		URL:       server.URL,
		FileName:  "test",
		OutputDir: dir,
	}
	savedPath, err := DownloadFile(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	expectedPath := filepath.Join(dir, cfg.FileName)
	if savedPath != expectedPath {
		t.Errorf("Expected %v, got %v", expectedPath, savedPath)
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
	dir := t.TempDir()
	cfg := DownloadConfig{
		URL:       server.URL + "/test.jpg",
		OutputDir: dir,
	}
	savedPath, err := DownloadFile(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	expected := filepath.Join(dir, "test.jpg")
	if savedPath != expected {
		t.Errorf("Expected %v, got %v", expected, savedPath)
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
	dir := t.TempDir()
	cfg := DownloadConfig{
		URL:       server.URL + "/test.jpg",
		FileName:  "test",
		OutputDir: dir,
	}
	savedPath, err := DownloadFile(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	expected := filepath.Join(dir, "test")
	if savedPath != expected {
		t.Errorf("Expected %v, got %v", expected, savedPath)
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
		t.Fatalf("Expected no error, got %v", err)
	}
	expected := filepath.Join(dir, "test.txt")
	if savedPath != expected {
		t.Errorf("Expected %v, got %v", expected, savedPath)
	}
	content, err := os.ReadFile(savedPath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if string(content) != "Hello World" {
		t.Errorf("Expected %v, got %v", "Hello World", string(content))

	}
}
func TestDownloadFileProgressLogged(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Hello World"))
			},
		),
	)
	defer server.Close()
	buffer := bytes.Buffer{}
	dir := t.TempDir()
	cfg := DownloadConfig{
		URL:       server.URL,
		OutputDir: dir,
		FileName:  "test",
		Out:       &buffer,
	}
	_, err := DownloadFile(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if buffer.String() == "" {
		t.Errorf("Expected progress output, got empty string")
	}
	if !strings.Contains(buffer.String(), "\r") {
		t.Errorf("Expected progress output to contain carriage return")
	}
}
