package ratelimit

import(
	"io"
	"strings"
	"testing"
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