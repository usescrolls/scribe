package scribe

import (
	"encoding/json"
	"time"
)

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
	Source         PluginSource    `json:"source"`             // Original source from request
	ResolvedSource interface{}     `json:"resolvedSource"`     // What goes in marketplace.json (string or object)
	InstalledAt    time.Time       `json:"installedAt"`
	Provides       *PluginProvides `json:"provides,omitempty"` // What extension types this plugin provides (for multi-skill plugins)
}
