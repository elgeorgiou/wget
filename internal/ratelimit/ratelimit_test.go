package ratelimit

import (
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

func TestRateLimitZeroIsUnlimited(t *testing.T) {
	input := "Hello World"
	reader := strings.NewReader(input)

	throttled := NewReader(reader, 0)

	start := time.Now()

	var output strings.Builder

	_, err := io.Copy(&output, throttled)
	if err != nil {
		t.Fatal(err)
	}

	elapsed := time.Since(start)

	if output.String() != input {
		t.Fatalf("expected %q, got %q", input, output.String())
	}

	if elapsed > 100*time.Millisecond {
		t.Fatalf("expected zero limit to be unlimited, took %v", elapsed)
	}
}

func TestRateLimitSlowsDown(t *testing.T) {
	input := strings.Repeat("A", 10)

	reader := strings.NewReader(input)
	throttled := NewReader(reader, 5)

	start := time.Now()

	_, err := io.Copy(io.Discard, throttled)
	if err != nil {
		t.Fatal(err)
	}

	elapsed := time.Since(start)

	if elapsed < 2*time.Second {
		t.Fatalf("expected at least 2 seconds, got %v", elapsed)
	}
}

func TestRateLimitMaxChunkSize(t *testing.T) {
	const maxChunkSize = 32 * 1024

	input := strings.Repeat("A", maxChunkSize+100)
	reader := strings.NewReader(input)

	throttled := NewReader(reader, 0)

	buffer := make([]byte, len(input))

	n, err := throttled.Read(buffer)
	if err != nil {
		t.Fatal(err)
	}

	if n != maxChunkSize {
		t.Fatalf("expected %d bytes, got %d", maxChunkSize, n)
	}
}
