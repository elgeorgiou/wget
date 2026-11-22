package progress

import (
	"strings"
	"testing"
	"time"
)

func bar(filled int) string {
	return strings.Repeat("=", filled) + strings.Repeat(" ", barWidth-filled)
}

func TestRenderZeroProgress(t *testing.T) {
	result := Render(0, 102400, time.Second)
	expected := " 0.00 KiB / 100.00 KiB [" + bar(0) + "] 0.00% 0.00 KiB/s --"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRenderHalfway(t *testing.T) {
	result := Render(1048576, 2097152, time.Second)
	expected := " 1.00 MiB / 2.00 MiB [" + bar(24) + "] 50.00% 1.00 MiB/s 1s"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRenderComplete(t *testing.T) {
	result := Render(56370, 56370, time.Second)
	expected := " 55.05 KiB / 55.05 KiB [" + bar(48) + "] 100.00% 55.05 KiB/s 0s"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRenderKiBUnits(t *testing.T) {
	result := Render(56370, 102400, 2*time.Second)
	expected := " 55.05 KiB / 100.00 KiB [" + bar(26) + "] 55.05% 27.52 KiB/s 1s"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRenderMiBUnits(t *testing.T) {
	result := Render(5*1048576, 10*1048576, time.Second)
	expected := " 5.00 MiB / 10.00 MiB [" + bar(24) + "] 50.00% 5.00 MiB/s 1s"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRenderETAMinutes(t *testing.T) {
	result := Render(1024, 66*1024, time.Second)
	if !strings.HasSuffix(result, " 1m 5s") {
		t.Errorf("expected ETA 1m 5s, got %q", result)
	}
}

func TestRenderUnknownTotal(t *testing.T) {
	result := Render(56370, -1, time.Second)
	expected := " 55.05 KiB 55.05 KiB/s"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRenderDownloadedOverTotal(t *testing.T) {
	result := Render(200, 100, time.Second)
	if !strings.Contains(result, "["+bar(48)+"] 100.00%") {
		t.Errorf("expected full bar at 100%%, got %q", result)
	}
}

func TestRenderZeroElapsed(t *testing.T) {
	result := Render(100, 200, 0)
	if strings.Contains(result, "Inf") || strings.Contains(result, "NaN") {
		t.Errorf("expected no Inf/NaN, got %q", result)
	}
	if !strings.HasSuffix(result, " --") {
		t.Errorf("expected unknown ETA --, got %q", result)
	}
}

func TestRenderSpeedUnits(t *testing.T) {
	tests := []struct {
		name       string
		downloaded int64
		want       string
	}{
		{"rate limit 300k", 300 * 1024, " 300.00 KiB/s "},
		{"just below 1 MiB/s", 1048575, " 1024.00 KiB/s "},
		{"exactly 1 MiB/s", 1048576, " 1.00 MiB/s "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(tt.downloaded, 10*1048576, time.Second)
			if !strings.Contains(got, tt.want) {
				t.Errorf("expected speed %q in %q", tt.want, got)
			}
		})
	}
}
