package scribe

import (
	"encoding/json"
	"time"
)

// Skill represents a skill with its SKILL.md content and metadata
type Skill struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Path        string         `json:"path,omitempty"`     // Local path to SKILL.md directory
	Content     string         `json:"content,omitempty"`  // Raw SKILL.md content
	Metadata    map[string]any `json:"metadata,omitempty"` // Additional frontmatter fields
	Meta        *SkillMeta     `json:"meta,omitempty"`     // Source tracking (from .scribe-meta.json)
}

// SkillMeta tracks the source and installation info for a skill
// Stored in .scribe-meta.json sidecar file alongside SKILL.md
type SkillMeta struct {
	Source      string `json:"source"`               // e.g., "owner/repo"
	SourceType  string `json:"sourceType"`           // github, gitlab, local, url, well-known
	SourceURL   string `json:"sourceUrl,omitempty"`  // Full URL if applicable
	SkillPath   string `json:"skillPath,omitempty"`  // Path within source repo
	ContentHash string `json:"contentHash"`          // SHA256 hash of SKILL.md content
	InstalledAt string `json:"installedAt"`          // ISO 8601 timestamp
	UpdatedAt   string `json:"updatedAt"`            // ISO 8601 timestamp
}

// Agent represents a coding agent with its skill directories
// All skills are managed globally - scribe manages all agents uniformly
type Agent struct {
	ID              string `json:"id"`              // Unique identifier (e.g., "claude-code")
	DisplayName     string `json:"displayName"`     // Human-readable name (e.g., "Claude Code")
	GlobalSkillsDir string `json:"globalSkillsDir"` // Absolute path for global skills (e.g., "~/.claude/skills")
	GlobalConfigDir string `json:"globalConfigDir"` // For detection (e.g., "~/.claude")
}

// Workspace defines which skills are active globally
type Workspace struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Skills      []string `json:"skills"` // Skill names in this workspace
}

// Config represents the global Scribe configuration
type Config struct {
	ActiveWorkspace string `json:"activeWorkspace"`
}

// InstallOptions configures skill installation behavior
type InstallOptions struct {
	Agents   []string // Target specific agents (empty = all detected)
	Skills   []string // Select specific skills to install (empty = all found)
	Yes      bool     // Skip interactive prompts
	ListOnly bool     // List available skills without installing
}

// SourceInfo represents a parsed source reference
type SourceInfo struct {
	Type      string // github, gitlab, local, url, well-known, zip
	Owner     string // GitHub/GitLab owner
	Repo      string // Repository name
	Ref       string // Branch, tag, or commit
	Subpath   string // Path within repo
	URL       string // Full URL for url/well-known types
	LocalPath string // Absolute path for local sources
}

// SkillInfo is the frontend-friendly representation of a skill
type SkillInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Source      string   `json:"source"`      // Human-readable source
	SourceType  string   `json:"sourceType"`  // github, local, url, etc.
	InstalledAt string   `json:"installedAt"` // ISO formatted timestamp
	Agents      []string `json:"agents"`      // List of agent IDs with this skill
}

// WorkspaceInfo is the frontend-friendly representation of a workspace
type WorkspaceInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Skills      []string `json:"skills"`   // Skill names in this workspace
	IsActive    bool     `json:"isActive"` // Whether this is the active workspace
}

// AgentStatus represents an agent's installation status for the frontend
type AgentStatus struct {
	ID              string `json:"id"`
	DisplayName     string `json:"displayName"`
	Installed       bool   `json:"installed"`
	SkillCount      int    `json:"skillCount"`
	GlobalSkillsDir string `json:"globalSkillsDir"`
}

// ======================================================================
// Legacy types (kept for backwards compatibility during transition)
// ======================================================================

// PluginSource represents the source of a plugin (LEGACY)
type PluginSource struct {
	Source  string `json:"source"`            // github, git, url, zip
	Repo    string `json:"repo,omitempty"`    // for github
	Package string `json:"package,omitempty"` // deprecated: was for npm
	URL     string `json:"url,omitempty"`     // for git/url/zip
	Ref     string `json:"ref,omitempty"`     // branch/tag
}

// PluginProvides describes what extension types a plugin provides (LEGACY)
type PluginProvides struct {
	Skills   []string `json:"skills,omitempty"`
	Agents   []string `json:"agents,omitempty"`
	Commands []string `json:"commands,omitempty"`
	Hooks    []string `json:"hooks,omitempty"`
}

// Plugin represents a plugin entry (LEGACY)
type Plugin struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Version     string          `json:"version,omitempty"`
	Category    string          `json:"category,omitempty"`
	Source      PluginSource    `json:"source"`
	Author      *Author         `json:"author,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	Provides    *PluginProvides `json:"provides,omitempty"`
}

// Author represents a plugin author (LEGACY)
type Author struct {
	Name string `json:"name"`
}

// UnmarshalJSON allows Author to be unmarshaled from either a string or an object
func (a *Author) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		a.Name = name
		return nil
	}

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

// RegistryEntry represents a plugin in the registry (LEGACY)
type RegistryEntry struct {
	Name           string          `json:"name"`
	Description    string          `json:"description,omitempty"`
	Version        string          `json:"version,omitempty"`
	Category       string          `json:"category,omitempty"`
	Author         *Author         `json:"author,omitempty"`
	Tags           []string        `json:"tags,omitempty"`
	Source         PluginSource    `json:"source"`
	ResolvedSource interface{}     `json:"resolvedSource"`
	InstalledAt    time.Time       `json:"installedAt"`
	Provides       *PluginProvides `json:"provides,omitempty"`
}
