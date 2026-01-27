package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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
