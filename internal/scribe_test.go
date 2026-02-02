package scribe

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestAuthorUnmarshalJSON tests that Author can be unmarshaled from both string and object formats
func TestAuthorUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "string format",
			input:    `"John Doe"`,
			expected: "John Doe",
		},
		{
			name:     "object format",
			input:    `{"name": "Jane Smith"}`,
			expected: "Jane Smith",
		},
		{
			name:     "empty string",
			input:    `""`,
			expected: "",
		},
		{
			name:     "object with empty name",
			input:    `{"name": ""}`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var author Author
			err := json.Unmarshal([]byte(tt.input), &author)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if author.Name != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, author.Name)
			}
		})
	}
}

func TestAuthorUnmarshalJSON_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid json",
			input: `{invalid}`,
		},
		{
			name:  "number",
			input: `123`,
		},
		{
			name:  "array",
			input: `["name"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var author Author
			err := json.Unmarshal([]byte(tt.input), &author)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

// TestNewServer tests server creation with correct default paths
func TestNewServer(t *testing.T) {
	server := NewServer()

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	homeDir, _ := os.UserHomeDir()
	expectedHubDir := filepath.Join(homeDir, HubDirName)

	if server.HubDir() != expectedHubDir {
		t.Errorf("expected hubDir %q, got %q", expectedHubDir, server.HubDir())
	}
}

// TestServerInitialize tests directory structure creation
func TestServerInitialize(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize logger for tests
	InitLogger(false)

	server := NewTestServer(tmpDir, filepath.Join(tmpDir, ".claude"))

	if err := server.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Check that directories were created
	expectedDirs := []string{
		filepath.Join(tmpDir, MarketplaceDirName),
		filepath.Join(tmpDir, PluginsDirName),
		filepath.Join(tmpDir, DataDirName),
	}

	for _, dir := range expectedDirs {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("directory %q was not created: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%q is not a directory", dir)
		}
	}
}

// TestParseURLScheme tests URL scheme parsing
func TestParseURLScheme(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedAction string
		expectedParams map[string]string
		expectError    bool
	}{
		{
			name:           "install with github source",
			url:            "agenthub://install?name=test-plugin&source=github&repo=user/repo",
			expectedAction: "install",
			expectedParams: map[string]string{
				"name":   "test-plugin",
				"source": "github",
				"repo":   "user/repo",
			},
		},
		{
			name:           "uninstall action",
			url:            "agenthub://uninstall?name=test-plugin",
			expectedAction: "uninstall",
			expectedParams: map[string]string{
				"name": "test-plugin",
			},
		},
		{
			name:           "open action",
			url:            "agenthub://open",
			expectedAction: "open",
			expectedParams: map[string]string{},
		},
		{
			name:           "install with ref",
			url:            "agenthub://install?name=test&source=github&repo=user/repo&ref=v1.0.0",
			expectedAction: "install",
			expectedParams: map[string]string{
				"name":   "test",
				"source": "github",
				"repo":   "user/repo",
				"ref":    "v1.0.0",
			},
		},
		{
			name:        "invalid scheme",
			url:         "http://install?name=test",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, err := ParseURLScheme(tt.url)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if action.Action != tt.expectedAction {
				t.Errorf("expected action %q, got %q", tt.expectedAction, action.Action)
			}

			for key, expected := range tt.expectedParams {
				got := action.Params.Get(key)
				if got != expected {
					t.Errorf("expected param %q=%q, got %q", key, expected, got)
				}
			}
		})
	}
}

// TestResolveSource tests source resolution for different source types
func TestResolveSource(t *testing.T) {
	// Initialize logger for tests
	InitLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	server := NewTestServer(tmpDir, filepath.Join(tmpDir, ".claude"))
	server.Initialize()

	tests := []struct {
		name        string
		pluginName  string
		source      PluginSource
		expectError bool
		validate    func(t *testing.T, result interface{})
	}{
		{
			name:       "github source",
			pluginName: "test-plugin",
			source: PluginSource{
				Source: "github",
				Repo:   "user/repo",
			},
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("expected map result")
				}
				if m["source"] != "github" {
					t.Errorf("expected source=github, got %v", m["source"])
				}
				if m["repo"] != "user/repo" {
					t.Errorf("expected repo=user/repo, got %v", m["repo"])
				}
			},
		},
		{
			name:       "github source with ref",
			pluginName: "test-plugin",
			source: PluginSource{
				Source: "github",
				Repo:   "user/repo",
				Ref:    "v1.0.0",
			},
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("expected map result")
				}
				if m["ref"] != "v1.0.0" {
					t.Errorf("expected ref=v1.0.0, got %v", m["ref"])
				}
			},
		},
		{
			name:       "git url source",
			pluginName: "git-plugin",
			source: PluginSource{
				Source: "git",
				URL:    "https://github.com/user/repo.git",
			},
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("expected map result")
				}
				if m["source"] != "url" {
					t.Errorf("expected source=url, got %v", m["source"])
				}
				if m["url"] != "https://github.com/user/repo.git" {
					t.Errorf("expected url, got %v", m["url"])
				}
			},
		},
		{
			name:       "github missing repo",
			pluginName: "bad-plugin",
			source: PluginSource{
				Source: "github",
			},
			expectError: true,
		},
		{
			name:       "unsupported source type",
			pluginName: "bad-plugin",
			source: PluginSource{
				Source: "unknown",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.ResolveSource(tt.pluginName, tt.source)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

// TestRegistryPersistence tests saving and loading the registry
func TestRegistryPersistence(t *testing.T) {
	InitLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	server := NewTestServer(tmpDir, filepath.Join(tmpDir, ".claude"))
	server.Initialize()

	// Add some entries
	now := time.Now()
	server.SetRegistryEntry("plugin1", RegistryEntry{
		Name:        "plugin1",
		Description: "Test plugin 1",
		Version:     "1.0.0",
		Source: PluginSource{
			Source: "github",
			Repo:   "user/plugin1",
		},
		ResolvedSource: map[string]interface{}{
			"source": "github",
			"repo":   "user/plugin1",
		},
		InstalledAt: now,
	})
	server.SetRegistryEntry("plugin2", RegistryEntry{
		Name:    "plugin2",
		Version: "2.0.0",
		Source: PluginSource{
			Source: "url",
			URL:    "https://github.com/test/plugin2.git",
		},
		ResolvedSource: map[string]interface{}{
			"source": "url",
			"url":    "https://github.com/test/plugin2.git",
		},
		InstalledAt: now,
	})

	// Save
	if err := server.SaveRegistry(); err != nil {
		t.Fatalf("SaveRegistry failed: %v", err)
	}

	// Verify file exists
	registryFile := filepath.Join(tmpDir, DataDirName, RegistryFile)
	if _, err := os.Stat(registryFile); err != nil {
		t.Fatalf("registry file not created: %v", err)
	}

	// Create new server and load
	server2 := NewTestServer(tmpDir, filepath.Join(tmpDir, ".claude"))

	if err := server2.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify loaded data
	if server2.PluginCount() != 2 {
		t.Errorf("expected 2 entries, got %d", server2.PluginCount())
	}

	entry1, ok := server2.GetRegistryEntry("plugin1")
	if !ok {
		t.Fatal("plugin1 not found in loaded registry")
	}
	if entry1.Description != "Test plugin 1" {
		t.Errorf("expected description 'Test plugin 1', got %q", entry1.Description)
	}
	if entry1.Source.Repo != "user/plugin1" {
		t.Errorf("expected repo 'user/plugin1', got %q", entry1.Source.Repo)
	}

	entry2, ok := server2.GetRegistryEntry("plugin2")
	if !ok {
		t.Fatal("plugin2 not found in loaded registry")
	}
	if entry2.Source.URL != "https://github.com/test/plugin2.git" {
		t.Errorf("expected url 'https://github.com/test/plugin2.git', got %q", entry2.Source.URL)
	}
}

// TestGenerateMarketplace tests marketplace.json generation
func TestGenerateMarketplace(t *testing.T) {
	InitLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	server := NewTestServer(tmpDir, filepath.Join(tmpDir, ".claude"))
	server.Initialize()

	// Add a plugin
	server.SetRegistryEntry("test-plugin", RegistryEntry{
		Name:        "test-plugin",
		Description: "A test plugin",
		Version:     "1.0.0",
		Category:    "testing",
		Author:      &Author{Name: "Test Author"},
		Tags:        []string{"test", "example"},
		ResolvedSource: map[string]interface{}{
			"source": "github",
			"repo":   "test/plugin",
		},
	})

	if err := server.GenerateMarketplace(); err != nil {
		t.Fatalf("GenerateMarketplace failed: %v", err)
	}

	// Read and verify marketplace.json
	marketplacePath := filepath.Join(tmpDir, MarketplaceDirName, MarketplaceFile)
	data, err := os.ReadFile(marketplacePath)
	if err != nil {
		t.Fatalf("failed to read marketplace file: %v", err)
	}

	var marketplace map[string]interface{}
	if err := json.Unmarshal(data, &marketplace); err != nil {
		t.Fatalf("failed to parse marketplace json: %v", err)
	}

	if marketplace["name"] != MarketplaceName {
		t.Errorf("expected name %q, got %v", MarketplaceName, marketplace["name"])
	}

	plugins, ok := marketplace["plugins"].([]interface{})
	if !ok {
		t.Fatal("plugins is not an array")
	}

	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}

	plugin := plugins[0].(map[string]interface{})
	if plugin["name"] != "test-plugin" {
		t.Errorf("expected plugin name 'test-plugin', got %v", plugin["name"])
	}
	if plugin["description"] != "A test plugin" {
		t.Errorf("expected description 'A test plugin', got %v", plugin["description"])
	}
	if plugin["category"] != "testing" {
		t.Errorf("expected category 'testing', got %v", plugin["category"])
	}

	tags, ok := plugin["tags"].([]interface{})
	if !ok || len(tags) != 2 {
		t.Errorf("expected 2 tags, got %v", plugin["tags"])
	}
}

// TestPluginCount tests the PluginCount method
func TestPluginCount(t *testing.T) {
	server := NewTestServer("/tmp", "/tmp/.claude")

	if server.PluginCount() != 0 {
		t.Errorf("expected 0, got %d", server.PluginCount())
	}

	server.SetRegistryEntry("plugin1", RegistryEntry{Name: "plugin1"})
	server.SetRegistryEntry("plugin2", RegistryEntry{Name: "plugin2"})

	if server.PluginCount() != 2 {
		t.Errorf("expected 2, got %d", server.PluginCount())
	}
}

// TestUpdateClaudeSettings tests Claude settings management
func TestUpdateClaudeSettings(t *testing.T) {
	InitLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	claudeDir := filepath.Join(tmpDir, ".claude")
	server := NewTestServer(tmpDir, claudeDir)

	// Test enabling a plugin (creates settings file)
	if err := server.UpdateClaudeSettings("test-plugin", true); err != nil {
		t.Fatalf("UpdateClaudeSettings failed: %v", err)
	}

	// Verify settings file
	settingsPath := filepath.Join(claudeDir, "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings file: %v", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("failed to parse settings: %v", err)
	}

	// Check marketplace is registered
	marketplaces, ok := settings["extraKnownMarketplaces"].(map[string]interface{})
	if !ok {
		t.Fatal("extraKnownMarketplaces not found")
	}
	if _, ok := marketplaces[MarketplaceName]; !ok {
		t.Errorf("marketplace %q not registered", MarketplaceName)
	}

	// Check plugin is enabled
	enabledPlugins, ok := settings["enabledPlugins"].(map[string]interface{})
	if !ok {
		t.Fatal("enabledPlugins not found")
	}
	pluginID := "test-plugin@" + MarketplaceName
	if enabledPlugins[pluginID] != true {
		t.Errorf("plugin %q not enabled", pluginID)
	}

	// Test disabling the plugin
	if err := server.UpdateClaudeSettings("test-plugin", false); err != nil {
		t.Fatalf("UpdateClaudeSettings (disable) failed: %v", err)
	}

	data, _ = os.ReadFile(settingsPath)
	json.Unmarshal(data, &settings)
	enabledPlugins = settings["enabledPlugins"].(map[string]interface{})

	if _, exists := enabledPlugins[pluginID]; exists {
		t.Errorf("plugin %q should be removed after disable", pluginID)
	}
}

// TestClaudeCodeDetected tests detection of Claude Code installation
func TestClaudeCodeDetected(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	claudeDir := filepath.Join(tmpDir, ".claude")
	server := NewTestServer(tmpDir, claudeDir)

	// Should return false when settings.json doesn't exist
	if server.ClaudeCodeDetected() {
		t.Error("expected false when claude dir doesn't exist")
	}

	// Create the directory and settings file
	os.MkdirAll(claudeDir, 0755)
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte("{}"), 0644)

	// Should return true now
	if !server.ClaudeCodeDetected() {
		t.Error("expected true when settings.json exists")
	}
}

// TestFullReset tests the complete reset functionality
func TestFullReset(t *testing.T) {
	InitLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	claudeDir := filepath.Join(tmpDir, ".claude")
	server := NewTestServer(tmpDir, claudeDir)
	server.Initialize()

	// Add a plugin
	server.SetRegistryEntry("test-plugin", RegistryEntry{
		Name: "test-plugin",
		ResolvedSource: map[string]interface{}{
			"source": "github",
			"repo":   "test/plugin",
		},
	})
	server.SaveRegistry()
	server.GenerateMarketplace()
	server.UpdateClaudeSettings("test-plugin", true)

	// Verify setup
	if server.PluginCount() != 1 {
		t.Fatal("setup failed: plugin not added")
	}

	// Perform full reset
	if err := server.FullReset(); err != nil {
		t.Fatalf("FullReset failed: %v", err)
	}

	// Verify reset
	if server.PluginCount() != 0 {
		t.Errorf("expected 0 plugins after reset, got %d", server.PluginCount())
	}

	// Verify registry file is deleted
	registryFile := filepath.Join(tmpDir, DataDirName, RegistryFile)
	if _, err := os.Stat(registryFile); !os.IsNotExist(err) {
		t.Error("registry file should be deleted after reset")
	}
}

// TestPluginSourceJSON tests JSON serialization of PluginSource
func TestPluginSourceJSON(t *testing.T) {
	source := PluginSource{
		Source: "github",
		Repo:   "user/repo",
		Ref:    "main",
	}

	data, err := json.Marshal(source)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded PluginSource
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Source != source.Source {
		t.Errorf("source mismatch: %q != %q", decoded.Source, source.Source)
	}
	if decoded.Repo != source.Repo {
		t.Errorf("repo mismatch: %q != %q", decoded.Repo, source.Repo)
	}
	if decoded.Ref != source.Ref {
		t.Errorf("ref mismatch: %q != %q", decoded.Ref, source.Ref)
	}
}

// TestPluginProvidesJSON tests JSON serialization of PluginProvides
func TestPluginProvidesJSON(t *testing.T) {
	provides := PluginProvides{
		Skills:   []string{"skill1", "skill2"},
		Agents:   []string{"agent1"},
		Commands: []string{"cmd1", "cmd2", "cmd3"},
		Hooks:    []string{},
	}

	data, err := json.Marshal(provides)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded PluginProvides
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(decoded.Skills) != 2 {
		t.Errorf("expected 2 skills, got %d", len(decoded.Skills))
	}
	if len(decoded.Commands) != 3 {
		t.Errorf("expected 3 commands, got %d", len(decoded.Commands))
	}
}

// TestRegistryEntryJSON tests JSON serialization of RegistryEntry
func TestRegistryEntryJSON(t *testing.T) {
	entry := RegistryEntry{
		Name:        "test-plugin",
		Description: "A test",
		Version:     "1.0.0",
		Category:    "testing",
		Author:      &Author{Name: "Tester"},
		Tags:        []string{"test"},
		Source: PluginSource{
			Source: "github",
			Repo:   "test/repo",
		},
		ResolvedSource: map[string]interface{}{
			"source": "github",
			"repo":   "test/repo",
		},
		InstalledAt: time.Now(),
		Provides: &PluginProvides{
			Skills: []string{"test-skill"},
		},
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded RegistryEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Name != entry.Name {
		t.Errorf("name mismatch")
	}
	if decoded.Author.Name != "Tester" {
		t.Errorf("author mismatch")
	}
	if decoded.Provides == nil || len(decoded.Provides.Skills) != 1 {
		t.Errorf("provides mismatch")
	}
}

// TestLoadNonExistentRegistry tests that loading a non-existent registry doesn't error
func TestLoadNonExistentRegistry(t *testing.T) {
	InitLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	server := NewTestServer(tmpDir, filepath.Join(tmpDir, ".claude"))

	// Should not error when registry file doesn't exist
	if err := server.Load(); err != nil {
		t.Errorf("Load should not error for non-existent registry: %v", err)
	}

	if server.PluginCount() != 0 {
		t.Errorf("registry should be empty, got %d entries", server.PluginCount())
	}
}

// TestDeletePluginData tests the DeletePluginData method
func TestDeletePluginData(t *testing.T) {
	InitLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	server := NewTestServer(tmpDir, filepath.Join(tmpDir, ".claude"))
	server.Initialize()

	// Add some data
	server.SetRegistryEntry("plugin1", RegistryEntry{Name: "plugin1"})
	server.SaveRegistry()

	// Create a plugin directory
	pluginDir := filepath.Join(tmpDir, PluginsDirName, "test-plugin")
	os.MkdirAll(pluginDir, 0755)
	os.WriteFile(filepath.Join(pluginDir, "test.txt"), []byte("test"), 0644)

	// Delete plugin data
	if err := server.DeletePluginData(); err != nil {
		t.Fatalf("DeletePluginData failed: %v", err)
	}

	// Verify registry is empty
	if server.PluginCount() != 0 {
		t.Errorf("registry should be empty")
	}

	// Verify registry file is deleted
	registryFile := filepath.Join(tmpDir, DataDirName, RegistryFile)
	if _, err := os.Stat(registryFile); !os.IsNotExist(err) {
		t.Error("registry file should be deleted")
	}

	// Verify plugins directory is recreated but empty
	pluginsDir := filepath.Join(tmpDir, PluginsDirName)
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		t.Fatalf("plugins dir should exist: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("plugins dir should be empty, got %d entries", len(entries))
	}
}

// TestClearClaudePluginSettings tests clearing plugin settings from Claude
func TestClearClaudePluginSettings(t *testing.T) {
	InitLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	claudeDir := filepath.Join(tmpDir, ".claude")
	server := NewTestServer(tmpDir, claudeDir)

	// Test clearing when no settings file exists (should not error)
	if err := server.ClearClaudePluginSettings(); err != nil {
		t.Fatalf("ClearClaudePluginSettings should not error when file doesn't exist: %v", err)
	}

	// Setup: create settings with our marketplace and plugins
	os.MkdirAll(claudeDir, 0755)
	settings := map[string]interface{}{
		"extraKnownMarketplaces": map[string]interface{}{
			MarketplaceName: map[string]interface{}{
				"source": map[string]interface{}{
					"source": "directory",
					"path":   tmpDir,
				},
			},
			"other-marketplace": map[string]interface{}{}, // Should remain
		},
		"enabledPlugins": map[string]interface{}{
			"plugin1@" + MarketplaceName:   true,
			"plugin2@" + MarketplaceName:   true,
			"other-plugin@other-market":    true, // Should remain
		},
		"otherSetting": "should remain",
	}
	data, _ := json.Marshal(settings)
	settingsPath := filepath.Join(claudeDir, "settings.json")
	os.WriteFile(settingsPath, data, 0644)

	// Clear settings
	if err := server.ClearClaudePluginSettings(); err != nil {
		t.Fatalf("ClearClaudePluginSettings failed: %v", err)
	}

	// Verify
	data, _ = os.ReadFile(settingsPath)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	// Our marketplace should be removed
	marketplaces := result["extraKnownMarketplaces"].(map[string]interface{})
	if _, exists := marketplaces[MarketplaceName]; exists {
		t.Error("our marketplace should be removed")
	}
	if _, exists := marketplaces["other-marketplace"]; !exists {
		t.Error("other marketplace should remain")
	}

	// Our plugins should be removed
	enabledPlugins := result["enabledPlugins"].(map[string]interface{})
	if _, exists := enabledPlugins["plugin1@"+MarketplaceName]; exists {
		t.Error("plugin1 should be removed")
	}
	if _, exists := enabledPlugins["other-plugin@other-market"]; !exists {
		t.Error("other plugin should remain")
	}

	// Other settings should remain
	if result["otherSetting"] != "should remain" {
		t.Error("other settings should remain unchanged")
	}
}

// TestUninstallAllPlugins tests uninstalling all plugins
func TestUninstallAllPlugins(t *testing.T) {
	InitLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	claudeDir := filepath.Join(tmpDir, ".claude")
	server := NewTestServer(tmpDir, claudeDir)
	server.Initialize()

	// Add plugins with different source types
	server.SetRegistryEntry("plugin1", RegistryEntry{
		Name: "plugin1",
		ResolvedSource: map[string]interface{}{
			"source": "github",
			"repo":   "user/plugin1",
		},
	})
	server.SetRegistryEntry("plugin2", RegistryEntry{
		Name:           "plugin2",
		ResolvedSource: "./plugins/plugin2", // Relative source - should delete directory
	})

	// Create plugin directory for plugin2
	pluginDir := filepath.Join(tmpDir, PluginsDirName, "plugin2")
	os.MkdirAll(pluginDir, 0755)
	os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte("{}"), 0644)

	// Save registry
	server.SaveRegistry()

	// Setup Claude settings
	os.MkdirAll(claudeDir, 0755)
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte("{}"), 0644)
	server.UpdateClaudeSettings("plugin1", true)
	server.UpdateClaudeSettings("plugin2", true)

	// Uninstall all
	if err := server.UninstallAllPlugins(); err != nil {
		t.Fatalf("UninstallAllPlugins failed: %v", err)
	}

	// Verify registry is empty
	if server.PluginCount() != 0 {
		t.Errorf("expected 0 plugins, got %d", server.PluginCount())
	}

	// Verify plugin2 directory was deleted
	if _, err := os.Stat(pluginDir); !os.IsNotExist(err) {
		t.Error("plugin2 directory should be deleted")
	}

	// Verify registry file is empty array
	registryFile := filepath.Join(tmpDir, DataDirName, RegistryFile)
	data, _ := os.ReadFile(registryFile)
	var entries []RegistryEntry
	json.Unmarshal(data, &entries)
	if len(entries) != 0 {
		t.Errorf("registry file should have 0 entries, got %d", len(entries))
	}
}

// TestMigrate tests migration from old plugins.json format
func TestMigrate(t *testing.T) {
	InitLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	claudeDir := filepath.Join(tmpDir, ".claude")
	server := NewTestServer(tmpDir, claudeDir)
	server.Initialize()

	// Test migration when no old file exists (should be no-op)
	if err := server.Migrate(); err != nil {
		t.Fatalf("Migrate should not error when no old file exists: %v", err)
	}

	// Create old format plugins.json
	oldPlugins := []Plugin{
		{
			Name:        "github-plugin",
			Description: "A GitHub plugin",
			Version:     "1.0.0",
			Source: PluginSource{
				Source: "github",
				Repo:   "user/repo",
				Ref:    "main",
			},
		},
		{
			Name: "git-plugin",
			Source: PluginSource{
				Source: "git",
				URL:    "https://example.com/repo.git",
			},
		},
	}
	oldData, _ := json.Marshal(oldPlugins)
	oldPath := filepath.Join(tmpDir, OldPluginsFile)
	os.WriteFile(oldPath, oldData, 0644)

	// Setup Claude settings with URL-based source
	os.MkdirAll(claudeDir, 0755)
	claudeSettings := map[string]interface{}{
		"extraKnownMarketplaces": map[string]interface{}{
			MarketplaceName: map[string]interface{}{
				"source": map[string]interface{}{
					"source": "url",
					"url":    "https://old.url/marketplace.json",
				},
			},
		},
	}
	claudeData, _ := json.Marshal(claudeSettings)
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), claudeData, 0644)

	// Run migration
	if err := server.Migrate(); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	// Verify plugins were migrated
	if server.PluginCount() != 2 {
		t.Errorf("expected 2 plugins, got %d", server.PluginCount())
	}

	// Verify GitHub plugin
	entry, ok := server.GetRegistryEntry("github-plugin")
	if !ok {
		t.Fatal("github-plugin not found")
	}
	if entry.Description != "A GitHub plugin" {
		t.Errorf("expected description 'A GitHub plugin', got %q", entry.Description)
	}
	resolved := entry.ResolvedSource.(map[string]interface{})
	if resolved["ref"] != "main" {
		t.Errorf("expected ref 'main', got %v", resolved["ref"])
	}

	// Verify old file was deleted
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error("old plugins.json should be deleted after migration")
	}

	// Verify new registry file exists
	registryFile := filepath.Join(tmpDir, DataDirName, RegistryFile)
	if _, err := os.Stat(registryFile); err != nil {
		t.Error("new registry file should exist")
	}

	// Verify Claude settings were migrated to directory source
	settingsData, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	var settings map[string]interface{}
	json.Unmarshal(settingsData, &settings)
	marketplaces := settings["extraKnownMarketplaces"].(map[string]interface{})
	marketplace := marketplaces[MarketplaceName].(map[string]interface{})
	source := marketplace["source"].(map[string]interface{})
	if source["source"] != "directory" {
		t.Errorf("expected source 'directory', got %v", source["source"])
	}
}

// TestHandleURLScheme tests URL scheme handling
func TestHandleURLScheme(t *testing.T) {
	InitLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	claudeDir := filepath.Join(tmpDir, ".claude")
	server := NewTestServer(tmpDir, claudeDir)
	server.Initialize()

	// Setup Claude directory
	os.MkdirAll(claudeDir, 0755)
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte("{}"), 0644)

	// Test install via URL scheme
	installURL := "agenthub://install?name=test-plugin&source=github&repo=user/test-repo&ref=v1.0.0"
	server.HandleURLScheme(installURL)

	// Verify plugin was installed
	if server.PluginCount() != 1 {
		t.Errorf("expected 1 plugin after install, got %d", server.PluginCount())
	}

	entry, ok := server.GetRegistryEntry("test-plugin")
	if !ok {
		t.Fatal("test-plugin not found after URL install")
	}
	if entry.Source.Repo != "user/test-repo" {
		t.Errorf("expected repo 'user/test-repo', got %q", entry.Source.Repo)
	}
	if entry.Source.Ref != "v1.0.0" {
		t.Errorf("expected ref 'v1.0.0', got %q", entry.Source.Ref)
	}

	// Test uninstall via URL scheme
	uninstallURL := "agenthub://uninstall?name=test-plugin"
	server.HandleURLScheme(uninstallURL)

	// Verify plugin was uninstalled
	if server.PluginCount() != 0 {
		t.Errorf("expected 0 plugins after uninstall, got %d", server.PluginCount())
	}

	// Test open action (should not error)
	openURL := "agenthub://open"
	server.HandleURLScheme(openURL) // Just verifying it doesn't panic

	// Test invalid URL (should not panic)
	server.HandleURLScheme("invalid://url")

	// Test unknown action (should not panic)
	server.HandleURLScheme("agenthub://unknown")
}

// TestHandleURLInstallWithAutoEnable tests autoEnable parameter
func TestHandleURLInstallWithAutoEnable(t *testing.T) {
	InitLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	claudeDir := filepath.Join(tmpDir, ".claude")
	server := NewTestServer(tmpDir, claudeDir)
	server.Initialize()

	os.MkdirAll(claudeDir, 0755)
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte("{}"), 0644)

	// Test with autoEnable=false
	installURL := "agenthub://install?name=no-auto&source=github&repo=user/repo&autoEnable=false"
	server.HandleURLScheme(installURL)

	// Plugin should be installed but not enabled in Claude settings
	settingsData, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	var settings map[string]interface{}
	json.Unmarshal(settingsData, &settings)

	// Plugin should NOT be in enabledPlugins since autoEnable=false
	enabledPlugins, _ := settings["enabledPlugins"].(map[string]interface{})
	pluginID := "no-auto@" + MarketplaceName
	if _, exists := enabledPlugins[pluginID]; exists {
		t.Error("plugin should not be auto-enabled when autoEnable=false")
	}
}

// TestHandleURLInstallMissingParams tests error handling for missing parameters
func TestHandleURLInstallMissingParams(t *testing.T) {
	InitLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	server := NewTestServer(tmpDir, filepath.Join(tmpDir, ".claude"))
	server.Initialize()

	// Missing name - should not install
	server.HandleURLScheme("agenthub://install?source=github&repo=user/repo")
	if server.PluginCount() != 0 {
		t.Error("should not install without name parameter")
	}

	// Missing source - should not install
	server.HandleURLScheme("agenthub://install?name=test")
	if server.PluginCount() != 0 {
		t.Error("should not install without source parameter")
	}
}

// TestHandleURLUninstallNonExistent tests uninstalling a non-existent plugin
func TestHandleURLUninstallNonExistent(t *testing.T) {
	InitLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	server := NewTestServer(tmpDir, filepath.Join(tmpDir, ".claude"))
	server.Initialize()

	// Should not panic when uninstalling non-existent plugin
	server.HandleURLScheme("agenthub://uninstall?name=nonexistent")
}

// TestInitLogger tests logger initialization
func TestInitLogger(t *testing.T) {
	// Test debug mode
	InitLogger(true)
	if Logger == nil {
		t.Error("Logger should not be nil after initialization")
	}

	// Test non-debug mode
	InitLogger(false)
	if Logger == nil {
		t.Error("Logger should not be nil after initialization")
	}
}


// TestResolveSourceZipMissingURL tests zip source with missing URL
func TestResolveSourceZipMissingURL(t *testing.T) {
	InitLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	server := NewTestServer(tmpDir, filepath.Join(tmpDir, ".claude"))
	server.Initialize()

	// Zip source without URL should error
	_, err = server.ResolveSource("test", PluginSource{
		Source: "zip",
	})
	if err == nil {
		t.Error("expected error for zip source without URL")
	}
}

// TestResolveSourceGitMissingURL tests git source with missing URL
func TestResolveSourceGitMissingURL(t *testing.T) {
	InitLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	server := NewTestServer(tmpDir, filepath.Join(tmpDir, ".claude"))

	// Git source without URL should error
	_, err = server.ResolveSource("test", PluginSource{
		Source: "git",
	})
	if err == nil {
		t.Error("expected error for git source without URL")
	}
}
