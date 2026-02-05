package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	scribe "github.com/usescrolls/scribe/internal"
)

// TestCLICommands tests that CLICommands returns the expected commands
func TestCLICommands(t *testing.T) {
	commands := CLICommands()

	expected := []string{"install", "uninstall", "remove", "rm", "list", "ls", "info", "version", "help", "workspace", "check", "update", "cache"}

	for _, exp := range expected {
		found := false
		for _, cmd := range commands {
			if cmd == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected command %q not found in CLICommands()", exp)
		}
	}
}

// TestExitCodes tests that exit codes are defined correctly
func TestExitCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected int
	}{
		{"ExitSuccess", ExitSuccess, 0},
		{"ExitError", ExitError, 1},
		{"ExitUsage", ExitUsage, 2},
		{"ExitNotFound", ExitNotFound, 3},
		{"ExitSourceFailed", ExitSourceFailed, 4},
		{"ExitRegistryError", ExitRegistryError, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.expected {
				t.Errorf("%s = %d, expected %d", tt.name, tt.code, tt.expected)
			}
		})
	}
}

// TestFormatSourceInfo tests the formatSourceInfo function
func TestFormatSourceInfo(t *testing.T) {
	tests := []struct {
		name     string
		source   *scribe.SourceInfo
		expected string
	}{
		{
			name: "github source",
			source: &scribe.SourceInfo{
				Type:  "github",
				Owner: "user",
				Repo:  "repo",
			},
			expected: "github:user/repo",
		},
		{
			name: "github source with ref",
			source: &scribe.SourceInfo{
				Type:  "github",
				Owner: "user",
				Repo:  "repo",
				Ref:   "v1.0.0",
			},
			expected: "github:user/repo#v1.0.0",
		},
		{
			name: "local source",
			source: &scribe.SourceInfo{
				Type:      "local",
				LocalPath: "/path/to/skills",
			},
			expected: "local:/path/to/skills",
		},
		{
			name: "zip source",
			source: &scribe.SourceInfo{
				Type: "zip",
				URL:  "https://example.com/plugin.zip",
			},
			expected: "zip:https://example.com/plugin.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSourceInfo(tt.source)
			if result != tt.expected {
				t.Errorf("formatSourceInfo() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestParseSource tests the parseSource function
func TestParseSource(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectType  string
		expectOwner string
		expectRepo  string
		expectRef   string
		expectError bool
	}{
		{
			name:        "github shorthand",
			input:       "owner/repo",
			expectType:  "github",
			expectOwner: "owner",
			expectRepo:  "repo",
		},
		{
			name:        "github shorthand with ref",
			input:       "owner/repo#v1.0.0",
			expectType:  "github",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectRef:   "v1.0.0",
		},
		{
			name:        "github URL",
			input:       "https://github.com/owner/repo",
			expectType:  "github",
			expectOwner: "owner",
			expectRepo:  "repo",
		},
		{
			name:       "local path relative",
			input:      "./my-skills",
			expectType: "local",
		},
		{
			name:       "local path absolute",
			input:      "/absolute/path",
			expectType: "local",
		},
		{
			name:       "zip URL",
			input:      "https://example.com/skills.zip",
			expectType: "zip",
		},
		{
			name:       "well-known URL",
			input:      "https://example.com",
			expectType: "well-known",
		},
		{
			name:        "invalid shorthand",
			input:       "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := parseSource(tt.input)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if source.Type != tt.expectType {
				t.Errorf("expected type %q, got %q", tt.expectType, source.Type)
			}
			if tt.expectOwner != "" && source.Owner != tt.expectOwner {
				t.Errorf("expected owner %q, got %q", tt.expectOwner, source.Owner)
			}
			if tt.expectRepo != "" && source.Repo != tt.expectRepo {
				t.Errorf("expected repo %q, got %q", tt.expectRepo, source.Repo)
			}
			if tt.expectRef != "" && source.Ref != tt.expectRef {
				t.Errorf("expected ref %q, got %q", tt.expectRef, source.Ref)
			}
		})
	}
}

// TestFormatSkillSource tests the formatSkillSource function
func TestFormatSkillSource(t *testing.T) {
	tests := []struct {
		name     string
		info     scribe.SkillInfo
		expected string
	}{
		{
			name: "github source",
			info: scribe.SkillInfo{
				Source:     "owner/repo",
				SourceType: "github",
			},
			expected: "github:owner/repo",
		},
		{
			name: "gitlab source",
			info: scribe.SkillInfo{
				Source:     "owner/repo",
				SourceType: "gitlab",
			},
			expected: "gitlab:owner/repo",
		},
		{
			name: "local source",
			info: scribe.SkillInfo{
				Source:     "/path/to/skill",
				SourceType: "local",
			},
			expected: "local:/path/to/skill",
		},
		{
			name: "zip source",
			info: scribe.SkillInfo{
				Source:     "https://example.com/plugin.zip",
				SourceType: "zip",
			},
			expected: "https://example.com/plugin.zip",
		},
		{
			name: "empty source",
			info: scribe.SkillInfo{
				Source:     "",
				SourceType: "",
			},
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSkillSource(tt.info)
			if result != tt.expected {
				t.Errorf("formatSkillSource() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestListEmptyOutput tests the list command with no skills installed
func TestListEmptyOutput(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "scribe-cli-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Override home directory for test
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	scribe.InitLoggerCLI(false)

	t.Run("empty list", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		jsonOutput = false
		namesOnly = false
		quiet = false
		_ = runList(listCmd, []string{})

		_ = w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		if !strings.Contains(output, "No skills installed") {
			t.Errorf("expected 'No skills installed' message, got: %s", output)
		}
	})
}

// TestVersionCommand tests the version command
func TestVersionCommand(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	runVersion(versionCmd, []string{})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "scribe version") {
		t.Error("expected version output to contain 'scribe version'")
	}
	if !strings.Contains(output, scribe.Version) {
		t.Errorf("expected version output to contain %q", scribe.Version)
	}
}

// TestQuietMode tests that quiet mode suppresses output
func TestQuietMode(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "scribe-cli-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Override home directory for test
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	scribe.InitLoggerCLI(false)

	t.Run("quiet list with no skills", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		quiet = true
		jsonOutput = false
		namesOnly = false
		_ = runList(listCmd, []string{})

		_ = w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		// In quiet mode, "No skills installed" should not be printed
		if strings.Contains(output, "No skills installed") {
			t.Error("quiet mode should suppress 'No skills installed' message")
		}
	})
}

// TestFilterSkills tests the filterSkills function
func TestFilterSkills(t *testing.T) {
	skills := []*scribe.Skill{
		{Name: "react-best-practices", Description: "React patterns"},
		{Name: "typescript-patterns", Description: "TypeScript tips"},
		{Name: "go-patterns", Description: "Go idioms"},
	}

	t.Run("filter single skill", func(t *testing.T) {
		filtered := filterSkills(skills, []string{"react-best-practices"})
		if len(filtered) != 1 {
			t.Errorf("expected 1 skill, got %d", len(filtered))
		}
		if filtered[0].Name != "react-best-practices" {
			t.Errorf("expected react-best-practices, got %s", filtered[0].Name)
		}
	})

	t.Run("filter multiple skills", func(t *testing.T) {
		filtered := filterSkills(skills, []string{"react-best-practices", "go-patterns"})
		if len(filtered) != 2 {
			t.Errorf("expected 2 skills, got %d", len(filtered))
		}
	})

	t.Run("filter non-existent skill", func(t *testing.T) {
		filtered := filterSkills(skills, []string{"non-existent"})
		if len(filtered) != 0 {
			t.Errorf("expected 0 skills, got %d", len(filtered))
		}
	})
}

// ---------------------------------------------------------------------------
// Tests for truncateHash (check.go)
// ---------------------------------------------------------------------------

func TestTruncateHash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "empty string", input: "", expected: "-"},
		{name: "short hash", input: "abc123", expected: "abc123"},
		{name: "exactly 20 chars", input: "12345678901234567890", expected: "12345678901234567890"},
		{name: "21 chars truncated", input: "123456789012345678901", expected: "12345678901234567890..."},
		{name: "long sha256 hash", input: "sha256:abcdef1234567890abcdef1234567890abcdef1234567890", expected: "sha256:abcdef1234567..."},
		{name: "single char", input: "x", expected: "x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateHash(tt.input)
			if result != tt.expected {
				t.Errorf("truncateHash(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for truncateString (list.go)
// ---------------------------------------------------------------------------

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{name: "short string fits", input: "hello", maxLen: 10, expected: "hello"},
		{name: "exact length", input: "hello", maxLen: 5, expected: "hello"},
		{name: "needs truncation", input: "hello world", maxLen: 8, expected: "hello..."},
		{name: "maxLen 3 no ellipsis", input: "hello", maxLen: 3, expected: "hel"},
		{name: "maxLen 2 no ellipsis", input: "hello", maxLen: 2, expected: "he"},
		{name: "maxLen 1 no ellipsis", input: "hello", maxLen: 1, expected: "h"},
		{name: "maxLen 4 with ellipsis", input: "hello", maxLen: 4, expected: "h..."},
		{name: "empty string", input: "", maxLen: 10, expected: ""},
		{name: "long description", input: "This is a very long description that should be truncated", maxLen: 20, expected: "This is a very lo..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%q, %d) = %q, expected %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for formatInstalledAt (list.go)
// ---------------------------------------------------------------------------

func TestFormatInstalledAt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "empty string", input: "", expected: "-"},
		{name: "full ISO8601", input: "2024-01-15T10:30:00Z", expected: "2024-01-15"},
		{name: "date only", input: "2024-01-15", expected: "2024-01-15"},
		{name: "short timestamp", input: "2024-01", expected: "2024-01"},
		{name: "very short", input: "2024", expected: "2024"},
		{name: "exactly 10 chars", input: "2024-01-15", expected: "2024-01-15"},
		{name: "longer than 10", input: "2024-01-15T00:00:00+05:30", expected: "2024-01-15"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatInstalledAt(tt.input)
			if result != tt.expected {
				t.Errorf("formatInstalledAt(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for reconstructSource (check.go)
// ---------------------------------------------------------------------------

func TestReconstructSource(t *testing.T) {
	tests := []struct {
		name          string
		meta          *scribe.SkillMeta
		expectType    string
		expectOwner   string
		expectRepo    string
		expectRef     string
		expectURL     string
		expectSubpath string
		expectLocal   string
	}{
		{
			name: "github owner/repo",
			meta: &scribe.SkillMeta{
				Source:     "myowner/myrepo",
				SourceType: "github",
			},
			expectType:  "github",
			expectOwner: "myowner",
			expectRepo:  "myrepo",
			expectURL:   "https://github.com/myowner/myrepo",
		},
		{
			name: "github owner/repo with ref",
			meta: &scribe.SkillMeta{
				Source:     "myowner/myrepo#v2.0",
				SourceType: "github",
			},
			expectType:  "github",
			expectOwner: "myowner",
			expectRepo:  "myrepo",
			expectRef:   "v2.0",
			expectURL:   "https://github.com/myowner/myrepo",
		},
		{
			name: "github with existing URL",
			meta: &scribe.SkillMeta{
				Source:     "myowner/myrepo",
				SourceType: "github",
				SourceURL:  "https://github.com/myowner/myrepo",
			},
			expectType:  "github",
			expectOwner: "myowner",
			expectRepo:  "myrepo",
			expectURL:   "https://github.com/myowner/myrepo",
		},
		{
			name: "github with subpath in source",
			meta: &scribe.SkillMeta{
				Source:     "myowner/myrepo/skills/react",
				SourceType: "github",
			},
			expectType:    "github",
			expectOwner:   "myowner",
			expectRepo:    "myrepo",
			expectSubpath: "skills/react",
			expectURL:     "https://github.com/myowner/myrepo",
		},
		{
			name: "gitlab owner/repo",
			meta: &scribe.SkillMeta{
				Source:     "glowner/glrepo",
				SourceType: "gitlab",
			},
			expectType:  "gitlab",
			expectOwner: "glowner",
			expectRepo:  "glrepo",
			expectURL:   "https://gitlab.com/glowner/glrepo",
		},
		{
			name: "zip source",
			meta: &scribe.SkillMeta{
				Source:     "https://example.com/skills.zip",
				SourceType: "zip",
			},
			expectType: "zip",
			expectURL:  "https://example.com/skills.zip",
		},
		{
			name: "zip with existing URL",
			meta: &scribe.SkillMeta{
				Source:     "https://example.com/skills.zip",
				SourceType: "zip",
				SourceURL:  "https://example.com/skills.zip",
			},
			expectType: "zip",
			expectURL:  "https://example.com/skills.zip",
		},
		{
			name: "local source",
			meta: &scribe.SkillMeta{
				Source:     "/home/user/my-skills",
				SourceType: "local",
			},
			expectType:  "local",
			expectLocal: "/home/user/my-skills",
		},
		{
			name: "url type",
			meta: &scribe.SkillMeta{
				Source:     "https://example.com/my-skills",
				SourceType: "url",
			},
			expectType: "url",
			expectURL:  "https://example.com/my-skills",
		},
		{
			name: "well-known type",
			meta: &scribe.SkillMeta{
				Source:     "https://example.com",
				SourceType: "well-known",
			},
			expectType: "well-known",
			expectURL:  "https://example.com",
		},
		{
			name: "skillPath overrides subpath",
			meta: &scribe.SkillMeta{
				Source:     "myowner/myrepo",
				SourceType: "github",
				SkillPath:  "deep/path/to/skill",
			},
			expectType:    "github",
			expectOwner:   "myowner",
			expectRepo:    "myrepo",
			expectSubpath: "deep/path/to/skill",
			expectURL:     "https://github.com/myowner/myrepo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reconstructSource(tt.meta)
			if result.Type != tt.expectType {
				t.Errorf("Type = %q, expected %q", result.Type, tt.expectType)
			}
			if tt.expectOwner != "" && result.Owner != tt.expectOwner {
				t.Errorf("Owner = %q, expected %q", result.Owner, tt.expectOwner)
			}
			if tt.expectRepo != "" && result.Repo != tt.expectRepo {
				t.Errorf("Repo = %q, expected %q", result.Repo, tt.expectRepo)
			}
			if tt.expectRef != "" && result.Ref != tt.expectRef {
				t.Errorf("Ref = %q, expected %q", result.Ref, tt.expectRef)
			}
			if tt.expectURL != "" && result.URL != tt.expectURL {
				t.Errorf("URL = %q, expected %q", result.URL, tt.expectURL)
			}
			if tt.expectSubpath != "" && result.Subpath != tt.expectSubpath {
				t.Errorf("Subpath = %q, expected %q", result.Subpath, tt.expectSubpath)
			}
			if tt.expectLocal != "" && result.LocalPath != tt.expectLocal {
				t.Errorf("LocalPath = %q, expected %q", result.LocalPath, tt.expectLocal)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for parseGitLabURL (install.go)
// ---------------------------------------------------------------------------

func TestParseGitLabURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectOwner string
		expectRepo  string
		expectType  string
		expectURL   string
	}{
		{
			name:        "basic gitlab URL",
			url:         "https://gitlab.com/myowner/myrepo",
			expectOwner: "myowner",
			expectRepo:  "myrepo",
			expectType:  "gitlab",
			expectURL:   "https://gitlab.com/myowner/myrepo",
		},
		{
			name:        "gitlab URL with .git suffix",
			url:         "https://gitlab.com/myowner/myrepo.git",
			expectOwner: "myowner",
			expectRepo:  "myrepo",
			expectType:  "gitlab",
			expectURL:   "https://gitlab.com/myowner/myrepo.git",
		},
		{
			name:        "gitlab URL with subpath",
			url:         "https://gitlab.com/myowner/myrepo/sub/path",
			expectOwner: "myowner",
			expectRepo:  "myrepo",
			expectType:  "gitlab",
			expectURL:   "https://gitlab.com/myowner/myrepo/sub/path",
		},
		{
			name:       "gitlab URL with only owner",
			url:        "https://gitlab.com/onlyowner",
			expectType: "gitlab",
			expectURL:  "https://gitlab.com/onlyowner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := parseGitLabURL(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if source.Type != tt.expectType {
				t.Errorf("Type = %q, expected %q", source.Type, tt.expectType)
			}
			if source.URL != tt.expectURL {
				t.Errorf("URL = %q, expected %q", source.URL, tt.expectURL)
			}
			if tt.expectOwner != "" && source.Owner != tt.expectOwner {
				t.Errorf("Owner = %q, expected %q", source.Owner, tt.expectOwner)
			}
			if tt.expectRepo != "" && source.Repo != tt.expectRepo {
				t.Errorf("Repo = %q, expected %q", source.Repo, tt.expectRepo)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for parseGitHubURL edge cases (install.go)
// ---------------------------------------------------------------------------

func TestParseGitHubURLEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectOwner   string
		expectRepo    string
		expectRef     string
		expectSubpath string
	}{
		{
			name:        "basic github URL",
			url:         "https://github.com/owner/repo",
			expectOwner: "owner",
			expectRepo:  "repo",
		},
		{
			name:        "github URL with .git suffix",
			url:         "https://github.com/owner/repo.git",
			expectOwner: "owner",
			expectRepo:  "repo",
		},
		{
			name:          "github URL with tree/branch",
			url:           "https://github.com/owner/repo/tree/main",
			expectOwner:   "owner",
			expectRepo:    "repo",
			expectRef:     "main",
			expectSubpath: "",
		},
		{
			name:          "github URL with tree/branch/path",
			url:           "https://github.com/owner/repo/tree/develop/skills/react",
			expectOwner:   "owner",
			expectRepo:    "repo",
			expectRef:     "develop",
			expectSubpath: "skills/react",
		},
		{
			name:          "github URL with subpath no tree",
			url:           "https://github.com/owner/repo/skills/react",
			expectOwner:   "owner",
			expectRepo:    "repo",
			expectSubpath: "skills/react",
		},
		{
			name:          "github URL tree with nested path",
			url:           "https://github.com/owner/repo/tree/v1.0/deep/nested/path",
			expectOwner:   "owner",
			expectRepo:    "repo",
			expectRef:     "v1.0",
			expectSubpath: "deep/nested/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := parseGitHubURL(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if source.Type != "github" {
				t.Errorf("Type = %q, expected %q", source.Type, "github")
			}
			if source.Owner != tt.expectOwner {
				t.Errorf("Owner = %q, expected %q", source.Owner, tt.expectOwner)
			}
			if source.Repo != tt.expectRepo {
				t.Errorf("Repo = %q, expected %q", source.Repo, tt.expectRepo)
			}
			if source.Ref != tt.expectRef {
				t.Errorf("Ref = %q, expected %q", source.Ref, tt.expectRef)
			}
			if source.Subpath != tt.expectSubpath {
				t.Errorf("Subpath = %q, expected %q", source.Subpath, tt.expectSubpath)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for checkOutputJSON (check.go)
// ---------------------------------------------------------------------------

func TestCheckOutputJSON(t *testing.T) {
	tests := []struct {
		name            string
		results         []CheckResult
		expectOutdated  int
		expectUpToDate  int
		expectErrors    int
		expectTotal     int
		expectSubstring string
	}{
		{
			name:           "empty results",
			results:        []CheckResult{},
			expectOutdated: 0,
			expectUpToDate: 0,
			expectErrors:   0,
			expectTotal:    0,
		},
		{
			name: "mixed results",
			results: []CheckResult{
				{Name: "skill-a", NeedsUpdate: true, CurrentHash: "hash1", RemoteHash: "hash2"},
				{Name: "skill-b", NeedsUpdate: false, CurrentHash: "hash3", RemoteHash: "hash3"},
				{Name: "skill-c", Error: "some error"},
			},
			expectOutdated: 1,
			expectUpToDate: 1,
			expectErrors:   1,
			expectTotal:    3,
		},
		{
			name: "all up-to-date",
			results: []CheckResult{
				{Name: "skill-x", NeedsUpdate: false, CurrentHash: "aaa", RemoteHash: "aaa"},
				{Name: "skill-y", NeedsUpdate: false, CurrentHash: "bbb", RemoteHash: "bbb"},
			},
			expectOutdated: 0,
			expectUpToDate: 2,
			expectErrors:   0,
			expectTotal:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := checkOutputJSON(tt.results)

			_ = w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Parse the JSON to verify structure
			var parsed struct {
				Results []CheckResult `json:"results"`
				Summary struct {
					Total    int `json:"total"`
					Outdated int `json:"outdated"`
					UpToDate int `json:"upToDate"`
					Errors   int `json:"errors"`
				} `json:"summary"`
			}
			if err := json.Unmarshal([]byte(output), &parsed); err != nil {
				t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, output)
			}

			if parsed.Summary.Total != tt.expectTotal {
				t.Errorf("summary.total = %d, expected %d", parsed.Summary.Total, tt.expectTotal)
			}
			if parsed.Summary.Outdated != tt.expectOutdated {
				t.Errorf("summary.outdated = %d, expected %d", parsed.Summary.Outdated, tt.expectOutdated)
			}
			if parsed.Summary.UpToDate != tt.expectUpToDate {
				t.Errorf("summary.upToDate = %d, expected %d", parsed.Summary.UpToDate, tt.expectUpToDate)
			}
			if parsed.Summary.Errors != tt.expectErrors {
				t.Errorf("summary.errors = %d, expected %d", parsed.Summary.Errors, tt.expectErrors)
			}
			if len(parsed.Results) != len(tt.results) {
				t.Errorf("results length = %d, expected %d", len(parsed.Results), len(tt.results))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for checkOutputTable (check.go)
// ---------------------------------------------------------------------------

func TestCheckOutputTable(t *testing.T) {
	// Save and restore quiet flag
	origQuiet := quiet
	defer func() { quiet = origQuiet }()

	t.Run("table with mixed results", func(t *testing.T) {
		quiet = false
		results := []CheckResult{
			{Name: "skill-a", NeedsUpdate: true, CurrentHash: "oldhash", RemoteHash: "newhash"},
			{Name: "skill-b", NeedsUpdate: false, CurrentHash: "samehash", RemoteHash: "samehash"},
			{Name: "skill-c", Error: "failed to fetch"},
		}

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := checkOutputTable(results)

		_ = w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify header
		if !strings.Contains(output, "NAME") || !strings.Contains(output, "STATUS") {
			t.Error("expected table header with NAME and STATUS")
		}
		// Verify skill names appear
		if !strings.Contains(output, "skill-a") {
			t.Error("expected skill-a in output")
		}
		if !strings.Contains(output, "skill-b") {
			t.Error("expected skill-b in output")
		}
		if !strings.Contains(output, "skill-c") {
			t.Error("expected skill-c in output")
		}
		// Verify statuses
		if !strings.Contains(output, "outdated") {
			t.Error("expected 'outdated' status in output")
		}
		if !strings.Contains(output, "up-to-date") {
			t.Error("expected 'up-to-date' status in output")
		}
		if !strings.Contains(output, "error:") {
			t.Error("expected 'error:' status in output")
		}
		// Verify summary
		if !strings.Contains(output, "3 skill(s) checked") {
			t.Errorf("expected summary line, got: %s", output)
		}
	})

	t.Run("table in quiet mode", func(t *testing.T) {
		quiet = true
		results := []CheckResult{
			{Name: "skill-a", NeedsUpdate: false, CurrentHash: "x", RemoteHash: "x"},
		}

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		_ = checkOutputTable(results)

		_ = w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		// In quiet mode, the summary line should not appear
		if strings.Contains(output, "skill(s) checked") {
			t.Error("quiet mode should suppress summary line")
		}
	})
}

// ---------------------------------------------------------------------------
// Tests for listSkillsJSON (list.go)
// ---------------------------------------------------------------------------

func TestListSkillsJSON(t *testing.T) {
	skills := []scribe.SkillInfo{
		{
			Name:        "react-patterns",
			Description: "React best practices",
			Source:      "owner/repo",
			SourceType:  "github",
			InstalledAt: "2024-06-15T10:00:00Z",
			Agents:      []string{"claude-code"},
		},
		{
			Name:        "go-idioms",
			Description: "Go idiomatic patterns",
			Source:      "/local/path",
			SourceType:  "local",
			InstalledAt: "2024-07-01T12:00:00Z",
			Agents:      []string{"cursor", "claude-code"},
		},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listSkillsJSON(skills)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Skills []struct {
			Name        string   `json:"name"`
			Description string   `json:"description"`
			Source      string   `json:"source"`
			SourceType  string   `json:"sourceType"`
			InstalledAt string   `json:"installedAt"`
			Agents      []string `json:"agents"`
		} `json:"skills"`
		Count int `json:"count"`
	}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
	}
	if parsed.Count != 2 {
		t.Errorf("count = %d, expected 2", parsed.Count)
	}
	if len(parsed.Skills) != 2 {
		t.Errorf("skills length = %d, expected 2", len(parsed.Skills))
	}
	if parsed.Skills[0].Name != "react-patterns" {
		t.Errorf("first skill name = %q, expected %q", parsed.Skills[0].Name, "react-patterns")
	}
	if parsed.Skills[1].SourceType != "local" {
		t.Errorf("second skill sourceType = %q, expected %q", parsed.Skills[1].SourceType, "local")
	}
}

func TestListSkillsJSONEmpty(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listSkillsJSON([]scribe.SkillInfo{})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Skills []interface{} `json:"skills"`
		Count  int           `json:"count"`
	}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if parsed.Count != 0 {
		t.Errorf("count = %d, expected 0", parsed.Count)
	}
	// skills should be an empty array, not null
	if parsed.Skills == nil {
		t.Error("expected empty array for skills, got null")
	}
}

// ---------------------------------------------------------------------------
// Tests for listSkillsNamesOnly (list.go)
// ---------------------------------------------------------------------------

func TestListSkillsNamesOnly(t *testing.T) {
	skills := []scribe.SkillInfo{
		{Name: "alpha-skill"},
		{Name: "beta-skill"},
		{Name: "gamma-skill"},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listSkillsNamesOnly(skills)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "alpha-skill" {
		t.Errorf("line 0 = %q, expected %q", lines[0], "alpha-skill")
	}
	if lines[1] != "beta-skill" {
		t.Errorf("line 1 = %q, expected %q", lines[1], "beta-skill")
	}
	if lines[2] != "gamma-skill" {
		t.Errorf("line 2 = %q, expected %q", lines[2], "gamma-skill")
	}
}

func TestListSkillsNamesOnlyEmpty(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listSkillsNamesOnly([]scribe.SkillInfo{})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "" {
		t.Errorf("expected empty output, got %q", output)
	}
}

// ---------------------------------------------------------------------------
// Tests for listSkillsTable (list.go)
// ---------------------------------------------------------------------------

func TestListSkillsTableWithSkills(t *testing.T) {
	origQuiet := quiet
	defer func() { quiet = origQuiet }()
	quiet = false

	skills := []scribe.SkillInfo{
		{
			Name:        "react-patterns",
			Description: "React best practices and patterns",
			Source:      "owner/repo",
			SourceType:  "github",
			InstalledAt: "2024-06-15T10:00:00Z",
			Agents:      []string{"claude-code", "cursor"},
		},
		{
			Name:        "go-idioms",
			Description: "Go idiomatic patterns",
			Source:      "/local/path",
			SourceType:  "local",
			InstalledAt: "2024-07-01",
			Agents:      []string{},
		},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listSkillsTable(skills)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check header
	if !strings.Contains(output, "NAME") {
		t.Error("expected NAME header")
	}
	if !strings.Contains(output, "DESCRIPTION") {
		t.Error("expected DESCRIPTION header")
	}
	if !strings.Contains(output, "SOURCE") {
		t.Error("expected SOURCE header")
	}
	// Check skill data
	if !strings.Contains(output, "react-patterns") {
		t.Error("expected react-patterns in output")
	}
	if !strings.Contains(output, "go-idioms") {
		t.Error("expected go-idioms in output")
	}
	// Check installed count
	if !strings.Contains(output, "2 skill(s) installed") {
		t.Errorf("expected '2 skill(s) installed' in output, got: %s", output)
	}
}

func TestListSkillsTableEmpty(t *testing.T) {
	origQuiet := quiet
	defer func() { quiet = origQuiet }()
	quiet = false

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listSkillsTable([]scribe.SkillInfo{})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "No skills installed") {
		t.Errorf("expected 'No skills installed', got: %s", output)
	}
}

func TestListSkillsTableEmptyQuiet(t *testing.T) {
	origQuiet := quiet
	defer func() { quiet = origQuiet }()
	quiet = true

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listSkillsTable([]scribe.SkillInfo{})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(output, "No skills installed") {
		t.Error("quiet mode should suppress 'No skills installed' message")
	}
}

// ---------------------------------------------------------------------------
// Tests for copyFile and copySkillDir (update.go)
// ---------------------------------------------------------------------------

func TestCopyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copyfile-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "dest.txt")

	content := "Hello, this is test content for copyFile."
	if err := os.WriteFile(srcPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// Verify content
	data, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}
	if string(data) != content {
		t.Errorf("copied content = %q, expected %q", string(data), content)
	}

	// Verify permissions match
	srcInfo, _ := os.Stat(srcPath)
	dstInfo, _ := os.Stat(dstPath)
	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("file mode mismatch: src=%v, dst=%v", srcInfo.Mode(), dstInfo.Mode())
	}
}

func TestCopyFileNonExistentSource(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copyfile-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	err = copyFile(filepath.Join(tmpDir, "nonexistent"), filepath.Join(tmpDir, "dest"))
	if err == nil {
		t.Error("expected error copying nonexistent file, got nil")
	}
}

func TestCopySkillDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copydir-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcDir := filepath.Join(tmpDir, "source-skill")
	dstDir := filepath.Join(tmpDir, "dest-skill")

	// Create source directory structure
	if err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0o755); err != nil {
		t.Fatalf("failed to create source dirs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("# Test Skill"), 0o644); err != nil {
		t.Fatalf("failed to write SKILL.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "subdir", "extra.txt"), []byte("extra content"), 0o644); err != nil {
		t.Fatalf("failed to write extra.txt: %v", err)
	}

	if err := copySkillDir(srcDir, dstDir); err != nil {
		t.Fatalf("copySkillDir failed: %v", err)
	}

	// Verify SKILL.md was copied
	data, err := os.ReadFile(filepath.Join(dstDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("failed to read copied SKILL.md: %v", err)
	}
	if string(data) != "# Test Skill" {
		t.Errorf("SKILL.md content = %q, expected %q", string(data), "# Test Skill")
	}

	// Verify subdirectory and file were copied
	data, err = os.ReadFile(filepath.Join(dstDir, "subdir", "extra.txt"))
	if err != nil {
		t.Fatalf("failed to read copied subdir/extra.txt: %v", err)
	}
	if string(data) != "extra content" {
		t.Errorf("extra.txt content = %q, expected %q", string(data), "extra content")
	}
}

func TestCopySkillDirOverwrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copydir-overwrite-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcDir := filepath.Join(tmpDir, "source-skill")
	dstDir := filepath.Join(tmpDir, "dest-skill")

	// Create source
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("new content"), 0o644); err != nil {
		t.Fatalf("failed to write source SKILL.md: %v", err)
	}

	// Create pre-existing destination with old content
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dstDir, "SKILL.md"), []byte("old content"), 0o644); err != nil {
		t.Fatalf("failed to write dest SKILL.md: %v", err)
	}

	// copySkillDir removes existing dest first
	if err := copySkillDir(srcDir, dstDir); err != nil {
		t.Fatalf("copySkillDir failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dstDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("failed to read copied SKILL.md: %v", err)
	}
	if string(data) != "new content" {
		t.Errorf("content = %q, expected %q", string(data), "new content")
	}
}

// ---------------------------------------------------------------------------
// Tests for formatSourceInfo edge cases (install.go)
// ---------------------------------------------------------------------------

func TestFormatSourceInfoEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		source   *scribe.SourceInfo
		expected string
	}{
		{
			name: "github with subpath",
			source: &scribe.SourceInfo{
				Type:    "github",
				Owner:   "user",
				Repo:    "repo",
				Subpath: "skills/react",
			},
			expected: "github:user/repo/skills/react",
		},
		{
			name: "github with ref and subpath",
			source: &scribe.SourceInfo{
				Type:    "github",
				Owner:   "user",
				Repo:    "repo",
				Ref:     "v2.0",
				Subpath: "skills/react",
			},
			expected: "github:user/repo#v2.0/skills/react",
		},
		{
			name: "well-known source",
			source: &scribe.SourceInfo{
				Type: "well-known",
				URL:  "https://example.com/.well-known/skills",
			},
			expected: "https://example.com/.well-known/skills",
		},
		{
			name: "unknown type falls through to URL",
			source: &scribe.SourceInfo{
				Type: "other",
				URL:  "https://custom.example.com/skills",
			},
			expected: "https://custom.example.com/skills",
		},
		{
			name: "gitlab source",
			source: &scribe.SourceInfo{
				Type:  "gitlab",
				Owner: "gluser",
				Repo:  "glrepo",
			},
			expected: "gitlab:gluser/glrepo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSourceInfo(tt.source)
			if result != tt.expected {
				t.Errorf("formatSourceInfo() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for runUninstall (uninstall.go)
// ---------------------------------------------------------------------------

func TestRunUninstallNoArgs(t *testing.T) {
	origQuiet := quiet
	origUninstallAll := uninstallAll
	defer func() {
		quiet = origQuiet
		uninstallAll = origUninstallAll
	}()
	quiet = true
	uninstallAll = false

	err := runUninstall(uninstallCmd, []string{})
	if err == nil {
		t.Fatal("expected error when no args and --all not set")
	}
	if !strings.Contains(err.Error(), "skill name is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunUninstallNonexistentSkill(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-uninstall-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	scribe.InitLoggerCLI(false)
	_ = scribe.EnsureScribeDirs()

	origQuiet := quiet
	origUninstallAll := uninstallAll
	defer func() {
		quiet = origQuiet
		uninstallAll = origUninstallAll
	}()
	quiet = true
	uninstallAll = false

	err = runUninstall(uninstallCmd, []string{"nonexistent-skill"})
	if err == nil {
		t.Fatal("expected error for nonexistent skill")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Tests for infoJSON and infoTable (info.go)
// ---------------------------------------------------------------------------

func TestInfoJSON(t *testing.T) {
	skill := &scribe.Skill{
		Name:        "test-skill",
		Description: "A test skill",
		Metadata:    map[string]any{"version": "1.0"},
		Meta: &scribe.SkillMeta{
			Source:      "owner/repo",
			SourceType:  "github",
			SourceURL:   "https://github.com/owner/repo",
			SkillPath:   "skills/test",
			ContentHash: "sha256:abc123",
			InstalledAt: "2024-06-15T10:00:00Z",
			UpdatedAt:   "2024-07-01T12:00:00Z",
		},
	}
	agents := []string{"claude-code", "cursor"}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := infoJSON(skill, agents)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed skillDetailJSON
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
	}
	if parsed.Name != "test-skill" {
		t.Errorf("name = %q, expected %q", parsed.Name, "test-skill")
	}
	if parsed.Description != "A test skill" {
		t.Errorf("description = %q, expected %q", parsed.Description, "A test skill")
	}
	if parsed.Source != "owner/repo" {
		t.Errorf("source = %q, expected %q", parsed.Source, "owner/repo")
	}
	if parsed.SourceType != "github" {
		t.Errorf("sourceType = %q, expected %q", parsed.SourceType, "github")
	}
	if parsed.SourceURL != "https://github.com/owner/repo" {
		t.Errorf("sourceUrl = %q, expected %q", parsed.SourceURL, "https://github.com/owner/repo")
	}
	if parsed.ContentHash != "sha256:abc123" {
		t.Errorf("contentHash = %q, expected %q", parsed.ContentHash, "sha256:abc123")
	}
	if parsed.InstalledAt != "2024-06-15T10:00:00Z" {
		t.Errorf("installedAt = %q, expected %q", parsed.InstalledAt, "2024-06-15T10:00:00Z")
	}
	if parsed.UpdatedAt != "2024-07-01T12:00:00Z" {
		t.Errorf("updatedAt = %q, expected %q", parsed.UpdatedAt, "2024-07-01T12:00:00Z")
	}
	if len(parsed.Agents) != 2 {
		t.Errorf("agents count = %d, expected 2", len(parsed.Agents))
	}
}

func TestInfoJSONNoMeta(t *testing.T) {
	skill := &scribe.Skill{
		Name:        "bare-skill",
		Description: "No metadata",
		Meta:        nil,
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := infoJSON(skill, []string{})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed skillDetailJSON
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if parsed.Name != "bare-skill" {
		t.Errorf("name = %q, expected %q", parsed.Name, "bare-skill")
	}
	if parsed.Source != "" {
		t.Errorf("source should be empty, got %q", parsed.Source)
	}
	if parsed.ContentHash != "" {
		t.Errorf("contentHash should be empty, got %q", parsed.ContentHash)
	}
}

func TestInfoTable(t *testing.T) {
	skill := &scribe.Skill{
		Name:        "test-skill",
		Description: "A test skill",
		Meta: &scribe.SkillMeta{
			Source:      "owner/repo",
			SourceType:  "github",
			SourceURL:   "https://github.com/owner/repo",
			ContentHash: "sha256:abc123",
			InstalledAt: "2024-06-15T10:00:00Z",
			UpdatedAt:   "2024-07-01T12:00:00Z",
		},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := infoTable(skill, []string{"claude-code"})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Name:") {
		t.Error("expected 'Name:' in output")
	}
	if !strings.Contains(output, "test-skill") {
		t.Error("expected 'test-skill' in output")
	}
	if !strings.Contains(output, "Source:") {
		t.Error("expected 'Source:' in output")
	}
	if !strings.Contains(output, "owner/repo") {
		t.Error("expected 'owner/repo' in output")
	}
	if !strings.Contains(output, "Agents:") {
		t.Error("expected 'Agents:' in output")
	}
	if !strings.Contains(output, "claude-code") {
		t.Error("expected 'claude-code' in output")
	}
	if !strings.Contains(output, "Installed:") {
		t.Error("expected 'Installed:' in output")
	}
	if !strings.Contains(output, "2024-06-15") {
		t.Error("expected formatted date in output")
	}
}

func TestInfoTableNoAgents(t *testing.T) {
	skill := &scribe.Skill{
		Name:        "lonely-skill",
		Description: "No agents",
		Meta:        nil,
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := infoTable(skill, []string{})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "(none)") {
		t.Error("expected '(none)' for agents when no agents present")
	}
}

// ---------------------------------------------------------------------------
// Tests for workspace commands (workspace.go)
// ---------------------------------------------------------------------------

func setupTempHome(t *testing.T) (dir string, cleanup func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "scribe-ws-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	scribe.InitLoggerCLI(false)
	_ = scribe.EnsureScribeDirs()

	cleanup = func() {
		_ = os.Setenv("HOME", origHome)
		_ = os.RemoveAll(tmpDir)
	}
	return tmpDir, cleanup
}

func TestRunWorkspaceList(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	origQuiet := quiet
	origJSON := jsonOutput
	defer func() {
		quiet = origQuiet
		jsonOutput = origJSON
	}()
	quiet = false
	jsonOutput = false

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runWorkspaceList(workspaceListCmd, []string{})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Workspaces") {
		t.Errorf("expected 'Workspaces' heading, got: %s", output)
	}
	if !strings.Contains(output, "default") {
		t.Errorf("expected 'default' workspace in listing, got: %s", output)
	}
}

func TestRunWorkspaceListJSON(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	origJSON := jsonOutput
	defer func() { jsonOutput = origJSON }()
	jsonOutput = true

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runWorkspaceList(workspaceListCmd, []string{})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed []scribe.WorkspaceInfo
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
	}
	if len(parsed) == 0 {
		t.Error("expected at least one workspace (default)")
	}

	foundDefault := false
	for _, ws := range parsed {
		if ws.Name == "default" {
			foundDefault = true
			break
		}
	}
	if !foundDefault {
		t.Error("expected default workspace in JSON output")
	}
}

func TestRunWorkspaceCreate(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	origQuiet := quiet
	defer func() { quiet = origQuiet }()
	quiet = false

	// Reset the flag variable
	workspaceDescription = ""

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runWorkspaceCreate(workspaceCreateCmd, []string{"test-workspace"})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "test-workspace") {
		t.Errorf("expected 'test-workspace' in output, got: %s", output)
	}
	if !strings.Contains(output, "created") {
		t.Errorf("expected 'created' in output, got: %s", output)
	}

	// Verify it actually exists
	ws, err := scribe.GetWorkspace("test-workspace")
	if err != nil {
		t.Fatalf("failed to get workspace: %v", err)
	}
	if ws.Name != "test-workspace" {
		t.Errorf("workspace name = %q, expected %q", ws.Name, "test-workspace")
	}
}

func TestRunWorkspaceCreateDuplicate(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	origQuiet := quiet
	defer func() { quiet = origQuiet }()
	quiet = true

	workspaceDescription = ""

	// Create first time
	err := runWorkspaceCreate(workspaceCreateCmd, []string{"dup-workspace"})
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	// Create again should fail
	err = runWorkspaceCreate(workspaceCreateCmd, []string{"dup-workspace"})
	if err == nil {
		t.Error("expected error creating duplicate workspace")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestRunWorkspaceCurrent(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	origQuiet := quiet
	origJSON := jsonOutput
	defer func() {
		quiet = origQuiet
		jsonOutput = origJSON
	}()
	quiet = false
	jsonOutput = false

	// Ensure default workspace is saved
	_ = scribe.EnsureDefaultWorkspace()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runWorkspaceCurrent(workspaceCurrentCmd, []string{})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Active workspace:") {
		t.Errorf("expected 'Active workspace:' in output, got: %s", output)
	}
	if !strings.Contains(output, "default") {
		t.Errorf("expected 'default' workspace, got: %s", output)
	}
}

func TestRunWorkspaceCurrentJSON(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	origJSON := jsonOutput
	defer func() { jsonOutput = origJSON }()
	jsonOutput = true

	_ = scribe.EnsureDefaultWorkspace()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runWorkspaceCurrent(workspaceCurrentCmd, []string{})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed scribe.WorkspaceInfo
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
	}
	if parsed.Name != "default" {
		t.Errorf("workspace name = %q, expected %q", parsed.Name, "default")
	}
	if !parsed.IsActive {
		t.Error("expected isActive to be true")
	}
}

func TestRunWorkspaceDelete(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	origQuiet := quiet
	defer func() { quiet = origQuiet }()
	quiet = false

	workspaceDescription = ""

	// Create a workspace to delete
	err := runWorkspaceCreate(workspaceCreateCmd, []string{"to-delete"})
	if err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runWorkspaceDelete(workspaceDeleteCmd, []string{"to-delete"})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", output)
	}
}

func TestRunWorkspaceDeleteDefault(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	origQuiet := quiet
	defer func() { quiet = origQuiet }()
	quiet = true

	_ = scribe.EnsureDefaultWorkspace()

	err := runWorkspaceDelete(workspaceDeleteCmd, []string{"default"})
	if err == nil {
		t.Error("expected error deleting default workspace")
	}
	if !strings.Contains(err.Error(), "cannot delete") {
		t.Errorf("expected 'cannot delete' error, got: %v", err)
	}
}

// ===========================================================================
// Helper: install a fake skill into the temp HOME's scrolls directory.
// Returns the skill directory path.
// ===========================================================================

func installFakeSkill(t *testing.T, name, description, sourceType, source string) string {
	t.Helper()
	scrollsDir, err := scribe.GetScrollsDir()
	if err != nil {
		t.Fatalf("failed to get scrolls dir: %v", err)
	}
	skillDir := filepath.Join(scrollsDir, name)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}

	content := fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n\n# %s\n\nTest skill content.\n", name, description, name)
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write SKILL.md: %v", err)
	}

	meta := scribe.NewSkillMeta(&scribe.SourceInfo{
		Type:  sourceType,
		Owner: "testowner",
		Repo:  "testrepo",
		URL:   "https://github.com/testowner/testrepo",
	}, "", content)
	meta.Source = source
	meta.SourceType = sourceType

	metaPath := filepath.Join(skillDir, ".scribe-meta.json")
	if err := scribe.WriteSkillMeta(metaPath, meta); err != nil {
		t.Fatalf("failed to write meta: %v", err)
	}

	return skillDir
}

// captureStdout runs fn() and captures whatever it writes to os.Stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

// saveAndRestoreFlags saves CLI flag globals and restores them in the cleanup function.
func saveAndRestoreFlags(t *testing.T) {
	t.Helper()
	origQuiet := quiet
	origJSON := jsonOutput
	origNamesOnly := namesOnly
	origInstallAgents := installAgents
	origInstallSkills := installSkills
	origInstallListOnly := installListOnly
	origInstallYes := installYes
	origInstallAll := installAll
	origUninstallAll := uninstallAll
	origUpdateForce := updateForce
	origDebug := debug
	origWsDesc := workspaceDescription

	t.Cleanup(func() {
		quiet = origQuiet
		jsonOutput = origJSON
		namesOnly = origNamesOnly
		installAgents = origInstallAgents
		installSkills = origInstallSkills
		installListOnly = origInstallListOnly
		installYes = origInstallYes
		installAll = origInstallAll
		uninstallAll = origUninstallAll
		updateForce = origUpdateForce
		debug = origDebug
		workspaceDescription = origWsDesc
	})
}

// ===========================================================================
// Tests for Execute (root.go:72)
// ===========================================================================

func TestExecuteVersionCommand(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	// Mark onboarding as complete so it doesn't interfere
	_ = scribe.CompleteOnboarding()

	// Execute with version command - should succeed
	os.Args = []string{"scribe", "version"}
	output := captureStdout(t, func() {
		code := Execute()
		if code != ExitSuccess {
			t.Errorf("Execute() returned %d, expected %d", code, ExitSuccess)
		}
	})
	if !strings.Contains(output, "scribe version") {
		t.Errorf("expected version output, got: %s", output)
	}
}

func TestExecuteInvalidCommand(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	_ = scribe.CompleteOnboarding()

	os.Args = []string{"scribe", "nonexistent-command"}
	quiet = false
	code := Execute()
	if code != ExitError {
		t.Errorf("Execute() returned %d, expected %d for invalid command", code, ExitError)
	}
}

func TestExecuteQuietMode(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	_ = scribe.CompleteOnboarding()

	os.Args = []string{"scribe", "--quiet", "nonexistent-command"}
	quiet = true
	output := captureStdout(t, func() {
		_ = Execute()
	})
	// In quiet mode stderr output should be suppressed
	// stdout should have nothing from Execute itself
	_ = output // we mainly care that it doesn't panic
}

// ===========================================================================
// Tests for checkOnboarding (onboarding.go:29)
// ===========================================================================

func TestCheckOnboardingNotCompleted(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	// Fresh HOME, onboarding not completed
	needed := checkOnboarding()
	if !needed {
		t.Error("expected checkOnboarding() to return true for fresh install")
	}
}

func TestCheckOnboardingCompleted(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	// Complete onboarding
	_ = scribe.CompleteOnboarding()

	needed := checkOnboarding()
	if needed {
		t.Error("expected checkOnboarding() to return false after completion")
	}
}

// ===========================================================================
// Tests for runOnboardingIfNeeded (onboarding.go:38)
// ===========================================================================

func TestRunOnboardingIfNeededAlreadyComplete(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	_ = scribe.CompleteOnboarding()

	err := runOnboardingIfNeeded()
	if err != nil {
		t.Errorf("expected nil error when onboarding already complete, got: %v", err)
	}
}

// ===========================================================================
// Tests for runCheck (check.go:42) and checkSkill (check.go:80)
// ===========================================================================

func TestCheckSkillNonexistent(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	result := checkSkill("nonexistent-skill")
	if result.Error == "" {
		t.Error("expected error for nonexistent skill")
	}
	if !strings.Contains(result.Error, "failed to read skill") {
		t.Errorf("unexpected error: %s", result.Error)
	}
}

func TestCheckSkillNoMeta(t *testing.T) {
	tmpDir, cleanup := setupTempHome(t)
	defer cleanup()
	_ = tmpDir

	// Install a skill without metadata
	scrollsDir, _ := scribe.GetScrollsDir()
	skillDir := filepath.Join(scrollsDir, "no-meta-skill")
	_ = os.MkdirAll(skillDir, 0o755)
	content := "---\nname: no-meta-skill\ndescription: A test skill\n---\n\nContent here.\n"
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644)
	// No .scribe-meta.json file

	result := checkSkill("no-meta-skill")
	if result.Error == "" {
		t.Error("expected error for skill without metadata")
	}
	if !strings.Contains(result.Error, "no metadata") {
		t.Errorf("unexpected error: %s", result.Error)
	}
}

func TestCheckSkillLocalSource(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	installFakeSkill(t, "local-skill", "A local skill", "local", "/some/local/path")

	result := checkSkill("local-skill")
	if result.Error == "" {
		t.Error("expected error for local source skill")
	}
	if !strings.Contains(result.Error, "local source") {
		t.Errorf("unexpected error: %s", result.Error)
	}
}

func TestRunCheckNoSkillsInstalled(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = false

	output := captureStdout(t, func() {
		err := runCheck(checkCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "No skills installed") {
		t.Errorf("expected 'No skills installed' message, got: %s", output)
	}
}

func TestRunCheckNoSkillsQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	jsonOutput = false

	output := captureStdout(t, func() {
		err := runCheck(checkCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "No skills installed") {
		t.Error("quiet mode should suppress 'No skills installed' message")
	}
}

func TestRunCheckSingleSkillNonexistent(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	jsonOutput = false

	output := captureStdout(t, func() {
		err := runCheck(checkCmd, []string{"nonexistent"})
		if err != nil {
			t.Errorf("unexpected error from runCheck: %v", err)
		}
	})

	// Should show an error in the table for the nonexistent skill
	if !strings.Contains(output, "nonexistent") {
		t.Errorf("expected skill name in output, got: %s", output)
	}
}

func TestRunCheckSingleSkillJSON(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	jsonOutput = true

	// Install a local skill (which can't be checked remotely)
	installFakeSkill(t, "check-me", "Check me", "local", "/local/path")

	output := captureStdout(t, func() {
		err := runCheck(checkCmd, []string{"check-me"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	var parsed struct {
		Results []CheckResult `json:"results"`
		Summary struct {
			Total  int `json:"total"`
			Errors int `json:"errors"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
	}
	if parsed.Summary.Total != 1 {
		t.Errorf("expected total 1, got %d", parsed.Summary.Total)
	}
	if parsed.Summary.Errors != 1 {
		t.Errorf("expected 1 error (local source), got %d", parsed.Summary.Errors)
	}
}

func TestRunCheckAllSkillsWithLocalSkill(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = false

	installFakeSkill(t, "skill-alpha", "Alpha skill", "local", "/path/alpha")
	installFakeSkill(t, "skill-beta", "Beta skill", "local", "/path/beta")

	output := captureStdout(t, func() {
		err := runCheck(checkCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Checking 2 skill(s)") {
		t.Errorf("expected 'Checking 2 skill(s)' in output, got: %s", output)
	}
	if !strings.Contains(output, "skill-alpha") {
		t.Errorf("expected 'skill-alpha' in output, got: %s", output)
	}
	if !strings.Contains(output, "skill-beta") {
		t.Errorf("expected 'skill-beta' in output, got: %s", output)
	}
}

func TestRunCheckAllSkillsJSON(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	jsonOutput = true

	installFakeSkill(t, "json-skill", "JSON check skill", "local", "/local/path")

	output := captureStdout(t, func() {
		err := runCheck(checkCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	var parsed struct {
		Results []CheckResult `json:"results"`
		Summary struct {
			Total int `json:"total"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON output: %v\n%s", err, output)
	}
	if parsed.Summary.Total != 1 {
		t.Errorf("expected 1 skill, got %d", parsed.Summary.Total)
	}
}

// ===========================================================================
// Tests for runInstall (install.go:55)
// ===========================================================================

func TestRunInstallFromLocalDir(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	installYes = true
	installAll = false
	installListOnly = false
	installAgents = ""
	installSkills = ""

	// Create a local skill source directory
	tmpSrc, err := os.MkdirTemp("", "scribe-install-src-*")
	if err != nil {
		t.Fatalf("failed to create temp source dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	skillContent := "---\nname: local-install-test\ndescription: Test skill for local install\n---\n\n# Local Install Test\n\nHello.\n"
	_ = os.WriteFile(filepath.Join(tmpSrc, "SKILL.md"), []byte(skillContent), 0o644)

	output := captureStdout(t, func() {
		err = runInstall(installCmd, []string{tmpSrc})
	})

	// Install should fail because no agents detected in temp HOME
	if err != nil {
		if !strings.Contains(err.Error(), "no coding agents detected") {
			t.Errorf("unexpected error: %v", err)
		}
		return
	}

	// If we got here, install succeeded (some agent dirs exist)
	_ = output
	exists, _ := scribe.SkillExists("local-install-test")
	if !exists {
		t.Error("expected skill to be installed")
	}
}

func TestRunInstallListOnly(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	installYes = false
	installAll = false
	installListOnly = true
	installAgents = ""
	installSkills = ""

	// Create a local skill source directory
	tmpSrc, err := os.MkdirTemp("", "scribe-install-list-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	skillContent := "---\nname: list-only-test\ndescription: List only test skill\n---\n\n# Test\n"
	_ = os.WriteFile(filepath.Join(tmpSrc, "SKILL.md"), []byte(skillContent), 0o644)

	output := captureStdout(t, func() {
		err = runInstall(installCmd, []string{tmpSrc})
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Found 1 skill(s)") {
		t.Errorf("expected 'Found 1 skill(s)' in output, got: %s", output)
	}
	if !strings.Contains(output, "list-only-test") {
		t.Errorf("expected skill name in output, got: %s", output)
	}

	// Skill should NOT be installed in list-only mode
	exists, _ := scribe.SkillExists("list-only-test")
	if exists {
		t.Error("skill should not be installed in list-only mode")
	}
}

func TestRunInstallInvalidSource(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	installListOnly = false

	err := runInstall(installCmd, []string{"invalid"})
	if err == nil {
		t.Error("expected error for invalid source")
	}
	if !strings.Contains(err.Error(), "invalid source") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunInstallNoSkillsFound(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	installListOnly = false

	// Create an empty local source directory
	tmpSrc, err := os.MkdirTemp("", "scribe-install-empty-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	err = runInstall(installCmd, []string{tmpSrc})
	if err == nil {
		t.Error("expected error when no skills found")
	}
	if !strings.Contains(err.Error(), "no skills found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunInstallFilterSkills(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	installYes = true
	installListOnly = false
	installSkills = "nonexistent-skill"
	installAgents = ""
	installAll = false

	// Create a local skill source directory with a skill
	tmpSrc, err := os.MkdirTemp("", "scribe-install-filter-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	skillContent := "---\nname: real-skill\ndescription: A real skill\n---\n\n# Real\n"
	_ = os.WriteFile(filepath.Join(tmpSrc, "SKILL.md"), []byte(skillContent), 0o644)

	err = runInstall(installCmd, []string{tmpSrc})
	if err == nil {
		t.Error("expected error when filtering for nonexistent skill")
	}
	if !strings.Contains(err.Error(), "no matching skills") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunInstallWithAgentFilter(t *testing.T) {
	homeDir, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	installYes = true
	installListOnly = false
	installSkills = ""
	installAgents = "claude-code"
	installAll = false

	// Create agent config dir so agent is "detected"
	_ = os.MkdirAll(filepath.Join(homeDir, ".claude"), 0o755)

	// Create a local skill source
	tmpSrc, err := os.MkdirTemp("", "scribe-install-agent-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	skillContent := "---\nname: agent-filter-skill\ndescription: Skill with agent filter\n---\n\n# Test\n"
	_ = os.WriteFile(filepath.Join(tmpSrc, "SKILL.md"), []byte(skillContent), 0o644)

	captureStdout(t, func() {
		err = runInstall(installCmd, []string{tmpSrc})
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	exists, _ := scribe.SkillExists("agent-filter-skill")
	if !exists {
		t.Error("expected skill to be installed")
	}
}

// ===========================================================================
// Tests for runUninstall with installed skill (uninstall.go)
// ===========================================================================

func TestRunUninstallExistingSkill(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	uninstallAll = false

	_ = scribe.EnsureDefaultWorkspace()
	installFakeSkill(t, "to-remove", "Skill to remove", "github", "owner/repo")

	// Add to default workspace
	_ = scribe.AddSkillToWorkspace("to-remove", "default")

	output := captureStdout(t, func() {
		err := runUninstall(uninstallCmd, []string{"to-remove"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "removed successfully") {
		t.Errorf("expected success message, got: %s", output)
	}

	exists, _ := scribe.SkillExists("to-remove")
	if exists {
		t.Error("skill should have been removed")
	}
}

func TestRunUninstallQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	uninstallAll = false

	installFakeSkill(t, "quiet-remove", "Quiet remove test", "github", "owner/repo")

	output := captureStdout(t, func() {
		err := runUninstall(uninstallCmd, []string{"quiet-remove"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "removed successfully") {
		t.Error("quiet mode should suppress success message")
	}
}

// ===========================================================================
// Tests for runUninstallAll (uninstall.go:75)
// ===========================================================================

func TestRunUninstallAllNoSkills(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false

	output := captureStdout(t, func() {
		err := runUninstallAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "No skills installed") {
		t.Errorf("expected 'No skills installed', got: %s", output)
	}
}

func TestRunUninstallAllNoSkillsQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	output := captureStdout(t, func() {
		err := runUninstallAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "No skills installed") {
		t.Error("quiet mode should suppress message")
	}
}

func TestRunUninstallAllWithSkills(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false

	_ = scribe.EnsureDefaultWorkspace()
	installFakeSkill(t, "skill-one", "First skill", "github", "owner/repo1")
	installFakeSkill(t, "skill-two", "Second skill", "github", "owner/repo2")
	_ = scribe.AddSkillToWorkspace("skill-one", "default")
	_ = scribe.AddSkillToWorkspace("skill-two", "default")

	output := captureStdout(t, func() {
		err := runUninstallAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Removing 2 skill(s)") {
		t.Errorf("expected 'Removing 2 skill(s)', got: %s", output)
	}
	if !strings.Contains(output, "All skills removed") {
		t.Errorf("expected 'All skills removed', got: %s", output)
	}

	// Verify all skills removed
	skills, _ := scribe.ListInstalledSkills()
	if len(skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(skills))
	}
}

func TestRunUninstallAllQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	installFakeSkill(t, "quiet-all-skill", "Quiet skill", "github", "owner/repo")

	output := captureStdout(t, func() {
		err := runUninstallAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "Removing") {
		t.Error("quiet mode should suppress messages")
	}
}

func TestRunUninstallViaAllFlag(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	uninstallAll = true

	installFakeSkill(t, "flag-all-skill", "Flag all skill", "github", "owner/repo")

	err := runUninstall(uninstallCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	exists, _ := scribe.SkillExists("flag-all-skill")
	if exists {
		t.Error("skill should have been removed via --all flag")
	}
}

// ===========================================================================
// Tests for runUpdate (update.go:37) and updateSkill (update.go:102)
// ===========================================================================

func TestRunUpdateNoSkillsInstalled(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	updateForce = false

	output := captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "No skills installed") {
		t.Errorf("expected 'No skills installed', got: %s", output)
	}
}

func TestRunUpdateNoSkillsInstalledQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	updateForce = false

	output := captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "No skills installed") {
		t.Error("quiet mode should suppress message")
	}
}

func TestRunUpdateAllUpToDate(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	updateForce = false

	// Install a local skill (can't be checked for updates)
	installFakeSkill(t, "local-update", "Local update test", "local", "/some/path")

	output := captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	// Local skills have errors in checkSkill, so nothing to update
	if !strings.Contains(output, "All skills are up-to-date") {
		t.Errorf("expected 'All skills are up-to-date', got: %s", output)
	}
}

func TestRunUpdateForceWithLocalSkills(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	updateForce = true

	// Local skill - updateSkill will fail with "local source, cannot update"
	installFakeSkill(t, "force-local", "Force local", "local", "/some/path")

	captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	// Should not panic, just have failures in output
}

func TestRunUpdateSingleSkillNonexistent(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	updateForce = false

	captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{"nonexistent-skill"})
		if err != nil {
			t.Errorf("unexpected error from runUpdate: %v", err)
		}
	})
}

func TestUpdateSkillNonexistent(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	err := updateSkill("nonexistent", false)
	if err == nil {
		t.Error("expected error for nonexistent skill")
	}
	if !strings.Contains(err.Error(), "skill not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateSkillNoMeta(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	// Install a skill without metadata
	scrollsDir, _ := scribe.GetScrollsDir()
	skillDir := filepath.Join(scrollsDir, "no-meta-update")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: no-meta-update\ndescription: No meta\n---\n\nContent\n"), 0o644)

	err := updateSkill("no-meta-update", false)
	if err == nil {
		t.Error("expected error for skill with no metadata")
	}
	if !strings.Contains(err.Error(), "no metadata") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateSkillLocalSource(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	installFakeSkill(t, "local-no-update", "Local no update", "local", "/some/path")

	err := updateSkill("local-no-update", false)
	if err == nil {
		t.Error("expected error for local source skill")
	}
	if !strings.Contains(err.Error(), "local source") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ===========================================================================
// Tests for runWorkspaceUse (workspace.go:155)
// ===========================================================================

func TestRunWorkspaceUseAlreadyActive(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false

	_ = scribe.EnsureDefaultWorkspace()

	output := captureStdout(t, func() {
		err := runWorkspaceUse(workspaceUseCmd, []string{"default"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Already using workspace") {
		t.Errorf("expected 'Already using workspace' message, got: %s", output)
	}
}

func TestRunWorkspaceUseAlreadyActiveQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	_ = scribe.EnsureDefaultWorkspace()

	output := captureStdout(t, func() {
		err := runWorkspaceUse(workspaceUseCmd, []string{"default"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "Already using") {
		t.Error("quiet mode should suppress message")
	}
}

func TestRunWorkspaceUseSwitchWorkspace(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	workspaceDescription = ""

	_ = scribe.EnsureDefaultWorkspace()

	// Create a new workspace
	err := runWorkspaceCreate(workspaceCreateCmd, []string{"my-workspace"})
	if err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	output := captureStdout(t, func() {
		err = runWorkspaceUse(workspaceUseCmd, []string{"my-workspace"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Switching from") {
		t.Errorf("expected 'Switching from' message, got: %s", output)
	}

	// Verify active workspace changed
	config, _ := scribe.LoadConfig()
	if config.ActiveWorkspace != "my-workspace" {
		t.Errorf("expected active workspace to be 'my-workspace', got: %s", config.ActiveWorkspace)
	}
}

func TestRunWorkspaceUseNonexistent(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	_ = scribe.EnsureDefaultWorkspace()

	err := runWorkspaceUse(workspaceUseCmd, []string{"nonexistent-ws"})
	if err == nil {
		t.Error("expected error for nonexistent workspace")
	}
}

// ===========================================================================
// Tests for runWorkspaceAdd (workspace.go:188)
// ===========================================================================

func TestRunWorkspaceAddExistingSkill(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false

	_ = scribe.EnsureDefaultWorkspace()
	installFakeSkill(t, "addable-skill", "Addable skill", "github", "owner/repo")

	output := captureStdout(t, func() {
		err := runWorkspaceAdd(workspaceAddCmd, []string{"addable-skill"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Added") {
		t.Errorf("expected 'Added' in output, got: %s", output)
	}
}

func TestRunWorkspaceAddQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	_ = scribe.EnsureDefaultWorkspace()
	installFakeSkill(t, "quiet-add-skill", "Quiet add", "github", "owner/repo")

	output := captureStdout(t, func() {
		err := runWorkspaceAdd(workspaceAddCmd, []string{"quiet-add-skill"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "Added") {
		t.Error("quiet mode should suppress output")
	}
}

func TestRunWorkspaceAddNonexistentSkill(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	_ = scribe.EnsureDefaultWorkspace()

	err := runWorkspaceAdd(workspaceAddCmd, []string{"nonexistent-skill"})
	if err == nil {
		t.Error("expected error for nonexistent skill")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ===========================================================================
// Tests for runWorkspaceRemove (workspace.go:217)
// ===========================================================================

func TestRunWorkspaceRemoveSkill(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false

	_ = scribe.EnsureDefaultWorkspace()
	installFakeSkill(t, "removable-skill", "Removable skill", "github", "owner/repo")
	_ = scribe.AddSkillToWorkspace("removable-skill", "default")

	output := captureStdout(t, func() {
		err := runWorkspaceRemove(workspaceRemoveCmd, []string{"removable-skill"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Removed") {
		t.Errorf("expected 'Removed' in output, got: %s", output)
	}
}

func TestRunWorkspaceRemoveQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	_ = scribe.EnsureDefaultWorkspace()
	installFakeSkill(t, "quiet-rm-skill", "Quiet remove", "github", "owner/repo")
	_ = scribe.AddSkillToWorkspace("quiet-rm-skill", "default")

	output := captureStdout(t, func() {
		err := runWorkspaceRemove(workspaceRemoveCmd, []string{"quiet-rm-skill"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "Removed") {
		t.Error("quiet mode should suppress output")
	}
}

// ===========================================================================
// Tests for runList (list.go) - additional coverage
// ===========================================================================

func TestRunListWithInstalledSkills(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = false
	namesOnly = false

	installFakeSkill(t, "list-skill-a", "First list skill", "github", "owner/repo-a")
	installFakeSkill(t, "list-skill-b", "Second list skill", "local", "/local/path")

	output := captureStdout(t, func() {
		err := runList(listCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "list-skill-a") {
		t.Errorf("expected 'list-skill-a' in output, got: %s", output)
	}
	if !strings.Contains(output, "list-skill-b") {
		t.Errorf("expected 'list-skill-b' in output, got: %s", output)
	}
	if !strings.Contains(output, "2 skill(s) installed") {
		t.Errorf("expected '2 skill(s) installed' in output, got: %s", output)
	}
}

func TestRunListJSON(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	jsonOutput = true
	namesOnly = false

	installFakeSkill(t, "json-list-skill", "JSON list skill", "github", "owner/repo")

	output := captureStdout(t, func() {
		err := runList(listCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	var parsed struct {
		Skills []skillJSON `json:"skills"`
		Count  int         `json:"count"`
	}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
	}
	if parsed.Count != 1 {
		t.Errorf("expected count 1, got %d", parsed.Count)
	}
	if parsed.Skills[0].Name != "json-list-skill" {
		t.Errorf("expected 'json-list-skill', got %s", parsed.Skills[0].Name)
	}
}

func TestRunListNamesOnly(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	jsonOutput = false
	namesOnly = true

	installFakeSkill(t, "name-only-a", "Name A", "github", "owner/repo")
	installFakeSkill(t, "name-only-b", "Name B", "github", "owner/repo")

	output := captureStdout(t, func() {
		err := runList(listCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %v", len(lines), lines)
	}
}

// ===========================================================================
// Additional edge case tests for parseSource (install.go)
// ===========================================================================

func TestParseSourceTildePath(t *testing.T) {
	source, err := parseSource("~/my-skills")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if source.Type != "local" {
		t.Errorf("expected type 'local', got %s", source.Type)
	}
}

func TestParseSourceGitHubWithSubpath(t *testing.T) {
	source, err := parseSource("owner/repo/skills/react")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if source.Type != "github" {
		t.Errorf("expected type 'github', got %s", source.Type)
	}
	if source.Owner != "owner" {
		t.Errorf("expected owner 'owner', got %s", source.Owner)
	}
	if source.Repo != "repo" {
		t.Errorf("expected repo 'repo', got %s", source.Repo)
	}
	if source.Subpath != "skills/react" {
		t.Errorf("expected subpath 'skills/react', got %s", source.Subpath)
	}
}

func TestParseSourceGitHubShorthandWithRefAndSubpath(t *testing.T) {
	source, err := parseSource("owner/repo/skills/react#v2.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if source.Ref != "v2.0" {
		t.Errorf("expected ref 'v2.0', got %s", source.Ref)
	}
	if source.Subpath != "skills/react" {
		t.Errorf("expected subpath 'skills/react', got %s", source.Subpath)
	}
}

func TestParseSourceGitLabURL(t *testing.T) {
	source, err := parseSource("https://gitlab.com/myowner/myrepo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if source.Type != "gitlab" {
		t.Errorf("expected type 'gitlab', got %s", source.Type)
	}
	if source.Owner != "myowner" {
		t.Errorf("expected owner 'myowner', got %s", source.Owner)
	}
}

func TestParseSourceHTTPURL(t *testing.T) {
	source, err := parseSource("http://example.com/skills")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if source.Type != "well-known" {
		t.Errorf("expected type 'well-known', got %s", source.Type)
	}
}

// ===========================================================================
// Additional edge case tests for formatSkillSource (list.go)
// ===========================================================================

func TestFormatSkillSourceURL(t *testing.T) {
	info := scribe.SkillInfo{
		Source:     "https://example.com/skills",
		SourceType: "url",
	}
	result := formatSkillSource(info)
	if result != "https://example.com/skills" {
		t.Errorf("expected URL directly, got %s", result)
	}
}

func TestFormatSkillSourceUnknownType(t *testing.T) {
	info := scribe.SkillInfo{
		Source:     "custom-source",
		SourceType: "custom",
	}
	result := formatSkillSource(info)
	if result != "custom-source" {
		t.Errorf("expected raw source for unknown type, got %s", result)
	}
}

// ===========================================================================
// Additional edge case tests for copySkillDir and copyFile (update.go)
// ===========================================================================

func TestCopySkillDirNonexistentSource(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copydir-noexist-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	err = copySkillDir(filepath.Join(tmpDir, "nonexistent"), filepath.Join(tmpDir, "dest"))
	if err == nil {
		t.Error("expected error copying nonexistent source directory")
	}
}

func TestCopySkillDirEmptySource(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copydir-empty-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcDir := filepath.Join(tmpDir, "empty-src")
	dstDir := filepath.Join(tmpDir, "empty-dst")
	_ = os.MkdirAll(srcDir, 0o755)

	err = copySkillDir(srcDir, dstDir)
	if err != nil {
		t.Fatalf("unexpected error copying empty dir: %v", err)
	}

	// Verify destination exists and is empty
	entries, err := os.ReadDir(dstDir)
	if err != nil {
		t.Fatalf("failed to read dest dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty dest dir, got %d entries", len(entries))
	}
}

func TestCopyFileDifferentPermissions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copyfile-perms-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcPath := filepath.Join(tmpDir, "executable.sh")
	dstPath := filepath.Join(tmpDir, "copy.sh")

	_ = os.WriteFile(srcPath, []byte("#!/bin/bash\necho hello"), 0o755)

	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	info, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("failed to stat dest: %v", err)
	}
	if info.Mode() != 0o755 {
		t.Errorf("expected mode 0755, got %v", info.Mode())
	}
}

func TestCopyFileToNonexistentDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copyfile-nodir-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcPath := filepath.Join(tmpDir, "source.txt")
	_ = os.WriteFile(srcPath, []byte("content"), 0o644)

	// copyFile in update.go doesn't create parent dirs, so writing to a deeply
	// nested path should fail
	dstPath := filepath.Join(tmpDir, "deep", "nested", "dest.txt")
	err = copyFile(srcPath, dstPath)
	if err == nil {
		t.Error("expected error writing to nonexistent parent directory")
	}
}

func TestCopySkillDirDeepNesting(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copydir-deep-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcDir := filepath.Join(tmpDir, "deep-src")
	dstDir := filepath.Join(tmpDir, "deep-dst")

	// Create deeply nested structure
	deepPath := filepath.Join(srcDir, "a", "b", "c")
	_ = os.MkdirAll(deepPath, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "top.txt"), []byte("top"), 0o644)
	_ = os.WriteFile(filepath.Join(srcDir, "a", "mid.txt"), []byte("mid"), 0o644)
	_ = os.WriteFile(filepath.Join(deepPath, "deep.txt"), []byte("deep"), 0o644)

	if err := copySkillDir(srcDir, dstDir); err != nil {
		t.Fatalf("copySkillDir failed: %v", err)
	}

	// Verify all files were copied
	for _, rel := range []string{"top.txt", "a/mid.txt", "a/b/c/deep.txt"} {
		data, err := os.ReadFile(filepath.Join(dstDir, rel))
		if err != nil {
			t.Errorf("failed to read %s: %v", rel, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("file %s is empty", rel)
		}
	}
}

// ===========================================================================
// Tests for runOnboarding (onboarding.go:51) - partial test
// We can't fully test interactive onboarding, but we can test the
// code paths that don't require stdin.
// ===========================================================================

func TestRunOnboardingNoAgents(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	// In temp HOME, no agents should be detected
	output := captureStdout(t, func() {
		err := runOnboarding()
		if err == nil {
			t.Error("expected error when no agents detected")
		}
		if !strings.Contains(err.Error(), "no coding agents detected") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Welcome to Scribe") {
		t.Errorf("expected welcome message, got: %s", output)
	}
	if !strings.Contains(output, "No coding agents detected") {
		t.Errorf("expected 'No coding agents detected', got: %s", output)
	}
}

// ===========================================================================
// Tests for workspace commands with description
// ===========================================================================

func TestRunWorkspaceCreateWithDescription(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	workspaceDescription = "My test workspace"

	err := runWorkspaceCreate(workspaceCreateCmd, []string{"described-ws"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ws, err := scribe.GetWorkspace("described-ws")
	if err != nil {
		t.Fatalf("failed to get workspace: %v", err)
	}
	if ws.Description != "My test workspace" {
		t.Errorf("expected description 'My test workspace', got: %s", ws.Description)
	}
}

// ===========================================================================
// Tests for runWorkspaceList with active marker and description
// ===========================================================================

func TestRunWorkspaceListWithMultiple(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = false
	workspaceDescription = "Dev workspace"

	_ = scribe.EnsureDefaultWorkspace()
	_ = runWorkspaceCreate(workspaceCreateCmd, []string{"dev"})

	output := captureStdout(t, func() {
		err := runWorkspaceList(workspaceListCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// default should have the active marker (*)
	if !strings.Contains(output, "* default") {
		t.Errorf("expected '* default' (active marker) in output, got: %s", output)
	}
	if !strings.Contains(output, "dev") {
		t.Errorf("expected 'dev' workspace in output, got: %s", output)
	}
	if !strings.Contains(output, "Dev workspace") {
		t.Errorf("expected description in output, got: %s", output)
	}
}

// ===========================================================================
// Tests for runWorkspaceCurrent with description
// ===========================================================================

func TestRunWorkspaceCurrentWithDescription(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = false
	workspaceDescription = "My custom workspace"

	_ = scribe.EnsureDefaultWorkspace()
	_ = runWorkspaceCreate(workspaceCreateCmd, []string{"custom-ws"})
	_ = scribe.SetActiveWorkspace("custom-ws")

	output := captureStdout(t, func() {
		err := runWorkspaceCurrent(workspaceCurrentCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "custom-ws") {
		t.Errorf("expected 'custom-ws' in output, got: %s", output)
	}
	if !strings.Contains(output, "Description:") {
		t.Errorf("expected 'Description:' in output, got: %s", output)
	}
}

func TestRunWorkspaceCurrentWithSkills(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = false

	_ = scribe.EnsureDefaultWorkspace()
	installFakeSkill(t, "ws-skill-1", "WS Skill", "github", "owner/repo")
	_ = scribe.AddSkillToWorkspace("ws-skill-1", "default")

	output := captureStdout(t, func() {
		err := runWorkspaceCurrent(workspaceCurrentCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "ws-skill-1") {
		t.Errorf("expected 'ws-skill-1' in skill listing, got: %s", output)
	}
	if !strings.Contains(output, "Skills (") {
		t.Errorf("expected 'Skills (' count in output, got: %s", output)
	}
}

// ===========================================================================
// Tests for listSkillsTable with agents (list.go)
// ===========================================================================

func TestListSkillsTableNoAgentsColumn(t *testing.T) {
	saveAndRestoreFlags(t)
	quiet = false

	skills := []scribe.SkillInfo{
		{
			Name:        "no-agents-skill",
			Description: "Skill with no agents",
			Source:      "owner/repo",
			SourceType:  "github",
			InstalledAt: "2024-06-15",
			Agents:      []string{},
		},
	}

	output := captureStdout(t, func() {
		_ = listSkillsTable(skills)
	})

	// Empty agents should display as "-"
	if !strings.Contains(output, "-") {
		t.Errorf("expected '-' for empty agents column, got: %s", output)
	}
}

func TestListSkillsTableQuiet(t *testing.T) {
	saveAndRestoreFlags(t)
	quiet = true

	skills := []scribe.SkillInfo{
		{
			Name:        "quiet-table-skill",
			Description: "Quiet table test",
			Source:      "owner/repo",
			SourceType:  "github",
			InstalledAt: "2024-06-15",
			Agents:      []string{"claude-code"},
		},
	}

	output := captureStdout(t, func() {
		_ = listSkillsTable(skills)
	})

	// Should still have table but not the summary count
	if strings.Contains(output, "skill(s) installed") {
		t.Error("quiet mode should suppress summary")
	}
	if !strings.Contains(output, "quiet-table-skill") {
		t.Error("should still show skill data in quiet mode")
	}
}

// ===========================================================================
// Additional tests for formatSourceInfo edge cases
// ===========================================================================

func TestFormatSourceInfoEmptyFields(t *testing.T) {
	source := &scribe.SourceInfo{
		Type:  "github",
		Owner: "",
		Repo:  "",
	}
	result := formatSourceInfo(source)
	if result != "github:/" {
		t.Errorf("expected 'github:/', got %q", result)
	}
}

// ===========================================================================
// Tests for reconstructSource edge cases (check.go)
// ===========================================================================

func TestReconstructSourceUnknownType(t *testing.T) {
	meta := &scribe.SkillMeta{
		Source:     "something",
		SourceType: "unknown",
	}
	result := reconstructSource(meta)
	if result.Type != "unknown" {
		t.Errorf("expected type 'unknown', got %s", result.Type)
	}
}

func TestReconstructSourceGitLabWithURL(t *testing.T) {
	meta := &scribe.SkillMeta{
		Source:     "owner/repo",
		SourceType: "gitlab",
		SourceURL:  "https://gitlab.com/owner/repo",
	}
	result := reconstructSource(meta)
	if result.URL != "https://gitlab.com/owner/repo" {
		t.Errorf("expected existing URL to be preserved, got %s", result.URL)
	}
}

func TestReconstructSourceURLType(t *testing.T) {
	meta := &scribe.SkillMeta{
		Source:     "https://example.com/skills",
		SourceType: "url",
	}
	result := reconstructSource(meta)
	if result.URL != "https://example.com/skills" {
		t.Errorf("expected URL to be set from source, got %s", result.URL)
	}
}

func TestReconstructSourceZipWithURL(t *testing.T) {
	meta := &scribe.SkillMeta{
		Source:     "https://example.com/skills.zip",
		SourceType: "zip",
		SourceURL:  "https://example.com/skills.zip",
	}
	result := reconstructSource(meta)
	// URL should be preserved from SourceURL
	if result.URL != "https://example.com/skills.zip" {
		t.Errorf("expected URL from SourceURL, got %s", result.URL)
	}
}

// ===========================================================================
// Tests for checkOutputTable with errors (check.go)
// ===========================================================================

func TestCheckOutputTableWithErrors(t *testing.T) {
	saveAndRestoreFlags(t)
	quiet = false

	results := []CheckResult{
		{Name: "err-skill-1", Error: "fetch failed"},
		{Name: "err-skill-2", Error: "timeout"},
	}

	output := captureStdout(t, func() {
		_ = checkOutputTable(results)
	})

	if !strings.Contains(output, "2 error(s)") {
		t.Errorf("expected '2 error(s)' in summary, got: %s", output)
	}
}

// ===========================================================================
// Tests for runUpdate with force and single skill
// ===========================================================================

func TestRunUpdateSingleLocalSkill(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	updateForce = true

	installFakeSkill(t, "single-update", "Single update", "local", "/local/path")

	captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{"single-update"})
		// Should not return an error even though individual update fails
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestRunUpdateForceAllLocalSkills(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	updateForce = true

	installFakeSkill(t, "force-update-1", "Force 1", "local", "/path1")
	installFakeSkill(t, "force-update-2", "Force 2", "local", "/path2")

	output := captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Updating 2 skill(s)") {
		t.Errorf("expected 'Updating 2 skill(s)', got: %s", output)
	}
	// Both should fail since they are local sources
	if !strings.Contains(output, "0/2 skill(s) updated") {
		t.Errorf("expected '0/2 skill(s) updated', got: %s", output)
	}
}

// ===========================================================================
// Tests for runInstall with local source and detected agents
// ===========================================================================

func TestRunInstallLocalWithMultipleSkills(t *testing.T) {
	homeDir, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	installYes = true
	installListOnly = false
	installSkills = ""
	installAgents = ""
	installAll = false

	// Create agent config dir
	_ = os.MkdirAll(filepath.Join(homeDir, ".claude"), 0o755)

	// Create source with multiple skills
	tmpSrc, err := os.MkdirTemp("", "scribe-multi-install-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	// Create two skills in subdirectories
	skill1Dir := filepath.Join(tmpSrc, "skills", "skill-one")
	skill2Dir := filepath.Join(tmpSrc, "skills", "skill-two")
	_ = os.MkdirAll(skill1Dir, 0o755)
	_ = os.MkdirAll(skill2Dir, 0o755)
	_ = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("---\nname: skill-one\ndescription: First skill\n---\n\nContent 1\n"), 0o644)
	_ = os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("---\nname: skill-two\ndescription: Second skill\n---\n\nContent 2\n"), 0o644)

	output := captureStdout(t, func() {
		err = runInstall(installCmd, []string{tmpSrc})
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "skill-one") {
		t.Errorf("expected 'skill-one' in output, got: %s", output)
	}
	if !strings.Contains(output, "skill-two") {
		t.Errorf("expected 'skill-two' in output, got: %s", output)
	}
}

func TestRunInstallFilterBySpecificSkill(t *testing.T) {
	homeDir, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	installYes = true
	installListOnly = false
	installSkills = "pick-me"
	installAgents = ""
	installAll = false

	// Create agent config dir
	_ = os.MkdirAll(filepath.Join(homeDir, ".claude"), 0o755)

	// Create source with two skills
	tmpSrc, err := os.MkdirTemp("", "scribe-filter-install-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	skill1Dir := filepath.Join(tmpSrc, "skills", "pick-me")
	skill2Dir := filepath.Join(tmpSrc, "skills", "skip-me")
	_ = os.MkdirAll(skill1Dir, 0o755)
	_ = os.MkdirAll(skill2Dir, 0o755)
	_ = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("---\nname: pick-me\ndescription: Pick this one\n---\n\nPicked\n"), 0o644)
	_ = os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("---\nname: skip-me\ndescription: Skip this one\n---\n\nSkipped\n"), 0o644)

	captureStdout(t, func() {
		err = runInstall(installCmd, []string{tmpSrc})
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	exists, _ := scribe.SkillExists("pick-me")
	if !exists {
		t.Error("expected 'pick-me' to be installed")
	}
	exists, _ = scribe.SkillExists("skip-me")
	if exists {
		t.Error("expected 'skip-me' to NOT be installed")
	}
}

// ===========================================================================
// Test for runInstall showing already-exists error
// ===========================================================================

func TestRunInstallDuplicateSkill(t *testing.T) {
	homeDir, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	installYes = true
	installListOnly = false
	installSkills = ""
	installAgents = ""
	installAll = false

	// Create agent config dir
	_ = os.MkdirAll(filepath.Join(homeDir, ".claude"), 0o755)

	// Pre-install the skill
	installFakeSkill(t, "already-here", "Already installed", "github", "owner/repo")

	// Create source with same skill name
	tmpSrc, err := os.MkdirTemp("", "scribe-dup-install-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	_ = os.WriteFile(filepath.Join(tmpSrc, "SKILL.md"), []byte("---\nname: already-here\ndescription: Duplicate\n---\n\nDup\n"), 0o644)

	// Capture stderr too since "Failed to install" goes to stderr
	oldStderr := os.Stderr
	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	output := captureStdout(t, func() {
		err = runInstall(installCmd, []string{tmpSrc})
	})

	_ = stderrW.Close()
	os.Stderr = oldStderr
	var stderrBuf bytes.Buffer
	_, _ = stderrBuf.ReadFrom(stderrR)
	stderrOutput := stderrBuf.String()

	// Should not fail overall, but the skill should fail to install
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "Failed to install" goes to stderr, "Installed 0/1" goes to stdout
	if !strings.Contains(output, "0/1") {
		t.Errorf("expected '0/1' installed count, got stdout: %s", output)
	}
	if !strings.Contains(stderrOutput, "already exists") {
		t.Errorf("expected 'already exists' in stderr, got: %s", stderrOutput)
	}
}

// ===========================================================================
// Test for infoTable with no metadata (info.go)
// ===========================================================================

func TestInfoTableNoMeta(t *testing.T) {
	skill := &scribe.Skill{
		Name:        "no-meta-info",
		Description: "No metadata info test",
		Meta:        nil,
	}

	output := captureStdout(t, func() {
		_ = infoTable(skill, []string{})
	})

	if !strings.Contains(output, "no-meta-info") {
		t.Errorf("expected skill name in output, got: %s", output)
	}
	// Should not have Source: line since meta is nil
	if strings.Contains(output, "Source:") {
		t.Error("should not have Source: line when meta is nil")
	}
	if !strings.Contains(output, "(none)") {
		t.Error("expected '(none)' for agents")
	}
}

// ===========================================================================
// Tests for runInstall verbose paths (install.go)
// ===========================================================================

func TestRunInstallVerboseOutput(t *testing.T) {
	homeDir, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	installYes = true
	installListOnly = false
	installSkills = ""
	installAgents = ""
	installAll = false

	// Create agent config dir
	_ = os.MkdirAll(filepath.Join(homeDir, ".claude"), 0o755)

	// Create a local skill source
	tmpSrc, err := os.MkdirTemp("", "scribe-verbose-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	_ = os.WriteFile(filepath.Join(tmpSrc, "SKILL.md"), []byte("---\nname: verbose-skill\ndescription: Verbose test skill\n---\n\n# Verbose\n"), 0o644)

	output := captureStdout(t, func() {
		err = runInstall(installCmd, []string{tmpSrc})
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verbose output should contain "Fetching skills from"
	if !strings.Contains(output, "Fetching skills from") {
		t.Errorf("expected 'Fetching skills from', got: %s", output)
	}
	// Should list found skills
	if !strings.Contains(output, "Found 1 skill(s) to install") {
		t.Errorf("expected 'Found 1 skill(s) to install', got: %s", output)
	}
	// Should list detected agents
	if !strings.Contains(output, "Detected") {
		t.Errorf("expected 'Detected' in output, got: %s", output)
	}
	// Should show install progress
	if !strings.Contains(output, "Installing verbose-skill") {
		t.Errorf("expected 'Installing verbose-skill', got: %s", output)
	}
	// Should show success count
	if !strings.Contains(output, "Installed 1/1 skill(s)") {
		t.Errorf("expected 'Installed 1/1 skill(s)', got: %s", output)
	}
}

// ===========================================================================
// Tests for runUpdate verbose paths (update.go)
// ===========================================================================

func TestRunUpdateVerboseChecking(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	updateForce = false

	// Install two local skills (local skills have errors in check, so 0 to update)
	installFakeSkill(t, "verbose-check-1", "Verbose check 1", "local", "/local/1")
	installFakeSkill(t, "verbose-check-2", "Verbose check 2", "local", "/local/2")

	output := captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Checking 2 skill(s) for updates") {
		t.Errorf("expected checking message, got: %s", output)
	}
	if !strings.Contains(output, "All skills are up-to-date") {
		t.Errorf("expected all up-to-date message, got: %s", output)
	}
}

func TestRunUpdateVerboseAllUpToDateQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	updateForce = false

	installFakeSkill(t, "quiet-check", "Quiet check", "local", "/local")

	output := captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "Checking") {
		t.Error("quiet mode should suppress checking message")
	}
	if strings.Contains(output, "All skills are up-to-date") {
		t.Error("quiet mode should suppress up-to-date message")
	}
}

func TestRunUpdateForceVerbose(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	updateForce = true

	installFakeSkill(t, "force-v-1", "Force verbose 1", "local", "/local")

	output := captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	// Force mode should show "Updating N skill(s)"
	if !strings.Contains(output, "Updating 1 skill(s)") {
		t.Errorf("expected 'Updating 1 skill(s)', got: %s", output)
	}
}

// ===========================================================================
// Tests for runUninstall verbose path
// ===========================================================================

func TestRunUninstallVerbosePath(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	uninstallAll = false

	_ = scribe.EnsureDefaultWorkspace()
	installFakeSkill(t, "verbose-rm", "Verbose remove", "github", "owner/repo")

	output := captureStdout(t, func() {
		err := runUninstall(uninstallCmd, []string{"verbose-rm"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Removing skill 'verbose-rm'") {
		t.Errorf("expected 'Removing skill' message, got: %s", output)
	}
	if !strings.Contains(output, "removed successfully") {
		t.Errorf("expected 'removed successfully', got: %s", output)
	}
}

// ===========================================================================
// Tests for runUninstallAll verbose with individual skill messages
// ===========================================================================

func TestRunUninstallAllVerboseIndividual(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false

	_ = scribe.EnsureDefaultWorkspace()
	installFakeSkill(t, "ua-skill-1", "Uninstall all 1", "github", "o/r1")
	installFakeSkill(t, "ua-skill-2", "Uninstall all 2", "github", "o/r2")

	output := captureStdout(t, func() {
		err := runUninstallAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Removed ua-skill-1") {
		t.Errorf("expected individual removal message for skill-1, got: %s", output)
	}
	if !strings.Contains(output, "Removed ua-skill-2") {
		t.Errorf("expected individual removal message for skill-2, got: %s", output)
	}
}

// ===========================================================================
// Tests for runWorkspaceUse verbose output
// ===========================================================================

func TestRunWorkspaceUseVerboseOutput(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	workspaceDescription = ""

	_ = scribe.EnsureDefaultWorkspace()
	_ = runWorkspaceCreate(workspaceCreateCmd, []string{"target-ws"})

	output := captureStdout(t, func() {
		err := runWorkspaceUse(workspaceUseCmd, []string{"target-ws"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Switching from") {
		t.Errorf("expected 'Switching from', got: %s", output)
	}
	if !strings.Contains(output, "Active workspace: target-ws") {
		t.Errorf("expected 'Active workspace: target-ws', got: %s", output)
	}
}

// ===========================================================================
// Test for runOnboarding with agents detected (requires agent config dirs)
// ===========================================================================

func TestRunOnboardingWithAgentDetectedNoExistingSkills(t *testing.T) {
	homeDir, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	// Create agent config dir so agents are detected
	_ = os.MkdirAll(filepath.Join(homeDir, ".claude"), 0o755)

	// Redirect stdin to provide "n" input for the onboarding prompt
	// Since no existing skills, onboarding should proceed to install demo skill
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	// Write empty input (just close it to auto-skip any prompts)
	_ = w.Close()

	output := captureStdout(t, func() {
		err := runOnboarding()
		// Restore stdin before checking
		os.Stdin = oldStdin
		if err != nil {
			t.Logf("onboarding error (expected for limited test): %v", err)
		}
	})
	os.Stdin = oldStdin

	// Should show agent detection
	if !strings.Contains(output, "Detecting installed coding agents") {
		t.Errorf("expected agent detection message, got: %s", output)
	}
	if !strings.Contains(output, "Found") && !strings.Contains(output, "coding agent") {
		t.Errorf("expected found agents message, got: %s", output)
	}
}

// ===========================================================================
// Test for checkSkill with github source type (will fail at fetch)
// ===========================================================================

func TestCheckSkillGitHubSourceFetchFails(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	// Install a skill with github source type pointing to invalid repo
	installFakeSkill(t, "github-check-skill", "GitHub check test", "github", "nonexistent-owner/nonexistent-repo")

	result := checkSkill("github-check-skill")
	// Should fail at the fetch step
	if result.Error == "" {
		t.Error("expected error when fetching nonexistent github repo")
	}
	if !strings.Contains(result.Error, "failed to fetch") {
		t.Errorf("expected fetch error, got: %s", result.Error)
	}
}

// ===========================================================================
// Test for updateSkill with github source (will fail at fetch)
// ===========================================================================

func TestUpdateSkillGitHubFetchFails(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	installFakeSkill(t, "github-update-skill", "GitHub update test", "github", "nonexistent-owner/nonexistent-repo")

	err := updateSkill("github-update-skill", false)
	if err == nil {
		t.Error("expected error for nonexistent github repo fetch")
	}
	if !strings.Contains(err.Error(), "failed to fetch") {
		t.Errorf("expected 'failed to fetch' error, got: %v", err)
	}
}

func TestUpdateSkillGitHubFetchFailsVerbose(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false

	installFakeSkill(t, "verbose-github-update", "Verbose github update", "github", "nonexistent-owner/nonexistent-repo")

	output := captureStdout(t, func() {
		err := updateSkill("verbose-github-update", false)
		if err == nil {
			t.Error("expected error for nonexistent github repo")
		}
	})

	if !strings.Contains(output, "Updating verbose-github-update") {
		t.Errorf("expected 'Updating' message, got: %s", output)
	}
}

// ===========================================================================
// Test cache commands (cache.go)
// ===========================================================================

func TestCacheClearCommand(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false

	output := captureStdout(t, func() {
		err := cacheClearCmd.RunE(cacheClearCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Cache cleared") {
		t.Errorf("expected 'Cache cleared', got: %s", output)
	}
}

func TestCacheClearCommandQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	output := captureStdout(t, func() {
		err := cacheClearCmd.RunE(cacheClearCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "Cache cleared") {
		t.Error("quiet mode should suppress message")
	}
}

func TestCachePathCommand(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	output := captureStdout(t, func() {
		err := cachePathCmd.RunE(cachePathCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, ".scribe") || !strings.Contains(output, "cache") {
		t.Errorf("expected cache path containing '.scribe' and 'cache', got: %s", output)
	}
}

// ===========================================================================
// Test for runWorkspaceAdd loading config error path
// ===========================================================================

func TestRunWorkspaceRemoveNonexistentSkill(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	_ = scribe.EnsureDefaultWorkspace()

	// Removing a skill that doesn't exist in the workspace should not error
	err := runWorkspaceRemove(workspaceRemoveCmd, []string{"not-in-workspace"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ===========================================================================
// Test for runList sorting behavior
// ===========================================================================

func TestRunListSortsAlphabetically(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	jsonOutput = false
	namesOnly = true

	// Install in reverse alphabetical order
	installFakeSkill(t, "zebra-skill", "Zebra", "github", "o/r")
	installFakeSkill(t, "alpha-skill", "Alpha", "github", "o/r")
	installFakeSkill(t, "middle-skill", "Middle", "github", "o/r")

	output := captureStdout(t, func() {
		err := runList(listCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "alpha-skill" {
		t.Errorf("expected first skill 'alpha-skill', got '%s'", lines[0])
	}
	if lines[1] != "middle-skill" {
		t.Errorf("expected second skill 'middle-skill', got '%s'", lines[1])
	}
	if lines[2] != "zebra-skill" {
		t.Errorf("expected third skill 'zebra-skill', got '%s'", lines[2])
	}
}

// ===========================================================================
// Test for runOnboardingIfNeeded when not complete
// ===========================================================================

func TestRunOnboardingIfNeededNotComplete(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	// onboarding is not completed in fresh HOME
	// runOnboardingIfNeeded should call runOnboarding which will fail
	// because no agents are detected
	output := captureStdout(t, func() {
		err := runOnboardingIfNeeded()
		if err == nil {
			t.Error("expected error from runOnboarding (no agents)")
		}
	})

	if !strings.Contains(output, "Welcome to Scribe") {
		t.Errorf("expected welcome message, got: %s", output)
	}
}

// ===========================================================================
// Test for Execute with help flag (root.go)
// ===========================================================================

func TestExecuteHelpCommand(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	_ = scribe.CompleteOnboarding()

	os.Args = []string{"scribe", "help"}
	output := captureStdout(t, func() {
		code := Execute()
		if code != ExitSuccess {
			t.Errorf("Execute() returned %d, expected %d", code, ExitSuccess)
		}
	})

	if !strings.Contains(output, "Scribe CLI provides") {
		t.Errorf("expected help text, got: %s", output)
	}
}

// ===========================================================================
// Test for runWorkspaceList with description on non-active workspace
// ===========================================================================

func TestRunWorkspaceListDescriptionFormat(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = false

	_ = scribe.EnsureDefaultWorkspace()

	// Create workspace with no description
	workspaceDescription = ""
	_ = runWorkspaceCreate(workspaceCreateCmd, []string{"no-desc-ws"})

	output := captureStdout(t, func() {
		err := runWorkspaceList(workspaceListCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// "no-desc-ws" line should not have " - " since no description
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "no-desc-ws") {
			if strings.Contains(line, " - ") {
				t.Errorf("workspace without description should not have ' - ' separator: %s", line)
			}
			break
		}
	}
}
