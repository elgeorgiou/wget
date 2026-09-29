package progress

import (
	"testing"
	"time"
)

func TestRenderSize(t *testing.T) {
	elapsed := 2 * time.Second
	result := Render(56370, 102400, elapsed)
	expected := "55.05 KiB / 100.00 KiB [===========         ] 55.05%"
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestRenderPercentage(t *testing.T) {
	elapsed := 2 * time.Second
	result := Render(50, 100, elapsed)
	expected := "0.05 KiB / 0.10 KiB [==========          ] 50.00%"
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
