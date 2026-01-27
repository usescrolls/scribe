package main

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

	if server.registry == nil {
		t.Error("registry map is nil")
	}

	homeDir, _ := os.UserHomeDir()
	expectedHubDir := filepath.Join(homeDir, HubDirName)
	expectedClaudeDir := filepath.Join(homeDir, ".claude")

	if server.hubDir != expectedHubDir {
		t.Errorf("expected hubDir %q, got %q", expectedHubDir, server.hubDir)
	}

	if server.claudeDir != expectedClaudeDir {
		t.Errorf("expected claudeDir %q, got %q", expectedClaudeDir, server.claudeDir)
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
	initLogger(false)

	server := &Server{
		registry:  make(map[string]RegistryEntry),
		hubDir:    tmpDir,
		claudeDir: filepath.Join(tmpDir, ".claude"),
	}

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
			name:           "install with npm source",
			url:            "agenthub://install?name=my-plugin&source=npm&package=@scope/package",
			expectedAction: "install",
			expectedParams: map[string]string{
				"name":    "my-plugin",
				"source":  "npm",
				"package": "@scope/package",
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
			action, err := parseURLScheme(tt.url)

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
	initLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	server := &Server{
		registry:  make(map[string]RegistryEntry),
		hubDir:    tmpDir,
		claudeDir: filepath.Join(tmpDir, ".claude"),
	}
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
			name:       "npm source",
			pluginName: "npm-plugin",
			source: PluginSource{
				Source:  "npm",
				Package: "@scope/package",
			},
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("expected map result")
				}
				if m["source"] != "npm" {
					t.Errorf("expected source=npm, got %v", m["source"])
				}
				if m["package"] != "@scope/package" {
					t.Errorf("expected package=@scope/package, got %v", m["package"])
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
			name:       "npm missing package",
			pluginName: "bad-plugin",
			source: PluginSource{
				Source: "npm",
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
			result, err := server.resolveSource(tt.pluginName, tt.source)

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
	initLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	server := &Server{
		registry:  make(map[string]RegistryEntry),
		hubDir:    tmpDir,
		claudeDir: filepath.Join(tmpDir, ".claude"),
	}
	server.Initialize()

	// Add some entries
	now := time.Now()
	server.registry["plugin1"] = RegistryEntry{
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
	}
	server.registry["plugin2"] = RegistryEntry{
		Name:    "plugin2",
		Version: "2.0.0",
		Source: PluginSource{
			Source:  "npm",
			Package: "@test/plugin2",
		},
		ResolvedSource: map[string]interface{}{
			"source":  "npm",
			"package": "@test/plugin2",
		},
		InstalledAt: now,
	}

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
	server2 := &Server{
		registry:  make(map[string]RegistryEntry),
		hubDir:    tmpDir,
		claudeDir: filepath.Join(tmpDir, ".claude"),
	}

	if err := server2.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify loaded data
	if len(server2.registry) != 2 {
		t.Errorf("expected 2 entries, got %d", len(server2.registry))
	}

	entry1, ok := server2.registry["plugin1"]
	if !ok {
		t.Fatal("plugin1 not found in loaded registry")
	}
	if entry1.Description != "Test plugin 1" {
		t.Errorf("expected description 'Test plugin 1', got %q", entry1.Description)
	}
	if entry1.Source.Repo != "user/plugin1" {
		t.Errorf("expected repo 'user/plugin1', got %q", entry1.Source.Repo)
	}

	entry2, ok := server2.registry["plugin2"]
	if !ok {
		t.Fatal("plugin2 not found in loaded registry")
	}
	if entry2.Source.Package != "@test/plugin2" {
		t.Errorf("expected package '@test/plugin2', got %q", entry2.Source.Package)
	}
}

// TestGenerateMarketplace tests marketplace.json generation
func TestGenerateMarketplace(t *testing.T) {
	initLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	server := &Server{
		registry:  make(map[string]RegistryEntry),
		hubDir:    tmpDir,
		claudeDir: filepath.Join(tmpDir, ".claude"),
	}
	server.Initialize()

	// Add a plugin
	server.registry["test-plugin"] = RegistryEntry{
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
	}

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
	server := &Server{
		registry: make(map[string]RegistryEntry),
	}

	if server.PluginCount() != 0 {
		t.Errorf("expected 0, got %d", server.PluginCount())
	}

	server.registry["plugin1"] = RegistryEntry{Name: "plugin1"}
	server.registry["plugin2"] = RegistryEntry{Name: "plugin2"}

	if server.PluginCount() != 2 {
		t.Errorf("expected 2, got %d", server.PluginCount())
	}
}

// TestUpdateClaudeSettings tests Claude settings management
func TestUpdateClaudeSettings(t *testing.T) {
	initLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	claudeDir := filepath.Join(tmpDir, ".claude")
	server := &Server{
		registry:  make(map[string]RegistryEntry),
		hubDir:    tmpDir,
		claudeDir: claudeDir,
	}

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
	server := &Server{
		claudeDir: claudeDir,
	}

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
	initLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	claudeDir := filepath.Join(tmpDir, ".claude")
	server := &Server{
		registry:  make(map[string]RegistryEntry),
		hubDir:    tmpDir,
		claudeDir: claudeDir,
	}
	server.Initialize()

	// Add a plugin
	server.registry["test-plugin"] = RegistryEntry{
		Name: "test-plugin",
		ResolvedSource: map[string]interface{}{
			"source": "github",
			"repo":   "test/plugin",
		},
	}
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
	initLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	server := &Server{
		registry: make(map[string]RegistryEntry),
		hubDir:   tmpDir,
	}

	// Should not error when registry file doesn't exist
	if err := server.Load(); err != nil {
		t.Errorf("Load should not error for non-existent registry: %v", err)
	}

	if len(server.registry) != 0 {
		t.Errorf("registry should be empty, got %d entries", len(server.registry))
	}
}

// TestDeletePluginData tests the DeletePluginData method
func TestDeletePluginData(t *testing.T) {
	initLogger(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	server := &Server{
		registry:  make(map[string]RegistryEntry),
		hubDir:    tmpDir,
		claudeDir: filepath.Join(tmpDir, ".claude"),
	}
	server.Initialize()

	// Add some data
	server.registry["plugin1"] = RegistryEntry{Name: "plugin1"}
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
	if len(server.registry) != 0 {
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
