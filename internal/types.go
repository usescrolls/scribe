package scribe

// ProgressEvent represents a progress update during discovery or installation
type ProgressEvent struct {
	Phase   string `json:"phase"`            // "discover" or "install"
	Step    string `json:"step"`             // e.g. "cloning", "scanning", "copying"
	Message string `json:"message"`          // human-readable message
	Detail  string `json:"detail,omitempty"` // e.g. agent name
}

// ProgressEmitter is a callback for emitting progress events
type ProgressEmitter func(ProgressEvent)

// emitProgress is a nil-safe helper for emitting progress events from a variadic slice
func emitProgress(emit []ProgressEmitter, phase, step, msg, detail string) {
	if len(emit) > 0 && emit[0] != nil {
		emit[0](ProgressEvent{Phase: phase, Step: step, Message: msg, Detail: detail})
	}
}

// Skill represents a skill with its SKILL.md content and metadata
type Skill struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Path        string         `json:"path,omitempty"`     // Local path to SKILL.md directory
	Content     string         `json:"content,omitempty"`  // Raw SKILL.md content
	Metadata    map[string]any `json:"metadata,omitempty"` // Additional frontmatter fields
	Meta        *SkillMeta     `json:"meta,omitempty"`     // Source tracking (from .scribe-meta.json)
}

// GitCommitInfo holds the short hash and date of a git commit
type GitCommitInfo struct {
	Hash string // short hash, 7 chars
	Date string // ISO 8601
}

// SkillMeta tracks the source and installation info for a skill
// Stored in .scribe-meta.json sidecar file alongside SKILL.md
type SkillMeta struct {
	Source      string `json:"source"`               // e.g., "owner/repo"
	SourceType  string `json:"sourceType"`           // github, gitlab, bitbucket, local, url, well-known
	SourceURL   string `json:"sourceUrl,omitempty"`  // Full URL if applicable
	SkillPath   string `json:"skillPath,omitempty"`  // Path within source repo
	ContentHash string `json:"contentHash"`          // SHA256 hash of SKILL.md content
	CommitHash  string `json:"commitHash,omitempty"` // Short git commit hash
	CommitDate  string `json:"commitDate,omitempty"` // ISO 8601 commit timestamp
	IsPrivate   bool   `json:"isPrivate,omitempty"`  // True if authentication was used to fetch
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

// CurrentTermsVersion is bumped whenever the terms and conditions text changes.
// When this is higher than the user's TermsAcceptedVersion, they must re-accept.
const CurrentTermsVersion = 1

// TermsClause is a single clause in the terms and conditions.
type TermsClause struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// TermsClauses is the single source of truth for the terms and conditions text.
// Both the CLI and frontend read from this. Bump CurrentTermsVersion when changing.
var TermsClauses = []TermsClause{
	{
		Title: "Skill Management",
		Body:  "Scribe manages files in your coding agents' configuration directories (e.g., ~/.claude/skills, ~/.cursor/skills). By using Scribe, you authorize it to create, update, and remove skill files in these directories on your behalf.",
	},
	{
		Title: "Community Content",
		Body:  "Skills available through Scribe may be created by third-party contributors. Scribe does not verify, audit, or endorse the content of any skill. You are responsible for reviewing skills before installing them.",
	},
	{
		Title: "No Warranty",
		Body:  "Scribe is provided \"as is\" without warranty of any kind. The authors and contributors are not liable for any damages or issues arising from the use of this tool or any installed skills.",
	},
	{
		Title: "Use at Your Own Risk",
		Body:  "You accept full responsibility for any consequences of using Scribe, including any changes made to your system or coding agent configurations.",
	},
}

// Config represents the global Scribe configuration
type Config struct {
	ActiveWorkspace             string `json:"activeWorkspace"`
	OnboardingCompleted         bool   `json:"onboardingCompleted"`
	TermsAcceptedAt             string `json:"termsAcceptedAt,omitempty"`      // RFC3339 timestamp when terms were accepted
	TermsAcceptedVersion        int    `json:"termsAcceptedVersion,omitempty"` // Version of terms the user accepted
	UpdateNotificationsDisabled bool   `json:"updateNotificationsDisabled,omitempty"`
	LastUpdateCheck             string `json:"lastUpdateCheck,omitempty"` // RFC3339 timestamp
}

// InstallOptions configures skill installation behavior
type InstallOptions struct {
	Skills    []string // Select specific skills to install (empty = all found)
	Yes       bool     // Skip interactive prompts
	ListOnly  bool     // List available skills without installing
	IsPrivate bool     // Source required authentication to fetch
}

// SourceInfo represents a parsed source reference
type SourceInfo struct {
	Type      string // github, gitlab, bitbucket, git, local, url, well-known, zip
	Owner     string // GitHub/GitLab owner
	Repo      string // Repository name
	Ref       string // Branch, tag, or commit
	Subpath   string // Path within repo
	URL       string // Full URL for url/well-known types
	LocalPath string // Absolute path for local sources
}

// SkillInfo is the frontend-friendly representation of a skill
type SkillInfo struct {
	Name        string   `json:"name"`                  // Storage/directory name (possibly source-qualified)
	DisplayName string   `json:"displayName,omitempty"` // Frontmatter name (always simple, e.g. "commit")
	Description string   `json:"description"`
	Source      string   `json:"source"`              // Human-readable source
	SourceType  string   `json:"sourceType"`          // github, local, url, etc.
	SourceURL   string   `json:"sourceUrl,omitempty"` // Full URL to source repo/page
	InstalledAt string   `json:"installedAt"`         // ISO formatted timestamp
	UpdatedAt   string   `json:"updatedAt,omitempty"`
	ContentHash string   `json:"contentHash,omitempty"`
	CommitHash  string   `json:"commitHash,omitempty"`
	CommitDate  string   `json:"commitDate,omitempty"`
	IsPrivate   bool     `json:"isPrivate,omitempty"` // True if source required authentication
	IsSystem    bool     `json:"isSystem,omitempty"`  // True for non-removable system skills
	Agents      []string `json:"agents"`              // List of agent IDs with this skill
}

// CheckResult represents the result of checking a skill for updates
type CheckResult struct {
	Name        string `json:"name"`
	NeedsUpdate bool   `json:"needsUpdate"`
	Error       string `json:"error,omitempty"`
	CurrentHash string `json:"currentHash,omitempty"`
	RemoteHash  string `json:"remoteHash,omitempty"`
}

// SourceGroupCheckResult summarizes an update check for all skills from a single source.
type SourceGroupCheckResult struct {
	Source            string   `json:"source"`
	HasUpdates        bool     `json:"hasUpdates"`
	UpdatedSkillNames []string `json:"updatedSkillNames"`
	CheckedAt         string   `json:"checkedAt"`
	Error             string   `json:"error,omitempty"`
}

// UpdateResult contains the outcome of a skill update operation
type UpdateResult struct {
	SkillName  string `json:"skillName"`
	Updated    bool   `json:"updated"`
	Removed    bool   `json:"removed,omitempty"`    // True if skill was removed (no longer in source)
	OldHash    string `json:"oldHash,omitempty"`    // Previous commit hash or truncated content hash
	NewHash    string `json:"newHash,omitempty"`    // New commit hash or truncated content hash
	CommitDate string `json:"commitDate,omitempty"` // New commit date if available
}

// UpdateInfo contains information about an available application update
type UpdateInfo struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
	ReleaseURL      string `json:"releaseURL"`
	PublishedAt     string `json:"publishedAt"`
}

// SelfUpdateResult contains the outcome of a self-update (binary upgrade) attempt.
type SelfUpdateResult struct {
	Updated       bool   `json:"updated"`
	OldVersion    string `json:"oldVersion"`
	NewVersion    string `json:"newVersion"`
	InstallMethod string `json:"installMethod"`
	Message       string `json:"message"`
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

// ExistingSkillInfo represents a skill found in an agent's directory during onboarding
type ExistingSkillInfo struct {
	Name      string `json:"name"`
	Path      string `json:"path"`      // Full path to the skill directory
	AgentID   string `json:"agentId"`   // Which agent it was found in
	AgentName string `json:"agentName"` // Display name of the agent
	IsGitRepo bool   `json:"isGitRepo"` // Whether the skill directory is a git repo
}

// SkillConflict represents a naming conflict between skills from different agents
type SkillConflict struct {
	Name    string              `json:"name"`
	Sources []ExistingSkillInfo `json:"sources"`
}

// DiscoverResult is the frontend-friendly result of discovering skills from a source
type DiscoverResult struct {
	Skills     []DiscoveredSkill `json:"skills"`
	Source     string            `json:"source"`
	SourceType string            `json:"sourceType"`
}

// DiscoveredSkill represents a skill found during source discovery (before install)
type DiscoveredSkill struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	AlreadyInstalled bool   `json:"alreadyInstalled"`
}
