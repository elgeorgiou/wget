package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
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
func TestRunMirror(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("<html><body>Hello Mirror</body></html>"))
			},
		),
	)
	defer server.Close()
	args := []string{"--mirror", server.URL}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	dir := t.TempDir()
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if err = os.Chdir(dir); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer os.Chdir(original)
	err = Run(args, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	parsedURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	expectedPath := filepath.Join(parsedURL.Host, "index.html")
	_, err = os.Stat(expectedPath)
	if err != nil {
		t.Fatalf("Expected file %s to exist, got %v", expectedPath, err)
	}
}
func TestRunMirrorReject(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/":
					w.Header().Set("Content-Type", "text/html")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`<html><body><img src="/photo.jpg"></body></html>`))
				case "/photo.jpg":
					w.Header().Set("Content-Type", "image/jpeg")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("fake image data"))
				}
			},
		),
	)
	defer server.Close()
	dir := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	args := []string{"--mirror", "-R=jpg", server.URL}
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	err = os.Chdir(dir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer os.Chdir(original)
	err = Run(args, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	parsedURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	expectedPath := filepath.Join(parsedURL.Host, "index.html")
	_, err = os.Stat(expectedPath)
	if err != nil {
		t.Fatalf("Expected file %s to exist, got %v", expectedPath, err)
	}
	imagePath := filepath.Join(parsedURL.Host, "photo.jpg")
	_, err = os.Stat(imagePath)
	result := os.IsNotExist(err)
	if !result {
		t.Fatalf("Expected file %s to not exist, got %v", imagePath, err)
	}
}
func TestRunMirrorConvertLinks(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/":
					w.Header().Set("Content-Type", "text/html")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`<html><body><a href="/about.html">About</a></body></html>`))
				case "/about.html":
					w.Header().Set("Content-Type", "text/html")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`<html><body><h1>About</h1></body></html>`))
				}
			},
		),
	)
	defer server.Close()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	dir := t.TempDir()
	args := []string{"--mirror", "--convert-links", server.URL}
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	err = os.Chdir(dir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer os.Chdir(original)
	err = Run(args, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	parsedURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	expectedPath := filepath.Join(parsedURL.Host, "index.html")
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Expected no error reading %s, got %v", expectedPath, err)
	}
	if !bytes.Contains(content, []byte(`href="about.html"`)) {
		t.Fatalf("Expected converted link in index.html, got %s", content)
	}
}
func TestRunCustomDir(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("test file content"))
			},
		),
	)
	defer server.Close()

	dir := t.TempDir()
	outputDir := filepath.Join(dir, "downloads")
	args := []string{"-P", outputDir, server.URL + "/file.txt"}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run(args, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedPath := filepath.Join(outputDir, "file.txt")
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Expected file %s to exist, got %v", expectedPath, err)
	}
	if string(content) != "test file content" {
		t.Fatalf("Expected test file content, got %q", string(content))
	}
}

func TestRunMirrorExclude(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/":
					w.Header().Set("Content-Type", "text/html")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`<html><body><a href="/private/page.html">Private</a></body></html>`))
				case "/private/page.html":
					w.Header().Set("Content-Type", "text/html")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("<html><body>Private page</body></html>"))
				}
			},
		),
	)
	defer server.Close()

	dir := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	args := []string{"--mirror", "-X=/private", server.URL}

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if err = os.Chdir(dir); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer os.Chdir(original)

	err = Run(args, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	parsedURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedPath := filepath.Join(parsedURL.Host, "index.html")
	if _, err = os.Stat(expectedPath); err != nil {
		t.Fatalf("Expected file %s to exist, got %v", expectedPath, err)
	}

	excludedPath := filepath.Join(parsedURL.Host, "private", "page.html")
	if _, err = os.Stat(excludedPath); !os.IsNotExist(err) {
		t.Fatalf("Expected file %s to not exist, got %v", excludedPath, err)
	}
}

func TestRunMirrorRejectLong(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/":
					w.Header().Set("Content-Type", "text/html")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`<html><body><img src="/photo.jpg"></body></html>`))
				case "/photo.jpg":
					w.Header().Set("Content-Type", "image/jpeg")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("fake image data"))
				}
			},
		),
	)
	defer server.Close()

	dir := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	args := []string{"--mirror", "--reject=jpg", server.URL}

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if err = os.Chdir(dir); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer os.Chdir(original)

	err = Run(args, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	parsedURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedPath := filepath.Join(parsedURL.Host, "index.html")
	if _, err = os.Stat(expectedPath); err != nil {
		t.Fatalf("Expected file %s to exist, got %v", expectedPath, err)
	}

	imagePath := filepath.Join(parsedURL.Host, "photo.jpg")
	if _, err = os.Stat(imagePath); !os.IsNotExist(err) {
		t.Fatalf("Expected file %s to not exist, got %v", imagePath, err)
	}
}

func TestRunMirrorExcludeLong(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/":
					w.Header().Set("Content-Type", "text/html")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`<html><body><a href="/private/page.html">Private</a></body></html>`))
				case "/private/page.html":
					w.Header().Set("Content-Type", "text/html")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("<html><body>Private page</body></html>"))
				}
			},
		),
	)
	defer server.Close()

	dir := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	args := []string{"--mirror", "--exclude=/private", server.URL}

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if err = os.Chdir(dir); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer os.Chdir(original)

	err = Run(args, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	parsedURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedPath := filepath.Join(parsedURL.Host, "index.html")
	if _, err = os.Stat(expectedPath); err != nil {
		t.Fatalf("Expected file %s to exist, got %v", expectedPath, err)
	}

	excludedPath := filepath.Join(parsedURL.Host, "private", "page.html")
	if _, err = os.Stat(excludedPath); !os.IsNotExist(err) {
		t.Fatalf("Expected file %s to not exist, got %v", excludedPath, err)
	}
}
