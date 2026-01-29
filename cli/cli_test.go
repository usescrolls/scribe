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

	expected := []string{"install", "uninstall", "remove", "rm", "list", "ls", "info", "version", "help"}

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

// TestFormatSource tests the formatSource function
func TestFormatSource(t *testing.T) {
	tests := []struct {
		name     string
		source   scribe.PluginSource
		expected string
	}{
		{
			name: "github source",
			source: scribe.PluginSource{
				Source: "github",
				Repo:   "user/repo",
			},
			expected: "github:user/repo",
		},
		{
			name: "github source with ref",
			source: scribe.PluginSource{
				Source: "github",
				Repo:   "user/repo",
				Ref:    "v1.0.0",
			},
			expected: "github:user/repo@v1.0.0",
		},
		{
			name: "npm source",
			source: scribe.PluginSource{
				Source:  "npm",
				Package: "@scope/package",
			},
			expected: "npm:@scope/package",
		},
		{
			name: "url source",
			source: scribe.PluginSource{
				Source: "url",
				URL:    "https://example.com/repo.git",
			},
			expected: "url:https://example.com/repo.git",
		},
		{
			name: "zip source",
			source: scribe.PluginSource{
				Source: "zip",
				URL:    "https://example.com/plugin.zip",
			},
			expected: "zip:https://example.com/plugin.zip",
		},
		{
			name: "unknown source",
			source: scribe.PluginSource{
				Source: "custom",
			},
			expected: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSource(tt.source)
			if result != tt.expected {
				t.Errorf("formatSource() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestFormatSourceEntry tests the formatSourceEntry function
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
			name: "npm source",
			source: scribe.PluginSource{
				Source:  "npm",
				Package: "@test/pkg",
			},
			expected: "npm:@test/pkg",
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

// TestInstallValidation tests install command validation
func TestInstallValidation(t *testing.T) {
	_, cleanup := setupTestServer(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "no source flag",
			args:        []string{"test-plugin"},
			expectError: true,
			errorMsg:    "exactly one source flag is required",
		},
		{
			name:        "multiple source flags",
			args:        []string{"test-plugin", "--github", "user/repo", "--npm", "pkg"},
			expectError: true,
			errorMsg:    "only one source flag can be specified",
		},
		{
			name:        "valid github source",
			args:        []string{"test-plugin", "--github", "user/repo"},
			expectError: false,
		},
		{
			name:        "valid npm source",
			args:        []string{"npm-plugin", "--npm", "@scope/package"},
			expectError: false,
		},
		{
			name:        "valid url source",
			args:        []string{"url-plugin", "--url", "https://example.com/repo.git"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags for each test
			githubRepo = ""
			npmPackage = ""
			gitURL = ""
			zipURL = ""
			ref = ""
			noEnable = false

			// Parse flags manually
			cmd := installCmd
			cmd.ResetFlags()
			cmd.Flags().StringVar(&githubRepo, "github", "", "")
			cmd.Flags().StringVar(&npmPackage, "npm", "", "")
			cmd.Flags().StringVar(&gitURL, "url", "", "")
			cmd.Flags().StringVar(&zipURL, "zip", "", "")
			cmd.Flags().StringVar(&ref, "ref", "", "")
			cmd.Flags().BoolVar(&noEnable, "no-enable", false, "")

			err := cmd.ParseFlags(tt.args[1:])
			if err != nil {
				if !tt.expectError {
					t.Fatalf("unexpected flag parse error: %v", err)
				}
				return
			}

			// Run the command
			err = runInstall(cmd, tt.args[:1])
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errorMsg)
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestListOutput tests the list command output formats
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
			Source:  "npm",
			Package: "@scope/plugin-b",
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

// TestUninstallCommand tests the uninstall command
func TestUninstallCommand(t *testing.T) {
	tmpDir, cleanup := setupTestServer(t)
	defer cleanup()

	t.Run("uninstall single plugin", func(t *testing.T) {
		// Add a test plugin
		server.SetRegistryEntry("to-uninstall", scribe.RegistryEntry{
			Name: "to-uninstall",
			Source: scribe.PluginSource{
				Source: "github",
				Repo:   "user/repo",
			},
			ResolvedSource: map[string]interface{}{
				"source": "github",
				"repo":   "user/repo",
			},
		})
		server.SaveRegistry()
		server.GenerateMarketplace()

		if server.PluginCount() != 1 {
			t.Fatalf("expected 1 plugin, got %d", server.PluginCount())
		}

		quiet = false
		uninstallAll = false
		err := runUninstallSingle("to-uninstall")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if server.PluginCount() != 0 {
			t.Errorf("expected 0 plugins after uninstall, got %d", server.PluginCount())
		}
	})

	t.Run("uninstall all plugins", func(t *testing.T) {
		// Add multiple plugins
		server.SetRegistryEntry("plugin1", scribe.RegistryEntry{
			Name: "plugin1",
			ResolvedSource: map[string]interface{}{
				"source": "github",
				"repo":   "user/plugin1",
			},
		})
		server.SetRegistryEntry("plugin2", scribe.RegistryEntry{
			Name: "plugin2",
			ResolvedSource: map[string]interface{}{
				"source": "npm",
				"package": "@scope/plugin2",
			},
		})
		server.SaveRegistry()

		if server.PluginCount() != 2 {
			t.Fatalf("expected 2 plugins, got %d", server.PluginCount())
		}

		quiet = false
		err := runUninstallAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if server.PluginCount() != 0 {
			t.Errorf("expected 0 plugins after uninstall --all, got %d", server.PluginCount())
		}
	})

	t.Run("uninstall with relative source deletes files", func(t *testing.T) {
		// Add a plugin with relative source (like zip-installed)
		pluginDir := filepath.Join(tmpDir, "plugins", "local-plugin")
		os.MkdirAll(pluginDir, 0755)
		os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte("{}"), 0644)

		server.SetRegistryEntry("local-plugin", scribe.RegistryEntry{
			Name:           "local-plugin",
			ResolvedSource: "./plugins/local-plugin",
		})
		server.SaveRegistry()

		// Verify directory exists
		if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
			t.Fatal("plugin directory should exist before uninstall")
		}

		runUninstallSingle("local-plugin")

		// Verify directory was deleted
		if _, err := os.Stat(pluginDir); !os.IsNotExist(err) {
			t.Error("plugin directory should be deleted after uninstall")
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

// TestInstallWithRef tests install command with --ref flag
func TestInstallWithRef(t *testing.T) {
	_, cleanup := setupTestServer(t)
	defer cleanup()

	githubRepo = "user/repo"
	npmPackage = ""
	gitURL = ""
	zipURL = ""
	ref = "v2.0.0"
	noEnable = true

	err := runInstall(installCmd, []string{"versioned-plugin"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry, ok := server.GetRegistryEntry("versioned-plugin")
	if !ok {
		t.Fatal("plugin not found after install")
	}

	if entry.Source.Ref != "v2.0.0" {
		t.Errorf("expected ref 'v2.0.0', got %q", entry.Source.Ref)
	}
}

// TestInstallNoEnable tests install command with --no-enable flag
func TestInstallNoEnable(t *testing.T) {
	tmpDir, cleanup := setupTestServer(t)
	defer cleanup()

	githubRepo = "user/repo"
	npmPackage = ""
	gitURL = ""
	zipURL = ""
	ref = ""
	noEnable = true

	err := runInstall(installCmd, []string{"no-enable-plugin"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read Claude settings to verify plugin is NOT enabled
	settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings: %v", err)
	}

	var settings map[string]interface{}
	json.Unmarshal(data, &settings)

	enabledPlugins, _ := settings["enabledPlugins"].(map[string]interface{})
	pluginID := "no-enable-plugin@" + scribe.MarketplaceName
	if _, exists := enabledPlugins[pluginID]; exists {
		t.Error("plugin should not be enabled when --no-enable flag is used")
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
			Name        string `json:"name"`
			Source      struct {
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
