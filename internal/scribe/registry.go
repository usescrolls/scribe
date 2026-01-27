package scribe

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Initialize creates the required directory structure
func (s *Server) Initialize() error {
	dirs := []string{
		filepath.Join(s.hubDir, MarketplaceDirName), // .claude-plugin/
		filepath.Join(s.hubDir, PluginsDirName),     // plugins/
		filepath.Join(s.hubDir, DataDirName),        // data/
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			Logger.Error("failed to create directory", "path", dir, "error", err)
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		Logger.Debug("ensured directory exists", "path", dir)
	}

	return nil
}

// Migrate migrates from old plugins.json format to new registry format
func (s *Server) Migrate() error {
	oldPath := filepath.Join(s.hubDir, OldPluginsFile)

	// Check if old format exists
	data, err := os.ReadFile(oldPath)
	if os.IsNotExist(err) {
		Logger.Debug("no old plugins.json found, nothing to migrate")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read old plugins file: %w", err)
	}

	Logger.Info("migrating from old plugins.json format")

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
		Logger.Warn("failed to migrate claude settings", "error", err)
	}

	// Remove old file
	if err := os.Remove(oldPath); err != nil {
		Logger.Warn("failed to remove old plugins.json", "path", oldPath, "error", err)
	}

	Logger.Info("migration completed successfully", "plugin_count", len(oldPlugins))
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
		Logger.Info("updating marketplace source from URL to directory")
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
		Logger.Info("claude settings migrated to directory source")
	}

	return nil
}

// Load reads persisted registry from disk
func (s *Server) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	registryFile := filepath.Join(s.hubDir, DataDirName, RegistryFile)
	Logger.Debug("loading registry from disk", "path", registryFile)

	data, err := os.ReadFile(registryFile)
	if os.IsNotExist(err) {
		Logger.Debug("no existing registry file found, starting fresh")
		return nil
	}
	if err != nil {
		Logger.Error("failed to read registry file", "path", registryFile, "error", err)
		return err
	}

	var entries []RegistryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		Logger.Error("failed to parse registry file", "path", registryFile, "error", err)
		return err
	}

	for _, e := range entries {
		s.registry[e.Name] = e
	}
	Logger.Info("loaded registry from disk", "count", len(entries))
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
	Logger.Debug("saving registry to disk", "path", registryFile, "count", len(entries))

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		Logger.Error("failed to marshal registry", "error", err)
		return err
	}

	if err := os.WriteFile(registryFile, data, 0644); err != nil {
		Logger.Error("failed to write registry file", "path", registryFile, "error", err)
		return err
	}

	Logger.Info("saved registry to disk", "count", len(entries))
	return nil
}
