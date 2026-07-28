package mirror

import (
	"testing"
)

func TestIsRejectedMatch(t *testing.T) {
	result := IsRejected("https://example.com/photo.jpg", []string{".jpg", ".png"})
	if result != true {
		t.Errorf("Expected true, got %v", result)
	}
}

func TestIsRejectedCaseInsensitive(t *testing.T) {
	result := IsRejected("https://example.com/photo.JPG", []string{".jpg"})
	if result != true {
		t.Errorf("Expected true, got %v", result)
	}
}
