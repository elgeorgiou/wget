package ratelimit

import (
	"errors"
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

type failingReader struct{}

func (failingReader) Read(p []byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestRateLimitPropagatesError(t *testing.T) {
	throttled := NewReader(failingReader{}, 5)

	buffer := make([]byte, 10)

	n, err := throttled.Read(buffer)

	if n != 0 {
		t.Fatalf("expected 0 bytes, got %d", n)
	}

	if err == nil {
		t.Fatal("expected a read error")
	}

	if err.Error() != "read failed" {
		t.Fatalf("expected read failed, got %v", err)
	}
}

type dataAndErrorReader struct{}

func (dataAndErrorReader) Read(p []byte) (int, error) {
	n := copy(p, []byte("Hello"))
	return n, io.ErrUnexpectedEOF
}

func TestRateLimitPreservesDataAndError(t *testing.T) {
	throttled := NewReader(dataAndErrorReader{}, 0)

	buffer := make([]byte, 10)

	n, err := throttled.Read(buffer)

	if n != 5 {
		t.Fatalf("expected 5 bytes, got %d", n)
	}

	if string(buffer[:n]) != "Hello" {
		t.Fatalf("expected Hello, got %q", string(buffer[:n]))
	}

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected unexpected EOF, got %v", err)
	}
}
