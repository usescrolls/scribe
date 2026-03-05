package scribe

import (
	"strings"
	"testing"
)

func TestSanitizeSubpath(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		// Valid subpaths
		{name: "simple path", input: "skills/my-skill"},
		{name: "nested path", input: "path/to/skill"},
		{name: "single segment", input: "src"},
		{name: "hidden dir", input: ".hidden"},
		{name: "dotfile", input: "file.txt"},
		{name: "dotted prefix", input: "..skill"},
		{name: "dotted suffix", input: "skill.."},
		{name: "nested dotfile", input: "path/to/.config"},

		// Traversal attacks
		{name: "parent dir", input: "../etc", expectErr: true},
		{name: "double parent", input: "../../etc/passwd", expectErr: true},
		{name: "mid traversal", input: "skills/../../etc", expectErr: true},
		{name: "deep traversal", input: "a/b/../../../etc", expectErr: true},
		{name: "backslash traversal", input: "..\\etc", expectErr: true},
		{name: "backslash double", input: "..\\..\\secret", expectErr: true},
		{name: "bare dotdot", input: "..", expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizeSubpath(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("SanitizeSubpath(%q) = %q, want error", tt.input, result)
				} else if !strings.Contains(err.Error(), "unsafe subpath") {
					t.Errorf("SanitizeSubpath(%q) error = %q, want 'unsafe subpath'", tt.input, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("SanitizeSubpath(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.input {
					t.Errorf("SanitizeSubpath(%q) = %q, want %q", tt.input, result, tt.input)
				}
			}
		})
	}
}

func TestIsSubpathSafe(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		subpath  string
		expected bool
	}{
		// Safe paths
		{name: "simple subdir", base: "/tmp/repo", subpath: "skills", expected: true},
		{name: "nested subdir", base: "/tmp/repo", subpath: "skills/my-skill", expected: true},
		{name: "deep subdir", base: "/tmp/repo", subpath: "a/b/c", expected: true},
		{name: "current dir", base: "/tmp/repo", subpath: ".", expected: true},
		{name: "normalized within", base: "/tmp/repo", subpath: "skills/../other", expected: true},
		{name: "back to base", base: "/tmp/repo", subpath: "skills/..", expected: true},

		// Unsafe paths
		{name: "parent escape", base: "/tmp/repo", subpath: "..", expected: false},
		{name: "parent with path", base: "/tmp/repo", subpath: "../etc", expected: false},
		{name: "double escape", base: "/tmp/repo", subpath: "../../etc/passwd", expected: false},
		{name: "deep escape", base: "/tmp/repo", subpath: "skills/../../..", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSubpathSafe(tt.base, tt.subpath)
			if result != tt.expected {
				t.Errorf("IsSubpathSafe(%q, %q) = %v, want %v", tt.base, tt.subpath, result, tt.expected)
			}
		})
	}
}

func TestParseSource_RejectsTraversal(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// GitHub shorthand
		{name: "shorthand parent escape", input: "owner/repo/../../etc"},
		{name: "shorthand deep escape", input: "owner/repo/a/../../../etc"},

		// GitHub tree URLs
		{name: "github tree URL escape", input: "https://github.com/owner/repo/tree/main/../../etc"},
		{name: "github tree URL deep escape", input: "https://github.com/owner/repo/tree/main/a/b/../../../etc"},

		// GitHub plain URLs with subpath
		{name: "github plain URL escape", input: "https://github.com/owner/repo/../../etc"},

		// SSH URLs
		{name: "ssh github escape", input: "git@github.com:owner/repo/../../etc"},
		{name: "ssh bitbucket escape", input: "git@bitbucket.org:owner/repo/../../etc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSourceString(tt.input)
			if err == nil {
				t.Errorf("ParseSourceString(%q) should have returned error for path traversal", tt.input)
			} else if !strings.Contains(err.Error(), "unsafe subpath") {
				t.Errorf("ParseSourceString(%q) error = %q, want 'unsafe subpath'", tt.input, err.Error())
			}
		})
	}
}

func TestParseSource_AllowsValidSubpaths(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectSubpath string
	}{
		{
			name:          "github shorthand with subpath",
			input:         "owner/repo/skills/my-skill",
			expectSubpath: "skills/my-skill",
		},
		{
			name:          "github tree URL with subpath",
			input:         "https://github.com/owner/repo/tree/main/skills/my-skill",
			expectSubpath: "skills/my-skill",
		},
		{
			name:          "ssh github with subpath",
			input:         "git@github.com:owner/repo/skills/my-skill",
			expectSubpath: "skills/my-skill",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := ParseSourceString(tt.input)
			if err != nil {
				t.Fatalf("ParseSourceString(%q) unexpected error: %v", tt.input, err)
			}
			if source.Subpath != tt.expectSubpath {
				t.Errorf("ParseSourceString(%q) Subpath = %q, want %q", tt.input, source.Subpath, tt.expectSubpath)
			}
		})
	}
}
