package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
		logger.Info("registering useScrolls.com marketplace in claude settings", "path", s.hubDir)
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

// ClearClaudePluginSettings removes all plugin entries from Claude settings
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

	// Remove all plugins from enabledPlugins
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
