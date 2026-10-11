package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunNoURL(t *testing.T) {
	var args []string
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := Run(args, &stdout, &stderr)
	if err == nil {
		t.Fatalf("Expected error for no URL, got nil")
	}
}
func TestRunSingleDownload(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	args := []string{server.URL}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := Run(args, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}
func TestRunBackground(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	args := []string{"-B", server.URL}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	dir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if err = os.Chdir(dir); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer os.Chdir(originalDir)
	err = Run(args, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if stdout.String() != "Output will be written to \"wget-log\".\n" {
		t.Fatalf("Expected Output will be written to wget-log., got %q", stdout.String())
	}
	if _, err = os.Stat("wget-log"); err != nil {
		t.Fatalf("Expected wget-log file to exist,got %v", err)
	}
	content, err := os.ReadFile("wget-log")
	if err != nil {
		t.Fatalf("Expected no error reading wget-log, got %v", err)
	}
	if len(content) == 0 {
		t.Fatalf("wget-log is empty")
	}
}
func TestRunRateLimit(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(bytes.Repeat([]byte("A"), 1024*1024))
			},
		),
	)
	defer server.Close()
	args := []string{"--rate-limit=400k", server.URL}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	start := time.Now()
	err := Run(args, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	elapsed := time.Since(start)
	if elapsed < 2*time.Second {
		t.Fatalf("Expected download to take at least 2 seconds, got %v", elapsed)
	}
}
func TestRunMultiFile(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	dir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if err = os.Chdir(dir); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer os.Chdir(originalDir)
	filePath := filepath.Join(dir, "links.txt")
	args := []string{"-i", filePath}
	content := server.URL + "/file1.txt" + "\n" + server.URL + "/file2.txt"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	err = Run(args, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if _, err := os.Stat("file1.txt"); err != nil {
		t.Fatalf("Expected file1.txt file to exist,got %v", err)
	}
	if _, err := os.Stat("file2.txt"); err != nil {
		t.Fatalf("Expected file2.txt file to exist,got %v", err)
	}
}
