package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
