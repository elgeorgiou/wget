package ratelimit

import(
	"io"
	"strings"
	"testing"
	"time"
)

func TestContentIntact(t *testing.T) {
	input := "Hello World"
	reader := strings.NewReader(input)

	var output strings.Builder

	throttled := NewReader(reader, 5)

	_, err := io.Copy(&output, throttled)
	
	if err != nil {
	t.Fatal(err)
	}

	if output.String() != input {
	t.Fatalf("expected %q, got %q", input, output.String())
	}
}

func TestRateLimiting(t *testing.T) {

	input := strings.Repeat("A", 10)

	reader := strings.NewReader(input)

	throttled := NewReader(reader, 5)

	start := time.Now()

	_, err := io.Copy(io.Discard, throttled)

	if err != nil {
		t.Fatal(err)
	}

	elapsed := time.Since(start)

	if elapsed <2*time.Second {
		t.Fatalf("expected at least 2 seconds, got %v", elapsed)
	}
}