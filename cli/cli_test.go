package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	scribe "github.com/usescrolls/scribe/internal"
)

// setupTestServer creates a test server with a temporary directory
func setupTestServer(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "scribe-cli-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	scribe.InitLoggerCLI(false)
	server = scribe.NewTestServer(tmpDir, filepath.Join(tmpDir, ".claude"))
	if err := server.Initialize(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to initialize server: %v", err)
	}

	// Create Claude settings directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte("{}"), 0644)

	return tmpDir, func() {
		os.RemoveAll(tmpDir)
	}
}

// TestCLICommands tests that CLICommands returns the expected commands
func TestCLICommands(t *testing.T) {
	commands := CLICommands()

	expected := []string{"install", "uninstall", "remove", "rm", "list", "ls", "info", "version", "help", "workspace", "check", "update"}

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
			name:        "local path relative",
			input:       "./my-skills",
			expectType:  "local",
		},
		{
			name:        "local path absolute",
			input:       "/absolute/path",
			expectType:  "local",
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
	defer os.RemoveAll(tmpDir)

	// Override home directory for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	scribe.InitLoggerCLI(false)

	t.Run("empty list", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		jsonOutput = false
		namesOnly = false
		quiet = false
		runList(listCmd, []string{})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		if !strings.Contains(output, "No skills installed") {
			t.Errorf("expected 'No skills installed' message, got: %s", output)
		}
	})
}

// Note: TestInfoCommand removed - info command now uses skills system
// which requires actual skill files to be installed. See internal/skills_system_test.go
// for comprehensive skills system tests.

// TestVersionCommand tests the version command
func TestVersionCommand(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	runVersion(versionCmd, []string{})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
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
	defer os.RemoveAll(tmpDir)

	// Override home directory for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	scribe.InitLoggerCLI(false)

	t.Run("quiet list with no skills", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		quiet = true
		jsonOutput = false
		namesOnly = false
		runList(listCmd, []string{})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// In quiet mode, "No skills installed" should not be printed
		if strings.Contains(output, "No skills installed") {
			t.Error("quiet mode should suppress 'No skills installed' message")
		}
	})
}

// TestInitServer tests the initServer helper function
func TestInitServer(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-cli-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	scribe.InitLoggerCLI(false)
	server = scribe.NewTestServer(tmpDir, filepath.Join(tmpDir, ".claude"))

	err = initServer()
	if err != nil {
		t.Fatalf("initServer failed: %v", err)
	}

	// Verify directories were created
	expectedDirs := []string{
		filepath.Join(tmpDir, ".claude-plugin"),
		filepath.Join(tmpDir, "plugins"),
		filepath.Join(tmpDir, "data"),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("expected directory %q to exist", dir)
		}
	}
}

// Note: TestListJSONFormat removed - list command now uses skills system
// which returns SkillInfo instead of RegistryEntry. See internal/skills_system_test.go
// for comprehensive skills system tests.

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
