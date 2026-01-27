package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

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
		"description": "Scribe Plugin Marketplace - One-click installs from useScrolls.com",
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
