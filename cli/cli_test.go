package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/usescrolls/scribe/internal"
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

	expected := []string{"install", "uninstall", "remove", "rm", "list", "ls", "info", "version", "help", "workspace"}

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

// TestFormatSourceEntry tests the formatSourceEntry function (legacy)
func TestFormatSourceEntry(t *testing.T) {
	tests := []struct {
		name     string
		source   scribe.PluginSource
		expected string
	}{
		{
			name: "github source",
			source: scribe.PluginSource{
				Source: "github",
				Repo:   "owner/repo",
			},
			expected: "github:owner/repo",
		},
		{
			name: "git source",
			source: scribe.PluginSource{
				Source: "git",
				URL:    "https://git.example.com/repo",
			},
			expected: "url:https://git.example.com/repo",
		},
		{
			name: "url source",
			source: scribe.PluginSource{
				Source: "url",
				URL:    "https://example.com/repo",
			},
			expected: "url:https://example.com/repo",
		},
		{
			name: "zip source",
			source: scribe.PluginSource{
				Source: "zip",
				URL:    "https://example.com/plugin.zip",
			},
			expected: "zip:https://example.com/plugin.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSourceEntry(tt.source)
			if result != tt.expected {
				t.Errorf("formatSourceEntry() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestListOutput tests the list command output formats (legacy plugin list)
func TestListOutput(t *testing.T) {
	_, cleanup := setupTestServer(t)
	defer cleanup()

	// Add test plugins
	now := time.Now()
	server.SetRegistryEntry("plugin-a", scribe.RegistryEntry{
		Name:    "plugin-a",
		Version: "1.0.0",
		Source: scribe.PluginSource{
			Source: "github",
			Repo:   "user/plugin-a",
		},
		InstalledAt: now,
	})
	server.SetRegistryEntry("plugin-b", scribe.RegistryEntry{
		Name: "plugin-b",
		Source: scribe.PluginSource{
			Source: "url",
			URL:    "https://github.com/scope/plugin-b.git",
		},
		InstalledAt: now,
	})

	t.Run("table output", func(t *testing.T) {
		// Capture stdout
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

		// Verify table headers
		if !strings.Contains(output, "NAME") {
			t.Error("expected NAME header in table output")
		}
		if !strings.Contains(output, "SOURCE") {
			t.Error("expected SOURCE header in table output")
		}
		if !strings.Contains(output, "plugin-a") {
			t.Error("expected plugin-a in table output")
		}
		if !strings.Contains(output, "plugin-b") {
			t.Error("expected plugin-b in table output")
		}
		if !strings.Contains(output, "2 plugin(s) installed") {
			t.Error("expected plugin count in table output")
		}
	})

	t.Run("json output", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		jsonOutput = true
		namesOnly = false
		runList(listCmd, []string{})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Verify valid JSON
		var result struct {
			Plugins []struct {
				Name   string `json:"name"`
				Source struct {
					Source string `json:"source"`
				} `json:"source"`
			} `json:"plugins"`
			Count int `json:"count"`
		}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Errorf("invalid JSON output: %v", err)
		}
		if result.Count != 2 {
			t.Errorf("expected count=2, got %d", result.Count)
		}
	})

	t.Run("names only output", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		jsonOutput = false
		namesOnly = true
		runList(listCmd, []string{})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 2 {
			t.Errorf("expected 2 lines, got %d: %v", len(lines), lines)
		}
		// Should be sorted alphabetically
		if lines[0] != "plugin-a" {
			t.Errorf("expected first line to be 'plugin-a', got %q", lines[0])
		}
		if lines[1] != "plugin-b" {
			t.Errorf("expected second line to be 'plugin-b', got %q", lines[1])
		}
	})

	t.Run("empty list", func(t *testing.T) {
		// Clear registry
		server.UninstallAllPlugins()

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

		if !strings.Contains(output, "No plugins installed") {
			t.Error("expected 'No plugins installed' message")
		}
	})
}

// TestInfoCommand tests the info command
func TestInfoCommand(t *testing.T) {
	_, cleanup := setupTestServer(t)
	defer cleanup()

	// Add a test plugin
	now := time.Now()
	server.SetRegistryEntry("test-plugin", scribe.RegistryEntry{
		Name:        "test-plugin",
		Description: "A test plugin",
		Version:     "1.2.3",
		Category:    "testing",
		Author:      &scribe.Author{Name: "Test Author"},
		Tags:        []string{"test", "example"},
		Source: scribe.PluginSource{
			Source: "github",
			Repo:   "user/test-plugin",
		},
		InstalledAt: now,
	})

	t.Run("existing plugin", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		runInfo(infoCmd, []string{"test-plugin"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		expectedFields := []string{
			"Name:        test-plugin",
			"Source:      github:user/test-plugin",
			"Version:     1.2.3",
			"Category:    testing",
			"Description: A test plugin",
			"Author:      Test Author",
			"Tags:        test, example",
		}

		for _, field := range expectedFields {
			if !strings.Contains(output, field) {
				t.Errorf("expected output to contain %q", field)
			}
		}
	})
}

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
	_, cleanup := setupTestServer(t)
	defer cleanup()

	t.Run("quiet list with no plugins", func(t *testing.T) {
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

		// In quiet mode, "No plugins installed" should not be printed
		if strings.Contains(output, "No plugins installed") {
			t.Error("quiet mode should suppress 'No plugins installed' message")
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

// TestListJSONFormat tests the JSON output structure
func TestListJSONFormat(t *testing.T) {
	_, cleanup := setupTestServer(t)
	defer cleanup()

	// Add a plugin with all fields
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	server.SetRegistryEntry("full-plugin", scribe.RegistryEntry{
		Name:    "full-plugin",
		Version: "3.0.0",
		Source: scribe.PluginSource{
			Source: "github",
			Repo:   "owner/full-plugin",
			Ref:    "main",
		},
		InstalledAt: now,
	})

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	jsonOutput = true
	namesOnly = false
	runList(listCmd, []string{})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	var result struct {
		Plugins []struct {
			Name   string `json:"name"`
			Source struct {
				Source string `json:"source"`
				Repo   string `json:"repo"`
				Ref    string `json:"ref,omitempty"`
			} `json:"source"`
			Version     string `json:"version,omitempty"`
			InstalledAt string `json:"installedAt"`
		} `json:"plugins"`
		Count int `json:"count"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("expected count=1, got %d", result.Count)
	}
	if len(result.Plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(result.Plugins))
	}

	plugin := result.Plugins[0]
	if plugin.Name != "full-plugin" {
		t.Errorf("expected name='full-plugin', got %q", plugin.Name)
	}
	if plugin.Version != "3.0.0" {
		t.Errorf("expected version='3.0.0', got %q", plugin.Version)
	}
	if plugin.Source.Source != "github" {
		t.Errorf("expected source.source='github', got %q", plugin.Source.Source)
	}
	if plugin.Source.Repo != "owner/full-plugin" {
		t.Errorf("expected source.repo='owner/full-plugin', got %q", plugin.Source.Repo)
	}
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
