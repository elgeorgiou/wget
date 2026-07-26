package downloader

import (
	"net/http"
	"net/http/httptest"
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
