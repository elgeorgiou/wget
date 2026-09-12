package progress

import (
	"testing"
	"time"
)

func TestRenderSize(t *testing.T) {
	elapsed := 2 * time.Second
	result := Render(56370, 102400, elapsed)
	expected := "55.05 KiB / 100.00 KiB"
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
