package scribe

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

// GetAllPlugins returns all registry entries
func (s *Server) GetAllPlugins() []RegistryEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entries := make([]RegistryEntry, 0, len(s.registry))
	for _, e := range s.registry {
		entries = append(entries, e)
	}
	return entries
}

// UninstallPlugin removes a single plugin from the registry
func (s *Server) UninstallPlugin(name string) error {
	s.mu.Lock()
	entry, existed := s.registry[name]
	if existed {
		// Delete plugin files for relative sources
		if resolved, ok := entry.ResolvedSource.(string); ok {
			if strings.HasPrefix(resolved, "./plugins/") {
				pluginDir := filepath.Join(s.hubDir, PluginsDirName, name)
				os.RemoveAll(pluginDir)
			}
		}
		delete(s.registry, name)
	}
	s.mu.Unlock()

	if !existed {
		return fmt.Errorf("plugin not found: %s", name)
	}

	Logger.Info("uninstalling plugin", "name", name)

	if err := s.SaveRegistry(); err != nil {
		Logger.Error("failed to save after uninstalling plugin", "error", err)
		return err
	}

	if err := s.GenerateMarketplace(); err != nil {
		Logger.Error("failed to regenerate marketplace after uninstalling plugin", "error", err)
	}

	if err := s.UpdateClaudeSettings(name, false); err != nil {
		Logger.Warn("failed to disable plugin in claude settings", "plugin", name, "error", err)
	}

	Logger.Info("plugin uninstalled successfully", "name", name)
	return nil
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

	Logger.Info("uninstalling all plugins", "count", len(pluginNames))

	// Save empty registry
	if err := s.SaveRegistry(); err != nil {
		Logger.Error("failed to save after uninstalling plugins", "error", err)
		return err
	}

	// Regenerate marketplace.json
	if err := s.GenerateMarketplace(); err != nil {
		Logger.Error("failed to regenerate marketplace after uninstalling plugins", "error", err)
	}

	// Disable each plugin in Claude settings
	for _, name := range pluginNames {
		if err := s.UpdateClaudeSettings(name, false); err != nil {
			Logger.Warn("failed to disable plugin in claude settings", "plugin", name, "error", err)
		}
	}

	Logger.Info("all plugins uninstalled successfully")
	return nil
}

// DeletePluginData removes the registry file and downloaded plugins
func (s *Server) DeletePluginData() error {
	s.mu.Lock()
	s.registry = make(map[string]RegistryEntry)
	s.mu.Unlock()

	// Delete registry file
	registryFile := filepath.Join(s.hubDir, DataDirName, RegistryFile)
	Logger.Info("deleting registry file", "path", registryFile)
	if err := os.Remove(registryFile); err != nil && !os.IsNotExist(err) {
		Logger.Error("failed to delete registry file", "error", err)
		return err
	}

	// Delete downloaded plugins directory
	pluginsDir := filepath.Join(s.hubDir, PluginsDirName)
	Logger.Info("deleting plugins directory", "path", pluginsDir)
	if err := os.RemoveAll(pluginsDir); err != nil {
		Logger.Error("failed to delete plugins directory", "error", err)
		return err
	}
	// Recreate empty plugins directory
	os.MkdirAll(pluginsDir, 0755)

	// Regenerate empty marketplace.json
	if err := s.GenerateMarketplace(); err != nil {
		Logger.Error("failed to regenerate marketplace", "error", err)
	}

	Logger.Info("plugin data deleted")
	return nil
}

// FullReset performs all reset operations: uninstall plugins, clear claude settings, delete data
func (s *Server) FullReset() error {
	Logger.Info("performing full reset")

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

	Logger.Info("full reset completed successfully")
	return nil
}
