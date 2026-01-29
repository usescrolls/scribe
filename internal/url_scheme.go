package scribe

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

// ParseURLScheme parses an agenthub:// URL and returns the action and parameters
func ParseURLScheme(urlStr string) (*URLAction, error) {
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

// HandleURLScheme processes an agenthub:// URL passed as a command line argument
func (s *Server) HandleURLScheme(urlStr string) {
	Logger.Info("handling URL scheme", "url", urlStr)

	action, err := ParseURLScheme(urlStr)
	if err != nil {
		Logger.Error("failed to parse URL scheme", "url", urlStr, "error", err)
		return
	}

	switch action.Action {
	case "install":
		s.handleURLInstall(action.Params)
	case "uninstall":
		s.handleURLUninstall(action.Params)
	case "open":
		Logger.Info("open action received - app is now in foreground")
		// The app being launched is sufficient to bring it to foreground
	default:
		Logger.Warn("unknown URL action", "action", action.Action)
	}
}

// handleURLInstall processes an install action from a URL scheme
func (s *Server) handleURLInstall(params url.Values) {
	name := params.Get("name")
	source := params.Get("source")

	if name == "" {
		Logger.Error("URL install: missing required parameter 'name'")
		return
	}
	if source == "" {
		Logger.Error("URL install: missing required parameter 'source'")
		return
	}

	Logger.Info("URL install: installing plugin",
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
	resolvedSource, err := s.ResolveSource(name, pluginSource)
	if err != nil {
		Logger.Error("URL install: failed to resolve source", "name", name, "error", err)
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
	s.mu.Lock()
	s.registry[name] = entry
	s.mu.Unlock()

	// Persist registry
	if err := s.SaveRegistry(); err != nil {
		Logger.Error("URL install: failed to save registry", "error", err)
		return
	}

	// Regenerate marketplace.json
	if err := s.GenerateMarketplace(); err != nil {
		Logger.Error("URL install: failed to regenerate marketplace", "error", err)
		return
	}

	// Auto-enable by default (unless explicitly disabled)
	autoEnable := params.Get("autoEnable")
	if autoEnable != "false" {
		if err := s.UpdateClaudeSettings(name, true); err != nil {
			Logger.Warn("URL install: failed to update claude settings", "name", name, "error", err)
		}
	}

	Logger.Info("URL install: plugin installed successfully", "name", name)
}

// handleURLUninstall processes an uninstall action from a URL scheme
func (s *Server) handleURLUninstall(params url.Values) {
	name := params.Get("name")

	if name == "" {
		Logger.Error("URL uninstall: missing required parameter 'name'")
		return
	}

	Logger.Info("URL uninstall: uninstalling plugin", "name", name)

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
		Logger.Warn("URL uninstall: plugin not found", "name", name)
		return
	}

	if err := s.SaveRegistry(); err != nil {
		Logger.Error("URL uninstall: failed to save registry", "error", err)
	}

	if err := s.GenerateMarketplace(); err != nil {
		Logger.Error("URL uninstall: failed to regenerate marketplace", "error", err)
	}

	if err := s.UpdateClaudeSettings(name, false); err != nil {
		Logger.Warn("URL uninstall: failed to update claude settings", "name", name, "error", err)
	}

	Logger.Info("URL uninstall: plugin uninstalled successfully", "name", name)
}
