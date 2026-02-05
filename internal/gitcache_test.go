package scribe

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCacheKeyForSource(t *testing.T) {
	tests := []struct {
		name     string
		source   *SourceInfo
		expected string
	}{
		{
			name: "github source",
			source: &SourceInfo{
				Type:  "github",
				Owner: "user",
				Repo:  "repo",
			},
			expected: filepath.Join("github.com", "user", "repo"),
		},
		{
			name: "gitlab source",
			source: &SourceInfo{
				Type:  "gitlab",
				Owner: "org",
				Repo:  "project",
			},
			expected: filepath.Join("gitlab.com", "org", "project"),
		},
		{
			name: "local source returns empty",
			source: &SourceInfo{
				Type:      "local",
				LocalPath: "/some/path",
			},
			expected: "",
		},
		{
			name: "zip source returns empty",
			source: &SourceInfo{
				Type: "zip",
				URL:  "https://example.com/archive.zip",
			},
			expected: "",
		},
		{
			name: "unknown type returns empty",
			source: &SourceInfo{
				Type: "well-known",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CacheKeyForSource(tt.source)
			if result != tt.expected {
				t.Errorf("CacheKeyForSource() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildCloneURL(t *testing.T) {
	tests := []struct {
		name     string
		source   *SourceInfo
		expected string
	}{
		{
			name:     "URL without .git suffix",
			source:   &SourceInfo{URL: "https://github.com/user/repo"},
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "URL already has .git suffix",
			source:   &SourceInfo{URL: "https://github.com/user/repo.git"},
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "gitlab URL",
			source:   &SourceInfo{URL: "https://gitlab.com/org/project"},
			expected: "https://gitlab.com/org/project.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCloneURL(tt.source)
			if result != tt.expected {
				t.Errorf("buildCloneURL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetCacheDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	dir, err := GetCacheDir()
	if err != nil {
		t.Fatalf("GetCacheDir() error: %v", err)
	}

	expected := filepath.Join(home, ".scribe", "cache")
	if dir != expected {
		t.Errorf("GetCacheDir() = %q, want %q", dir, expected)
	}
}

func TestClearCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// Create cache dir with some content
	cacheDir := filepath.Join(tmpDir, ".scribe", "cache")
	fakeRepo := filepath.Join(cacheDir, "github.com", "user", "repo")
	if err := os.MkdirAll(fakeRepo, 0o755); err != nil {
		t.Fatalf("failed to create fake repo dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fakeRepo, "dummy"), []byte("data"), 0o644); err != nil {
		t.Fatalf("failed to write dummy file: %v", err)
	}

	// Clear cache
	if err := ClearCache(); err != nil {
		t.Fatalf("ClearCache() error: %v", err)
	}

	// Cache dir should still exist but be empty
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		t.Fatalf("failed to read cache dir after clear: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected cache dir to be empty, got %d entries", len(entries))
	}
}

func TestClearCacheForSource(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// Create two cached repos
	cacheDir := filepath.Join(tmpDir, ".scribe", "cache")
	repo1 := filepath.Join(cacheDir, "github.com", "user", "repo1")
	repo2 := filepath.Join(cacheDir, "github.com", "user", "repo2")
	if err := os.MkdirAll(repo1, 0o755); err != nil {
		t.Fatalf("failed to create repo1: %v", err)
	}
	if err := os.MkdirAll(repo2, 0o755); err != nil {
		t.Fatalf("failed to create repo2: %v", err)
	}

	// Clear only repo1
	source := &SourceInfo{Type: "github", Owner: "user", Repo: "repo1"}
	if err := ClearCacheForSource(source); err != nil {
		t.Fatalf("ClearCacheForSource() error: %v", err)
	}

	// repo1 should be gone
	if _, err := os.Stat(repo1); !os.IsNotExist(err) {
		t.Error("expected repo1 to be removed")
	}
	// repo2 should still exist
	if _, err := os.Stat(repo2); err != nil {
		t.Error("expected repo2 to still exist")
	}
}

func TestClearCacheForSource_NonCacheable(t *testing.T) {
	// Non-cacheable source should be a no-op
	source := &SourceInfo{Type: "local", LocalPath: "/some/path"}
	if err := ClearCacheForSource(source); err != nil {
		t.Fatalf("ClearCacheForSource() for local source should not error: %v", err)
	}
}
