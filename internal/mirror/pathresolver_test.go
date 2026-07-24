package mirror

import "testing"

func TestIsRejected(t *testing.T) {
	result := IsRejected("https://example.com/photo.jpg", []string{".jpg", ".png"})
	if result != true {
		t.Errorf("Expected true, got %v", result)
	}
}
