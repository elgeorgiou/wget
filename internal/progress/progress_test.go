package progress

import (
	"testing"
	"time"
)

func TestRenderSize(t *testing.T) {
	elapsed := 2 * time.Second
	result := Render(56370, 102400, elapsed)
	expected := "55.05 KiB / 100.00 KiB [===========         ] 55.05% 0.03 MiB/s 1s"
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestRenderPercentage(t *testing.T) {
	elapsed := 2 * time.Second
	result := Render(50, 100, elapsed)
	expected := "0.05 KiB / 0.10 KiB [==========          ] 50.00% 0.00 MiB/s 2s"
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestRenderComplete(t *testing.T) {
	elapsed := 1 * time.Second
	result := Render(1048576, 2097152, elapsed)
	expected := "1024.00 KiB / 2048.00 KiB [==========          ] 50.00% 1.00 MiB/s 1s"
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
