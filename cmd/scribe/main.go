package main

import (
	"archive/zip"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

//go:embed icon.png
var iconData []byte

const (
	MarketplaceName  = "agenthub"
	MarketplaceOwner = "AgentHub"
	Version          = "1.0.0"

	// Directory structure constants
	HubDirName         = ".agenthub-middleware"
	MarketplaceDirName = ".claude-plugin"
	MarketplaceFile    = "marketplace.json"
	PluginsDirName     = "plugins"
	DataDirName        = "data"
	RegistryFile       = "registry.json"
	OldPluginsFile     = "plugins.json" // For migration
)

// logger is the global structured logger instance
var logger *slog.Logger

// initLogger initializes the structured logger
func initLogger(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	logger = slog.New(handler)
	slog.SetDefault(logger)
}

// PluginSource represents the source of a plugin
type PluginSource struct {
	Source  string `json:"source"`            // github, npm, git, url, zip
	Repo    string `json:"repo,omitempty"`    // for github
	Package string `json:"package,omitempty"` // for npm
	URL     string `json:"url,omitempty"`     // for git/url/zip
	Ref     string `json:"ref,omitempty"`     // branch/tag
}

// PluginProvides describes what extension types a plugin provides (for multi-skill plugins)
type PluginProvides struct {
	Skills   []string `json:"skills,omitempty"`   // List of skill names this plugin provides
	Agents   []string `json:"agents,omitempty"`   // List of subagent names this plugin provides
	Commands []string `json:"commands,omitempty"` // List of command names this plugin provides
	Hooks    []string `json:"hooks,omitempty"`    // List of hook names this plugin provides
}


// Plugin represents a plugin entry
type Plugin struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Version     string          `json:"version,omitempty"`
	Category    string          `json:"category,omitempty"`
	Source      PluginSource    `json:"source"`
	Author      *Author         `json:"author,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	Provides    *PluginProvides `json:"provides,omitempty"` // What extension types this plugin provides (for multi-skill plugins)
}

// Author represents a plugin author
// It can be unmarshaled from either a string or an object {"name": "..."}
type Author struct {
	Name string `json:"name"`
}

// UnmarshalJSON allows Author to be unmarshaled from either a string or an object
func (a *Author) UnmarshalJSON(data []byte) error {
	// First, try to unmarshal as a string
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		a.Name = name
		return nil
	}

	// If that fails, try to unmarshal as an object
	type authorObj struct {
		Name string `json:"name"`
	}
	var obj authorObj
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	a.Name = obj.Name
	return nil
}

// RegistryEntry represents a plugin in the middleware's internal registry
type RegistryEntry struct {
	Name           string          `json:"name"`
	Description    string          `json:"description,omitempty"`
	Version        string          `json:"version,omitempty"`
	Category       string          `json:"category,omitempty"`
	Author         *Author         `json:"author,omitempty"`
	Tags           []string        `json:"tags,omitempty"`
	Source         PluginSource    `json:"source"`                   // Original source from request
	ResolvedSource interface{}     `json:"resolvedSource"`           // What goes in marketplace.json (string or object)
	InstalledAt    time.Time       `json:"installedAt"`
	Provides       *PluginProvides `json:"provides,omitempty"`       // What extension types this plugin provides (for multi-skill plugins)
}

// Server holds the middleware server state
type Server struct {
	mu        sync.RWMutex
	registry  map[string]RegistryEntry
	hubDir    string // ~/.agenthub-middleware
	claudeDir string // ~/.claude
}

// NewServer creates a new middleware server
func NewServer() *Server {
	homeDir, _ := os.UserHomeDir()
	return &Server{
		registry:  make(map[string]RegistryEntry),
		hubDir:    filepath.Join(homeDir, HubDirName),
		claudeDir: filepath.Join(homeDir, ".claude"),
	}
}

// Initialize creates the required directory structure
func (s *Server) Initialize() error {
	dirs := []string{
		filepath.Join(s.hubDir, MarketplaceDirName), // .claude-plugin/
		filepath.Join(s.hubDir, PluginsDirName),     // plugins/
		filepath.Join(s.hubDir, DataDirName),        // data/
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Error("failed to create directory", "path", dir, "error", err)
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		logger.Debug("ensured directory exists", "path", dir)
	}

	return nil
}

// Migrate migrates from old plugins.json format to new registry format
func (s *Server) Migrate() error {
	oldPath := filepath.Join(s.hubDir, OldPluginsFile)

	// Check if old format exists
	data, err := os.ReadFile(oldPath)
	if os.IsNotExist(err) {
		logger.Debug("no old plugins.json found, nothing to migrate")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read old plugins file: %w", err)
	}

	logger.Info("migrating from old plugins.json format")

	// Parse old format
	var oldPlugins []Plugin
	if err := json.Unmarshal(data, &oldPlugins); err != nil {
		return fmt.Errorf("failed to parse old plugins file: %w", err)
	}

	// Convert to new registry format
	s.mu.Lock()
	for _, p := range oldPlugins {
		// For existing plugins, the source is the resolved source (pass-through)
		var resolvedSource interface{}
		switch p.Source.Source {
		case "github":
			resolvedSource = map[string]interface{}{
				"source": "github",
				"repo":   p.Source.Repo,
			}
			if p.Source.Ref != "" {
				resolvedSource.(map[string]interface{})["ref"] = p.Source.Ref
			}
		case "npm":
			resolvedSource = map[string]interface{}{
				"source":  "npm",
				"package": p.Source.Package,
			}
		case "git", "url":
			resolvedSource = map[string]interface{}{
				"source": "url",
				"url":    p.Source.URL,
			}
			if p.Source.Ref != "" {
				resolvedSource.(map[string]interface{})["ref"] = p.Source.Ref
			}
		default:
			// Keep as-is for unknown types
			resolvedSource = p.Source
		}

		s.registry[p.Name] = RegistryEntry{
			Name:           p.Name,
			Description:    p.Description,
			Version:        p.Version,
			Category:       p.Category,
			Author:         p.Author,
			Tags:           p.Tags,
			Source:         p.Source,
			ResolvedSource: resolvedSource,
			InstalledAt:    time.Now(),
		}
	}
	s.mu.Unlock()

	// Save in new format
	if err := s.SaveRegistry(); err != nil {
		return fmt.Errorf("failed to save migrated registry: %w", err)
	}

	// Generate marketplace.json
	if err := s.GenerateMarketplace(); err != nil {
		return fmt.Errorf("failed to generate marketplace after migration: %w", err)
	}

	// Update Claude settings to use directory source
	if err := s.migrateClaudeSettings(); err != nil {
		logger.Warn("failed to migrate claude settings", "error", err)
	}

	// Remove old file
	if err := os.Remove(oldPath); err != nil {
		logger.Warn("failed to remove old plugins.json", "path", oldPath, "error", err)
	}

	logger.Info("migration completed successfully", "plugin_count", len(oldPlugins))
	return nil
}

// migrateClaudeSettings updates Claude settings from URL to directory source
func (s *Server) migrateClaudeSettings() error {
	settingsPath := filepath.Join(s.claudeDir, "settings.json")

	data, err := os.ReadFile(settingsPath)
	if os.IsNotExist(err) {
		return nil // No settings to migrate
	}
	if err != nil {
		return err
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}

	// Check if our marketplace exists and update it
	marketplaces, ok := settings["extraKnownMarketplaces"].(map[string]interface{})
	if !ok {
		return nil // No marketplaces
	}

	marketplace, ok := marketplaces[MarketplaceName].(map[string]interface{})
	if !ok {
		return nil // Our marketplace not found
	}

	source, ok := marketplace["source"].(map[string]interface{})
	if !ok {
		return nil
	}

	// Check if it's still URL-based
	if source["source"] == "url" {
		logger.Info("updating marketplace source from URL to directory")
		marketplace["source"] = map[string]interface{}{
			"source": "directory",
			"path":   s.hubDir,
		}

		output, err := json.MarshalIndent(settings, "", "  ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(settingsPath, output, 0644); err != nil {
			return err
		}
		logger.Info("claude settings migrated to directory source")
	}

	return nil
}

// Load reads persisted registry from disk
func (s *Server) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	registryFile := filepath.Join(s.hubDir, DataDirName, RegistryFile)
	logger.Debug("loading registry from disk", "path", registryFile)

	data, err := os.ReadFile(registryFile)
	if os.IsNotExist(err) {
		logger.Debug("no existing registry file found, starting fresh")
		return nil
	}
	if err != nil {
		logger.Error("failed to read registry file", "path", registryFile, "error", err)
		return err
	}

	var entries []RegistryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		logger.Error("failed to parse registry file", "path", registryFile, "error", err)
		return err
	}

	for _, e := range entries {
		s.registry[e.Name] = e
	}
	logger.Info("loaded registry from disk", "count", len(entries))
	return nil
}

// SaveRegistry persists registry to disk
func (s *Server) SaveRegistry() error {
	s.mu.RLock()
	entries := make([]RegistryEntry, 0, len(s.registry))
	for _, e := range s.registry {
		entries = append(entries, e)
	}
	s.mu.RUnlock()

	registryFile := filepath.Join(s.hubDir, DataDirName, RegistryFile)
	logger.Debug("saving registry to disk", "path", registryFile, "count", len(entries))

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		logger.Error("failed to marshal registry", "error", err)
		return err
	}

	if err := os.WriteFile(registryFile, data, 0644); err != nil {
		logger.Error("failed to write registry file", "path", registryFile, "error", err)
		return err
	}

	logger.Info("saved registry to disk", "count", len(entries))
	return nil
}

// GenerateMarketplace writes the marketplace.json file to the filesystem
func (s *Server) GenerateMarketplace() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Build plugins array for marketplace.json using resolved sources
	plugins := make([]map[string]interface{}, 0, len(s.registry))
	for _, entry := range s.registry {
		plugin := map[string]interface{}{
			"name":   entry.Name,
			"source": entry.ResolvedSource,
		}
		if entry.Description != "" {
			plugin["description"] = entry.Description
		}
		if entry.Version != "" {
			plugin["version"] = entry.Version
		}
		if entry.Category != "" {
			plugin["category"] = entry.Category
		}
		if entry.Author != nil {
			plugin["author"] = entry.Author
		}
		if len(entry.Tags) > 0 {
			plugin["tags"] = entry.Tags
		}
		// Include provides field for multi-skill plugins
		if entry.Provides != nil {
			plugin["provides"] = entry.Provides
		}
		plugins = append(plugins, plugin)
	}

	marketplace := map[string]interface{}{
		"name":        MarketplaceName,
		"description": "AgentHub Plugin Marketplace - One-click installs from agenthub.dev",
		"owner":       map[string]string{"name": MarketplaceOwner},
		"version":     Version,
		"plugins":     plugins,
	}

	marketplacePath := filepath.Join(s.hubDir, MarketplaceDirName, MarketplaceFile)
	logger.Debug("writing marketplace.json", "path", marketplacePath, "plugin_count", len(plugins))

	data, err := json.MarshalIndent(marketplace, "", "  ")
	if err != nil {
		logger.Error("failed to marshal marketplace", "error", err)
		return err
	}

	if err := os.WriteFile(marketplacePath, data, 0644); err != nil {
		logger.Error("failed to write marketplace file", "path", marketplacePath, "error", err)
		return err
	}

	logger.Info("marketplace.json updated", "plugin_count", len(plugins))
	return nil
}

// resolveSource resolves the plugin source based on type
// For github/npm/git/url sources, it passes through the source definition
// For zip sources, it downloads and extracts the plugin files locally
func (s *Server) resolveSource(name string, source PluginSource) (interface{}, error) {
	switch source.Source {
	case "github":
		// Pass through - Claude Code handles GitHub sources directly
		if source.Repo == "" {
			return nil, fmt.Errorf("github source requires repo")
		}
		resolved := map[string]interface{}{
			"source": "github",
			"repo":   source.Repo,
		}
		if source.Ref != "" {
			resolved["ref"] = source.Ref
		}
		return resolved, nil

	case "npm":
		// Pass through - Claude Code handles npm sources directly
		if source.Package == "" {
			return nil, fmt.Errorf("npm source requires package")
		}
		resolved := map[string]interface{}{
			"source":  "npm",
			"package": source.Package,
		}
		return resolved, nil

	case "git", "url":
		// Pass through - Claude Code handles git URLs directly
		if source.URL == "" {
			return nil, fmt.Errorf("git/url source requires url")
		}
		resolved := map[string]interface{}{
			"source": "url",
			"url":    source.URL,
		}
		if source.Ref != "" {
			resolved["ref"] = source.Ref
		}
		return resolved, nil

	case "zip":
		// Download zip file and extract to plugins directory
		if source.URL == "" {
			return nil, fmt.Errorf("zip source requires url")
		}
		if err := s.downloadZip(name, source.URL); err != nil {
			return nil, err
		}
		return "./plugins/" + name, nil

	default:
		return nil, fmt.Errorf("unsupported source type: %s (supported: github, npm, git, url, zip)", source.Source)
	}
}

// downloadZip downloads a zip file from a URL and extracts it to the plugins directory
func (s *Server) downloadZip(name string, zipURL string) error {
	targetDir := filepath.Join(s.hubDir, PluginsDirName, name)
	logger.Info("downloading zip plugin", "name", name, "url", zipURL, "target", targetDir)

	// Remove existing if present
	if err := os.RemoveAll(targetDir); err != nil {
		logger.Warn("failed to remove existing plugin directory", "path", targetDir, "error", err)
	}

	// Download the zip file
	resp, err := http.Get(zipURL)
	if err != nil {
		return fmt.Errorf("failed to download zip: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download zip: status %d", resp.StatusCode)
	}

	// Create a temporary file for the zip
	tmpFile, err := os.CreateTemp("", "plugin-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Copy the response body to the temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to save zip: %w", err)
	}
	tmpFile.Close()

	// Open the zip file
	zipReader, err := zip.OpenReader(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer zipReader.Close()

	// Create the target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Extract files
	for _, file := range zipReader.File {
		// Handle the case where zip contains a root folder with the plugin name
		// e.g., test-runner/plugin.json should extract to targetDir/plugin.json
		filePath := file.Name

		// Strip the first directory component if it matches the plugin name
		parts := strings.SplitN(filePath, "/", 2)
		if len(parts) > 1 && parts[0] == name {
			filePath = parts[1]
		}

		if filePath == "" {
			continue
		}

		destPath := filepath.Join(targetDir, filePath)

		// Check for zip slip vulnerability
		if !strings.HasPrefix(destPath, filepath.Clean(targetDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path in zip: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(destPath, file.Mode())
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Extract file
		srcFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %w", err)
		}

		dstFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			srcFile.Close()
			return fmt.Errorf("failed to create file: %w", err)
		}

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			srcFile.Close()
			dstFile.Close()
			return fmt.Errorf("failed to extract file: %w", err)
		}

		srcFile.Close()
		dstFile.Close()
	}

	logger.Info("plugin extracted successfully", "name", name)
	return nil
}

// UpdateClaudeSettings updates ~/.claude/settings.json
func (s *Server) UpdateClaudeSettings(pluginName string, enabled bool) error {
	settingsPath := filepath.Join(s.claudeDir, "settings.json")
	action := "enabling"
	if !enabled {
		action = "disabling"
	}
	logger.Debug("updating claude settings", "plugin", pluginName, "action", action, "path", settingsPath)

	// Ensure .claude directory exists
	if err := os.MkdirAll(s.claudeDir, 0755); err != nil {
		logger.Error("failed to create .claude directory", "path", s.claudeDir, "error", err)
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	// Read existing settings or create new
	var settings map[string]interface{}
	data, err := os.ReadFile(settingsPath)
	if os.IsNotExist(err) {
		logger.Debug("claude settings file not found, creating new", "path", settingsPath)
		settings = make(map[string]interface{})
	} else if err != nil {
		logger.Error("failed to read claude settings", "path", settingsPath, "error", err)
		return fmt.Errorf("failed to read settings: %w", err)
	} else {
		if err := json.Unmarshal(data, &settings); err != nil {
			logger.Error("failed to parse claude settings", "path", settingsPath, "error", err)
			return fmt.Errorf("failed to parse settings: %w", err)
		}
	}

	// Ensure extraKnownMarketplaces exists with our marketplace
	marketplaces, ok := settings["extraKnownMarketplaces"].(map[string]interface{})
	if !ok {
		marketplaces = make(map[string]interface{})
		settings["extraKnownMarketplaces"] = marketplaces
	}

	// Add our marketplace if not present (using directory source)
	if _, exists := marketplaces[MarketplaceName]; !exists {
		logger.Info("registering agenthub marketplace in claude settings", "path", s.hubDir)
		marketplaces[MarketplaceName] = map[string]interface{}{
			"source": map[string]interface{}{
				"source": "directory",
				"path":   s.hubDir,
			},
		}
	}

	// Update enabledPlugins
	enabledPlugins, ok := settings["enabledPlugins"].(map[string]interface{})
	if !ok {
		enabledPlugins = make(map[string]interface{})
		settings["enabledPlugins"] = enabledPlugins
	}

	pluginID := fmt.Sprintf("%s@%s", pluginName, MarketplaceName)
	if enabled {
		enabledPlugins[pluginID] = true
	} else {
		delete(enabledPlugins, pluginID)
	}

	// Write back
	output, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		logger.Error("failed to marshal claude settings", "error", err)
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, output, 0644); err != nil {
		logger.Error("failed to write claude settings", "path", settingsPath, "error", err)
		return fmt.Errorf("failed to write settings: %w", err)
	}

	logger.Info("updated claude settings", "plugin", pluginName, "action", action, "plugin_id", pluginID)
	return nil
}

// ClaudeCodeDetected checks if Claude Code is installed
func (s *Server) ClaudeCodeDetected() bool {
	settingsPath := filepath.Join(s.claudeDir, "settings.json")
	_, err := os.Stat(settingsPath)
	return err == nil
}

// PluginCount returns the number of installed plugins
func (s *Server) PluginCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.registry)
}

// UninstallAllPlugins removes all plugins from memory and disk
func (s *Server) UninstallAllPlugins() error {
	s.mu.Lock()
	pluginNames := make([]string, 0, len(s.registry))
	for name, entry := range s.registry {
		pluginNames = append(pluginNames, name)
		// Delete plugin files for relative sources
		if resolved, ok := entry.ResolvedSource.(string); ok {
			if strings.HasPrefix(resolved, "./plugins/") {
				pluginDir := filepath.Join(s.hubDir, PluginsDirName, name)
				os.RemoveAll(pluginDir)
			}
		}
	}
	s.registry = make(map[string]RegistryEntry)
	s.mu.Unlock()

	logger.Info("uninstalling all plugins", "count", len(pluginNames))

	// Save empty registry
	if err := s.SaveRegistry(); err != nil {
		logger.Error("failed to save after uninstalling plugins", "error", err)
		return err
	}

	// Regenerate marketplace.json
	if err := s.GenerateMarketplace(); err != nil {
		logger.Error("failed to regenerate marketplace after uninstalling plugins", "error", err)
	}

	// Disable each plugin in Claude settings
	for _, name := range pluginNames {
		if err := s.UpdateClaudeSettings(name, false); err != nil {
			logger.Warn("failed to disable plugin in claude settings", "plugin", name, "error", err)
		}
	}

	logger.Info("all plugins uninstalled successfully")
	return nil
}

// ClearClaudePluginSettings removes all agenthub plugin entries from Claude settings
func (s *Server) ClearClaudePluginSettings() error {
	settingsPath := filepath.Join(s.claudeDir, "settings.json")
	logger.Info("clearing claude plugin settings", "path", settingsPath)

	data, err := os.ReadFile(settingsPath)
	if os.IsNotExist(err) {
		logger.Debug("claude settings file not found, nothing to clear")
		return nil
	}
	if err != nil {
		logger.Error("failed to read claude settings", "error", err)
		return err
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		logger.Error("failed to parse claude settings", "error", err)
		return err
	}

	// Remove our marketplace from extraKnownMarketplaces
	if marketplaces, ok := settings["extraKnownMarketplaces"].(map[string]interface{}); ok {
		delete(marketplaces, MarketplaceName)
		if len(marketplaces) == 0 {
			delete(settings, "extraKnownMarketplaces")
		}
	}

	// Remove all agenthub plugins from enabledPlugins
	if enabledPlugins, ok := settings["enabledPlugins"].(map[string]interface{}); ok {
		for key := range enabledPlugins {
			if strings.HasSuffix(key, "@"+MarketplaceName) {
				delete(enabledPlugins, key)
			}
		}
		if len(enabledPlugins) == 0 {
			delete(settings, "enabledPlugins")
		}
	}

	output, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		logger.Error("failed to marshal claude settings", "error", err)
		return err
	}

	if err := os.WriteFile(settingsPath, output, 0644); err != nil {
		logger.Error("failed to write claude settings", "error", err)
		return err
	}

	logger.Info("claude plugin settings cleared")
	return nil
}

// DeletePluginData removes the registry file and downloaded plugins
func (s *Server) DeletePluginData() error {
	s.mu.Lock()
	s.registry = make(map[string]RegistryEntry)
	s.mu.Unlock()

	// Delete registry file
	registryFile := filepath.Join(s.hubDir, DataDirName, RegistryFile)
	logger.Info("deleting registry file", "path", registryFile)
	if err := os.Remove(registryFile); err != nil && !os.IsNotExist(err) {
		logger.Error("failed to delete registry file", "error", err)
		return err
	}

	// Delete downloaded plugins directory
	pluginsDir := filepath.Join(s.hubDir, PluginsDirName)
	logger.Info("deleting plugins directory", "path", pluginsDir)
	if err := os.RemoveAll(pluginsDir); err != nil {
		logger.Error("failed to delete plugins directory", "error", err)
		return err
	}
	// Recreate empty plugins directory
	os.MkdirAll(pluginsDir, 0755)

	// Regenerate empty marketplace.json
	if err := s.GenerateMarketplace(); err != nil {
		logger.Error("failed to regenerate marketplace", "error", err)
	}

	logger.Info("plugin data deleted")
	return nil
}

// FullReset performs all reset operations: uninstall plugins, clear claude settings, delete data
func (s *Server) FullReset() error {
	logger.Info("performing full reset")

	var errors []string

	if err := s.UninstallAllPlugins(); err != nil {
		errors = append(errors, fmt.Sprintf("uninstall plugins: %v", err))
	}

	if err := s.ClearClaudePluginSettings(); err != nil {
		errors = append(errors, fmt.Sprintf("clear claude settings: %v", err))
	}

	if err := s.DeletePluginData(); err != nil {
		errors = append(errors, fmt.Sprintf("delete plugin data: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("full reset completed with errors: %s", strings.Join(errors, "; "))
	}

	logger.Info("full reset completed successfully")
	return nil
}

// Global server instance for systray callbacks
var server *Server

// URLAction represents a parsed agenthub:// URL
type URLAction struct {
	Action string
	Params url.Values
}

// parseURLScheme parses an agenthub:// URL and returns the action and parameters
func parseURLScheme(urlStr string) (*URLAction, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if u.Scheme != "agenthub" {
		return nil, fmt.Errorf("invalid scheme: %s (expected agenthub)", u.Scheme)
	}

	// The host part is the action (install, uninstall, open)
	action := u.Host
	if action == "" {
		// Try opaque form: agenthub:install?...
		action = u.Opaque
		if idx := strings.Index(action, "?"); idx != -1 {
			action = action[:idx]
		}
	}

	return &URLAction{
		Action: action,
		Params: u.Query(),
	}, nil
}

// handleURLScheme processes an agenthub:// URL passed as a command line argument
func handleURLScheme(urlStr string) {
	logger.Info("handling URL scheme", "url", urlStr)

	action, err := parseURLScheme(urlStr)
	if err != nil {
		logger.Error("failed to parse URL scheme", "url", urlStr, "error", err)
		return
	}

	switch action.Action {
	case "install":
		handleURLInstall(action.Params)
	case "uninstall":
		handleURLUninstall(action.Params)
	case "open":
		logger.Info("open action received - app is now in foreground")
		// The app being launched is sufficient to bring it to foreground
	default:
		logger.Warn("unknown URL action", "action", action.Action)
	}
}

// handleURLInstall processes an install action from a URL scheme
func handleURLInstall(params url.Values) {
	name := params.Get("name")
	source := params.Get("source")

	if name == "" {
		logger.Error("URL install: missing required parameter 'name'")
		return
	}
	if source == "" {
		logger.Error("URL install: missing required parameter 'source'")
		return
	}

	logger.Info("URL install: installing plugin",
		"name", name,
		"source", source,
		"repo", params.Get("repo"),
		"package", params.Get("package"),
		"url", params.Get("url"),
		"ref", params.Get("ref"),
	)

	// Create plugin source from URL parameters
	pluginSource := PluginSource{
		Source:  source,
		Repo:    params.Get("repo"),
		Package: params.Get("package"),
		URL:     params.Get("url"),
		Ref:     params.Get("ref"),
	}

	// Resolve the source
	resolvedSource, err := server.resolveSource(name, pluginSource)
	if err != nil {
		logger.Error("URL install: failed to resolve source", "name", name, "error", err)
		return
	}

	// Create registry entry
	entry := RegistryEntry{
		Name:           name,
		Source:         pluginSource,
		ResolvedSource: resolvedSource,
		InstalledAt:    time.Now(),
	}

	// Add to registry
	server.mu.Lock()
	server.registry[name] = entry
	server.mu.Unlock()

	// Persist registry
	if err := server.SaveRegistry(); err != nil {
		logger.Error("URL install: failed to save registry", "error", err)
		return
	}

	// Regenerate marketplace.json
	if err := server.GenerateMarketplace(); err != nil {
		logger.Error("URL install: failed to regenerate marketplace", "error", err)
		return
	}

	// Auto-enable by default (unless explicitly disabled)
	autoEnable := params.Get("autoEnable")
	if autoEnable != "false" {
		if err := server.UpdateClaudeSettings(name, true); err != nil {
			logger.Warn("URL install: failed to update claude settings", "name", name, "error", err)
		}
	}

	logger.Info("URL install: plugin installed successfully", "name", name)
}

// handleURLUninstall processes an uninstall action from a URL scheme
func handleURLUninstall(params url.Values) {
	name := params.Get("name")

	if name == "" {
		logger.Error("URL uninstall: missing required parameter 'name'")
		return
	}

	logger.Info("URL uninstall: uninstalling plugin", "name", name)

	server.mu.Lock()
	entry, existed := server.registry[name]
	if existed {
		// Delete plugin files for relative sources
		if resolved, ok := entry.ResolvedSource.(string); ok {
			if strings.HasPrefix(resolved, "./plugins/") {
				pluginDir := filepath.Join(server.hubDir, PluginsDirName, name)
				os.RemoveAll(pluginDir)
			}
		}
		delete(server.registry, name)
	}
	server.mu.Unlock()

	if !existed {
		logger.Warn("URL uninstall: plugin not found", "name", name)
		return
	}

	if err := server.SaveRegistry(); err != nil {
		logger.Error("URL uninstall: failed to save registry", "error", err)
	}

	if err := server.GenerateMarketplace(); err != nil {
		logger.Error("URL uninstall: failed to regenerate marketplace", "error", err)
	}

	if err := server.UpdateClaudeSettings(name, false); err != nil {
		logger.Warn("URL uninstall: failed to update claude settings", "name", name, "error", err)
	}

	logger.Info("URL uninstall: plugin uninstalled successfully", "name", name)
}

func main() {
	noGui := flag.Bool("no-gui", false, "Run without system tray icon")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Initialize logger
	initLogger(*debug)

	logger.Info("initializing agenthub middleware", "version", Version, "debug", *debug)

	server = NewServer()

	// Check for URL scheme argument (passed by OS when opening agenthub:// links)
	args := flag.Args()
	if len(args) > 0 && strings.HasPrefix(args[0], "agenthub://") {
		urlArg := args[0]
		logger.Info("URL scheme argument detected", "url", urlArg)

		// On Linux/Windows: try to send URL to running instance before full startup.
		// This allows the new process to exit quickly if another instance handles the URL.
		// macOS doesn't need this - Apple Events handles the "already running" case natively.
		if runtime.GOOS != "darwin" && TrySendToRunningInstance != nil {
			if TrySendToRunningInstance(urlArg) {
				logger.Info("URL forwarded to running instance, exiting")
				return
			}
			logger.Debug("no running instance found, continuing startup")
		}

		// Initialize directory structure first
		if err := server.Initialize(); err != nil {
			logger.Error("failed to initialize directory structure", "error", err)
			os.Exit(1)
		}

		// Load existing registry
		if err := server.Load(); err != nil {
			logger.Warn("failed to load existing registry", "error", err)
		}

		// Handle the URL scheme action
		handleURLScheme(urlArg)

		// Regenerate marketplace after any changes
		if err := server.GenerateMarketplace(); err != nil {
			logger.Warn("failed to generate marketplace", "error", err)
		}

		logger.Info("URL scheme handling completed")
		return
	}

	// Initialize directory structure
	if err := server.Initialize(); err != nil {
		logger.Error("failed to initialize directory structure", "error", err)
		os.Exit(1)
	}

	// Migrate from old format if needed
	if err := server.Migrate(); err != nil {
		logger.Warn("migration failed", "error", err)
	}

	// Load existing registry
	if err := server.Load(); err != nil {
		logger.Warn("failed to load existing registry", "error", err)
	}

	// Generate initial marketplace.json
	if err := server.GenerateMarketplace(); err != nil {
		logger.Warn("failed to generate initial marketplace", "error", err)
	}

	if *noGui {
		logger.Info("running in headless mode")
		// Register URL scheme handler (IPC server on Linux/Windows)
		RegisterURLSchemeHandler()
		// Block forever - IPC server handles URL scheme requests
		select {}
	} else {
		logger.Info("running with system tray")
		// Run with system tray (or headless fallback if CGO disabled)
		RunWithGUI()
	}
}

// getIcon returns the embedded icon PNG
func getIcon() []byte {
	return iconData
}
