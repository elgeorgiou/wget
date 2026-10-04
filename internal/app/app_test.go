package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
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
