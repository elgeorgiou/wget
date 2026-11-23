package mirror

import (
	"testing"
)

func TestIsRejectedMatch(t *testing.T) {
	if !IsRejected("https://example.com/photo.jpg", []string{".jpg", ".png"}) {
		t.Errorf("expected true, got false")
	}
}

func TestIsRejectedCaseInsensitive(t *testing.T) {
	if !IsRejected("https://example.com/photo.JPG", []string{".jpg"}) {
		t.Errorf("expected true, got false")
	}
}

func TestIsRejectedSuffixForms(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		suffixes []string
		want     bool
	}{
		{"suffix without dot", "https://example.com/a.gif", []string{"gif"}, true},
		{"uppercase suffix", "https://example.com/a.jpg", []string{"JPG"}, true},
		{"suffix with spaces", "https://example.com/a.jpg", []string{" jpg "}, true},
		{"no dot does not match name ending", "https://example.com/notjpg", []string{"jpg"}, false},
		{"empty suffix rejects nothing", "https://example.com/index.html", []string{"jpg", ""}, false},
		{"query is ignored", "https://example.com/a.jpg?v=2", []string{"jpg"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRejected(tt.url, tt.suffixes); got != tt.want {
				t.Errorf("IsRejected(%q, %q) = %v, want %v", tt.url, tt.suffixes, got, tt.want)
			}
		})
	}
}

func TestIsExcludedMatch(t *testing.T) {
	if !IsExcluded("https://example.com/js/app.js", []string{"/js"}) {
		t.Errorf("expected true, got false")
	}
}

func TestIsExcludedNoMatch(t *testing.T) {
	if IsExcluded("https://example.com/css/style.css", []string{"/js", "/assets"}) {
		t.Errorf("expected false, got true")
	}
}

func TestIsExcludedEdgeCase(t *testing.T) {
	if IsExcluded("https://example.com/csslib/", []string{"/css"}) {
		t.Errorf("expected false, got true")
	}
}

func TestIsExcludedPrefixForms(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		prefixes []string
		want     bool
	}{
		{"trailing slash", "https://example.com/js/app.js", []string{"/js/"}, true},
		{"no leading slash", "https://example.com/js/app.js", []string{"js"}, true},
		{"exact directory", "https://example.com/js", []string{"/js/"}, true},
		{"nested prefix", "https://example.com/assets/img/a.png", []string{"/assets/img"}, true},
		{"empty prefix excludes nothing", "https://example.com/index.html", []string{""}, false},
		{"slash only excludes nothing", "https://example.com/index.html", []string{"/"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsExcluded(tt.url, tt.prefixes); got != tt.want {
				t.Errorf("IsExcluded(%q, %q) = %v, want %v", tt.url, tt.prefixes, got, tt.want)
			}
		})
	}
}

func TestResolveLocalPathFile(t *testing.T) {
	result, err := ResolveLocalPath("www.example.com", "https://www.example.com/js/app.js")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "www.example.com/js/app.js"
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestResolveLocalPathRoot(t *testing.T) {
	result, err := ResolveLocalPath("www.example.com", "https://www.example.com/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "www.example.com/index.html"
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestResolveLocalPathDirNoExt(t *testing.T) {
	result, err := ResolveLocalPath("www.example.com", "https://www.example.com/about")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "www.example.com/about/index.html"
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

// A trailing slash means a directory, even when the last segment looks like
// a file name, so files inside it can still be saved.
func TestResolveLocalPathTrailingSlash(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://www.example.com/v1.2/", "www.example.com/v1.2/index.html"},
		{"https://www.example.com/jquery-3.6.0/", "www.example.com/jquery-3.6.0/index.html"},
		{"https://www.example.com/v1.2/app.css", "www.example.com/v1.2/app.css"},
		{"https://www.example.com/about/", "www.example.com/about/index.html"},
		{"https://www.example.com/v1.2", "www.example.com/v1.2"},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got, err := ResolveLocalPath("www.example.com", tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ResolveLocalPath(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestResolveLocalPathStaysInHostDir(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://www.example.com/../../etc/passwd", "www.example.com/etc/passwd/index.html"},
		{"https://www.example.com/%2e%2e/%2e%2e/etc/passwd", "www.example.com/etc/passwd/index.html"},
		{"https://www.example.com/a/../../b.css", "www.example.com/b.css"},
	}
	for _, tt := range tests {
		got, err := ResolveLocalPath("www.example.com", tt.url)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tt.url, err)
		}
		if got != tt.want {
			t.Errorf("ResolveLocalPath(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}
