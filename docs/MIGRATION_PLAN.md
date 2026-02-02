# Scribe Architecture Migration Plan

## Implementation Progress

### Completed (MVP Backend) ✅

| Task | File | Status |
|------|------|--------|
| New types (Skill, SkillMeta, Agent, Workspace, Config, SourceInfo) | `internal/types.go` | ✅ Done |
| 45 coding agent definitions with detection | `internal/agents.go` | ✅ Done |
| Canonical storage paths (~/.scribe/scrolls/) | `internal/storage.go` | ✅ Done |
| SKILL.md parsing with YAML frontmatter | `internal/skills.go` | ✅ Done |
| Sidecar .scribe-meta.json management | `internal/meta.go` | ✅ Done |
| Symlink-based installation to agents | `internal/installer.go` | ✅ Done |
| Workspace system (CRUD, switching) | `internal/workspace.go` | ✅ Done |
| CLI install command (GitHub, local, GitLab) | `cli/install.go` | ✅ Done |
| CLI uninstall command | `cli/uninstall.go` | ✅ Done |
| CLI workspace commands | `cli/workspace.go` | ✅ Done |
| Updated tests for new system | `cli/cli_test.go` | ✅ Done |
| **Global-only simplification** | All files | ✅ Done |
| Backend unit tests (72.5% coverage) | `internal/skills_system_test.go` | ✅ Done |
| Docker test infrastructure | `test.Dockerfile`, `docker-compose.test.yml` | ✅ Done |
| CI workflow updated to Go 1.25 | `.github/workflows/release.yml` | ✅ Done |

**Note:** The architecture has been simplified to global-only skills. There are no per-project skills. All coding agents are managed uniformly by Scribe.

### Completed (MVP Frontend) ✅

| Task | File | Status |
|------|------|--------|
| Wails bindings for skills API | `main.go` | ✅ Done |
| Wails bindings for workspaces API | `main.go` | ✅ Done |
| Wails bindings for agents API | `main.go` | ✅ Done |
| TypeScript types (SkillInfo, WorkspaceInfo, AgentStatus) | `frontend/src/types/skill.ts` | ✅ Done |
| useSkills.ts composable | `frontend/src/composables/useSkills.ts` | ✅ Done |
| useWorkspaces.ts composable | `frontend/src/composables/useWorkspaces.ts` | ✅ Done |
| useAgents.ts composable | `frontend/src/composables/useAgents.ts` | ✅ Done |
| SkillList.vue component | `frontend/src/components/SkillList.vue` | ✅ Done |
| SkillCard.vue component | `frontend/src/components/SkillCard.vue` | ✅ Done |
| WorkspaceSelector.vue component | `frontend/src/components/WorkspaceSelector.vue` | ✅ Done |
| AgentStatusPanel.vue component | `frontend/src/components/AgentStatusPanel.vue` | ✅ Done |
| Updated App.vue with sidebar layout | `frontend/src/App.vue` | ✅ Done |
| Updated EmptyState.vue for skills | `frontend/src/components/EmptyState.vue` | ✅ Done |
| System tray shows skills count | `main.go` | ✅ Done |
| System tray shows active workspace | `main.go` | ✅ Done |
| **Frontend unit tests (Vitest)** | `frontend/src/**/*.test.ts` | ✅ Done |

**Test Coverage:** 63 tests across 6 test files covering composables and components.

### Completed (CLI Migration) ✅

| Task | File | Status |
|------|------|--------|
| Zip URL download & extraction | `cli/install.go` | ✅ Done |
| `scribe check` - Check for updates | `cli/check.go` | ✅ Done |
| `scribe update` - Update skills | `cli/update.go` | ✅ Done |
| Update `scribe list` for skills | `cli/list.go` | ✅ Done |
| Update `scribe info` for skills | `cli/info.go` | ✅ Done |
| Export `GetAgentsWithSkill` | `internal/skills.go` | ✅ Done |
| Update CLI tests for skills | `cli/cli_test.go` | ✅ Done |

### Completed (Legacy Cleanup) ✅

| Task | File | Status |
|------|------|--------|
| Remove internal/marketplace.go | Deleted | ✅ Done |
| Remove internal/claude.go | Deleted | ✅ Done |
| Remove internal/plugins.go | Deleted | ✅ Done |
| Remove internal/registry.go | Deleted | ✅ Done |
| Remove internal/url_scheme.go | Deleted | ✅ Done |
| Remove internal/source.go | Deleted | ✅ Done |
| Remove internal/server.go | Deleted | ✅ Done |
| Remove legacy types from types.go | `internal/types.go` | ✅ Done |
| Remove old Plugin* Vue components | Deleted PluginCard.vue, PluginList.vue | ✅ Done |
| Remove legacy composables/types | Deleted usePlugins.ts, plugin.ts | ✅ Done |
| Remove legacy tests | Deleted scribe_test.go | ✅ Done |
| Update main.go to skills-only | `main.go` | ✅ Done |
| Update cli/root.go | `cli/root.go` | ✅ Done |
| Update cli/cli_test.go | `cli/cli_test.go` | ✅ Done |

### Pending Work 🚧

| Task | Priority | Notes |
|------|----------|-------|
| **Source Types** | | |
| Well-known endpoint (/.well-known/skills/) | Low | Not yet implemented |
| Direct URL (single SKILL.md) | Low | Not yet implemented |

---

## Executive Summary

**Goal:** Transform Scribe from a multi-purpose plugin manager into a focused skills-only manager with workspace support for 40+ coding agents.

### Key Changes
- **Skills only** - Drop support for plugins containing agents, hooks, or commands
- **Canonical storage** - All skills stored in `~/.scribe/scrolls/<skill-name>/`
- **Symlink distribution** - Skills symlinked to each agent's skills directory
- **Workspaces** - Switch between skill sets globally with a single command
- **40+ agents** - Support all major coding agents (Claude Code, Cursor, Copilot, etc.)
- **Clean break** - No migration from existing `~/.scribe` data
- **Updated GUI** - Vue 3 frontend with workspace selector, agent status panel, and skill cards
- **Agent visualization** - See all 40+ supported agents, which are installed, and skill counts per agent

### Key Decisions
| Decision | Choice |
|----------|--------|
| Migration approach | Clean break (no migration) |
| Skill scope | **Global only** - no per-project skills |
| Agent management | **Unified** - all agents managed globally by Scribe |
| Auto-add to workspace | Yes |
| Agent selection | All detected agents |
| URL scheme | Keep `agenthub://` |
| GUI/System tray | Keep and update |
| Skill tracking | Sidecar `.scribe-meta.json` per skill (no central lock file) |

---

## Overview

Migrate Scribe from a multi-purpose plugin manager (skills, agents, hooks, commands) to a skills-only manager with workspace support, inspired by Vercel's Skills CLI architecture.

---

## Current State → Target State

| Aspect | Current (Scribe) | Target (New Scribe) |
|--------|------------------|---------------------|
| **Scope** | Plugins (skills, agents, hooks, commands) | Skills only |
| **Canonical Storage** | `~/.scribe/plugins/<name>/` | `~/.scribe/scrolls/<name>/` |
| **Metadata** | Central `registry.json` | Sidecar `.scribe-meta.json` per skill |
| **Agent Integration** | Claude Code only (via marketplace.json) | 40+ agents via symlinks |
| **Installation** | Downloads to local, writes marketplace.json | Symlinks to agent directories |
| **Workspaces** | None | Global workspace switching |

---

## Architecture Components

### 1. Canonical Storage (`~/.scribe/scrolls/`)

All skills are stored globally in a single canonical location. Scribe manages all coding agents uniformly - there are no per-project skills.

```
~/.scribe/
├── scrolls/                    # Canonical skill storage (global only)
│   ├── react-best-practices/
│   │   ├── SKILL.md
│   │   └── .scribe-meta.json   # Source tracking (sidecar)
│   ├── typescript-patterns/
│   │   ├── SKILL.md
│   │   └── .scribe-meta.json
│   └── ...
├── workspaces/                 # Workspace definitions
│   ├── default.json            # Default workspace (all skills)
│   ├── web-dev.json
│   └── backend.json
└── config.json                 # Global config (active workspace)
```

**Note:** All skills are global. Workspaces control which global skills are active across all agents.

### 2. Skill Format (SKILL.md)

```markdown
---
name: my-skill-name
description: Brief description of what this skill does
---

# My Skill Name

Detailed instructions for AI agents...
```

**Frontmatter fields:**
- `name` (required): Unique skill identifier (kebab-case)
- `description` (required): Brief description for listings
- `metadata.internal` (optional): Hide from discovery

### 3. Supported Agents (40+)

Skills are symlinked to each agent's global skills directory. Scribe manages all agents uniformly.

| Agent | Global Skills Dir |
|-------|------------------|
| Claude Code | `~/.claude/skills/` |
| Cursor | `~/.cursor/skills/` |
| GitHub Copilot | `~/.copilot/skills/` |
| Cline | `~/.cline/skills/` |
| Continue | `~/.continue/skills/` |
| Windsurf | `~/.codeium/windsurf/skills/` |
| OpenCode | `~/.config/opencode/skills/` |
| Codex | `~/.codex/skills/` |
| Gemini CLI | `~/.gemini/skills/` |
| Goose | `~/.config/goose/skills/` |
| + 35 more... | |

**Key rule**: Only create symlinks to agent directories that already exist (agent is installed).

### 4. Workspace System

Workspaces define which skills are active globally. When you switch workspaces, Scribe:
1. Removes symlinks from all agent directories for skills NOT in the target workspace
2. Creates symlinks for skills IN the target workspace to all detected agents

```json
// ~/.scribe/workspaces/default.json
{
  "name": "default",
  "description": "All installed skills",
  "skills": ["react-best-practices", "typescript-patterns", "go-patterns", "api-design"]
}

// ~/.scribe/workspaces/web-dev.json
{
  "name": "web-dev",
  "description": "Web development skills only",
  "skills": ["react-best-practices", "typescript-patterns"]
}

// ~/.scribe/workspaces/backend.json
{
  "name": "backend",
  "description": "Backend development skills",
  "skills": ["go-patterns", "api-design"]
}
```

**Example workspace switch:**
```bash
$ scribe workspace use backend
Switching from 'web-dev' to 'backend'...
✗ Removed: react-best-practices
✗ Removed: typescript-patterns
✓ Added: go-patterns
✓ Added: api-design
Active workspace: backend (2 skills)
```

**Note:** Skills installed while workspace X is active are automatically added to workspace X AND the "default" workspace (which always contains all skills).

### 5. Sidecar Metadata (`.scribe-meta.json`)

Each skill folder contains its own metadata file - no central lock file needed:

```
~/.scribe/scrolls/
├── react-best-practices/
│   ├── SKILL.md                # The skill content
│   └── .scribe-meta.json       # Source and tracking info
├── typescript-patterns/
│   ├── SKILL.md
│   └── .scribe-meta.json
└── ...
```

**Sidecar file format:**
```json
// ~/.scribe/scrolls/react-best-practices/.scribe-meta.json
{
  "source": "vercel-labs/agent-skills",
  "sourceType": "github",
  "sourceUrl": "https://github.com/vercel-labs/agent-skills",
  "skillPath": "skills/react-best-practices",
  "contentHash": "sha256:abc123...",
  "installedAt": "2025-01-29T10:30:00Z",
  "updatedAt": "2025-01-29T10:30:00Z"
}
```

**Benefits:**
- Self-contained: moving a skill folder moves its metadata too
- No central lock file to corrupt or conflict
- Easy to inspect per-skill
- Skills without metadata are treated as manually added

### 6. Config File (`~/.scribe/config.json`)

Global configuration only:

```json
{
  "activeWorkspace": "default",
  "preferences": {
    "defaultScope": "global"
  }
}
```

---

## CLI Commands

### Core Commands ✅ IMPLEMENTED

```bash
# Install skills from various sources
scribe install owner/repo                    # GitHub shorthand
scribe install https://github.com/owner/repo # Full GitHub URL
scribe install ./local/path                  # Local directory
scribe install https://example.com           # Well-known endpoint (pending)

# List installed skills
scribe list                                  # (currently shows legacy plugins)
scribe ls

# Remove skills
scribe uninstall <skill-name>
scribe remove <skill-name>
scribe rm <skill-name>

# Show skill info
scribe info <skill-name>                     # (currently shows legacy plugins)

# Check for updates (pending)
scribe check

# Update all skills (pending)
scribe update
```

### Workspace Commands

```bash
# List workspaces
scribe workspace list

# Create a new workspace
scribe workspace create <name>

# Switch to a workspace
scribe workspace use <name>

# Add skill to workspace
scribe workspace add <skill-name>

# Remove skill from workspace
scribe workspace remove <skill-name>

# Show current workspace
scribe workspace current
```

### Installation Options ✅ IMPLEMENTED

```bash
scribe install owner/repo [flags]

Flags:
  -a, --agent <agents>   Target specific agents (comma-separated)
  -s, --skill <skills>   Select specific skills to install
  -l, --list             List available skills without installing
  -y, --yes              Skip interactive prompts
  --all                  Install all skills to all detected agents
```

**Note:** All skills are installed globally. There is no project-level installation.

---

## Installation Flow

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Parse Source                                              │
│    URL/path/shorthand → structured source object            │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. Fetch Content                                             │
│    Git clone OR HTTP fetch OR read local path               │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. Discover Skills                                           │
│    Find folders with valid SKILL.md files                   │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. Detect Installed Agents                                   │
│    Check for existing agent config directories              │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│ 5. User Selection (Interactive)                              │
│    - Select skills to install                               │
│    - Select agents to target                                │
│    - Choose scope (project/global)                          │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│ 6. Install Skills                                            │
│    - Copy to ~/.scribe/scrolls/<skill>/                     │
│    - Create symlinks to agent directories                   │
│    - Fallback to copy if symlinks fail                      │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│ 7. Write Metadata & Update Workspace                         │
│    - Write .scribe-meta.json in skill folder                │
│    - Add skill to active workspace                          │
└─────────────────────────────────────────────────────────────┘
```

---

## Source Types

| Type | Example | Handler |
|------|---------|---------|
| GitHub Shorthand | `owner/repo` | Clone repo, discover skills |
| GitHub URL | `https://github.com/owner/repo` | Clone repo, discover skills |
| GitHub Subpath | `owner/repo/path/to/skills` | Clone repo, use subpath |
| GitHub Branch | `owner/repo#branch` | Clone specific branch |
| GitLab | `https://gitlab.com/owner/repo` | Clone repo |
| Local Path | `./local/path` | Read directly |
| Direct URL | `https://example.com/skill.md` | HTTP fetch |
| Well-Known | `https://example.com` | Fetch `/.well-known/skills/` |
| **Zip URL** | `https://example.com/plugin.zip` | Download, extract, discover skills |
| **Old Plugin Zip** | (legacy format) | Extract skills only, ignore commands/agents |

---

## Skill Discovery

When fetching from a source, discover skills by:

1. Look for `SKILL.md` in root (single-skill repo)
2. Search common directories:
   - `skills/`
   - `.claude/skills/`, `.cursor/skills/`, etc.
3. Recursive search (max depth 5, exclude: `node_modules`, `.git`, `dist`, `build`)

**Valid skill**: A directory containing a `SKILL.md` file with valid frontmatter (name + description).

---

## Handling Old Plugin Format (Zip Sources)

Old plugin zips from the existing registry have a different structure:

```
old-plugin.zip/
├── .claude-plugin/
│   └── plugin.json       ← IGNORED (old metadata format)
├── commands/             ← IGNORED (not supported in new system)
├── agents/               ← IGNORED (not supported in new system)
├── hooks/                ← IGNORED (not supported in new system)
└── skills/               ← SEARCHED for SKILL.md files
    ├── skill-one/
    │   └── SKILL.md      ✓ Found and installed
    └── skill-two/
        └── SKILL.md      ✓ Found and installed
```

### How It Works

1. **Skill discovery searches recursively** - The `skills/` subdirectory is in the search path, so SKILL.md files are found automatically
2. **Non-skill content is ignored** - `commands/`, `agents/`, `hooks/`, `plugin.json` are simply not processed
3. **No explicit migration needed** - Old zips "just work" for their skill content

### Edge Cases

| Scenario | Behavior |
|----------|----------|
| Old zip with skills | Skills extracted, rest ignored |
| Old zip with NO skills (commands-only) | Warning: "No skills found in source" |
| Old zip with `plugin.json` but no `SKILL.md` | Warning: "Legacy plugin format - no skills found" |
| Mixed: some skills have `SKILL.md`, some don't | Only valid skills installed |

### Detection Logic

```go
func discoverSkills(root string) ([]*Skill, error) {
    var skills []*Skill

    // Check for legacy plugin.json (for warning purposes only)
    if exists(filepath.Join(root, ".claude-plugin", "plugin.json")) {
        log.Debug("Legacy plugin format detected, extracting skills only")
    }

    // Standard recursive SKILL.md discovery
    // This naturally finds skills/ subdirectory content
    filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        if d.Name() == "SKILL.md" {
            skill, err := parseSkillMd(path)
            if err == nil && skill != nil {
                skills = append(skills, skill)
            }
        }
        return nil
    })

    if len(skills) == 0 {
        return nil, fmt.Errorf("no skills found in source (may be legacy commands-only plugin)")
    }

    return skills, nil
}
```

### Test Cases for Old Format

```go
func TestDiscoverSkills_LegacyPluginWithSkills(t *testing.T) {
    // Create old-style plugin structure with skills/ dir
    // Verify skills are found
}

func TestDiscoverSkills_LegacyPluginNoSkills(t *testing.T) {
    // Create old-style plugin with only commands/
    // Verify appropriate error/warning
}

func TestDiscoverSkills_MixedContent(t *testing.T) {
    // Plugin with both plugin.json and SKILL.md files
    // Verify only skills extracted
}
```

---

## Agent Detection

Only create symlinks to agents that are already installed:

```go
func detectInstalledAgents() []Agent {
    var installed []Agent
    for _, agent := range allAgents {
        // Check if agent's global config directory exists
        if exists(agent.GlobalConfigDir) {
            installed = append(installed, agent)
        }
    }
    return installed
}
```

---

## Migration Path

### Phase 1: Core Infrastructure (Types & Storage)

**New files to create:**
- `internal/types.go` - Rewrite with new types:
  ```go
  type Skill struct {
      Name        string
      Description string
      Path        string            // Local path to SKILL.md directory
      Content     string            // Raw SKILL.md content
      Metadata    map[string]any    // Additional frontmatter fields
      Meta        *SkillMeta        // Source tracking (from .scribe-meta.json)
  }

  type SkillMeta struct {
      Source      string `json:"source"`       // e.g., "owner/repo"
      SourceType  string `json:"sourceType"`   // github, local, url, well-known
      SourceURL   string `json:"sourceUrl"`
      SkillPath   string `json:"skillPath,omitempty"`
      ContentHash string `json:"contentHash"`
      InstalledAt string `json:"installedAt"`
      UpdatedAt   string `json:"updatedAt"`
  }

  // Agent - all agents managed globally by Scribe
  type Agent struct {
      ID              string
      DisplayName     string
      GlobalSkillsDir string        // e.g., "~/.claude/skills"
      GlobalConfigDir string        // For detection, e.g., "~/.claude"
  }

  type Workspace struct {
      Name        string   `json:"name"`
      Description string   `json:"description"`
      Skills      []string `json:"skills"`  // Skill names in this workspace
  }

  type Config struct {
      ActiveWorkspace string `json:"activeWorkspace"`
  }
  ```

- `internal/agents.go` - Agent registry (40+ agents from Vercel's skills CLI):
  - Define all agents with their directories
  - `DetectInstalledAgents()` - Check which agents have config directories
  - `GetAgent(id string)` - Get agent by ID

- `internal/storage.go` - Canonical storage management:
  - `GetScrollsDir()` → `~/.scribe/scrolls/` (global only)
  - `GetWorkspacesDir()` → `~/.scribe/workspaces/`
  - `GetConfigPath()` → `~/.scribe/config.json`

### Phase 2: Skill Discovery & Parsing

**New files to create:**
- `internal/skills.go`:
  - `ParseSkillMd(path string) (*Skill, error)` - Parse frontmatter + content
  - `DiscoverSkills(dir string) ([]*Skill, error)` - Find all SKILL.md files
  - `ValidateSkill(skill *Skill) error` - Ensure required fields exist
  - `SanitizeName(name string) string` - Path-safe skill names

**Update files:**
- `internal/source.go` - Add source types:
  - GitHub shorthand: `owner/repo`, `owner/repo#branch`, `owner/repo/path`
  - GitHub URL: `https://github.com/owner/repo`
  - GitLab URL: `https://gitlab.com/owner/repo`
  - Local path: `./path`, `/absolute/path`
  - Direct URL: `https://example.com/SKILL.md`
  - Well-known: `https://example.com` → fetch `/.well-known/skills/`

### Phase 3: Installation System

**New files to create:**
- `internal/installer.go`:
  - `InstallSkill(skill *Skill, opts InstallOptions) error`
  - `CreateSymlink(target, link string) error` - With Windows junction support
  - `CopyDirectory(src, dst string) error` - Fallback for symlink failures
  - `SyncSkillToAgents(skillName string, agents []Agent) error`
  - `RemoveSkillFromAgents(skillName string, agents []Agent) error`

- `internal/meta.go` - Sidecar metadata management:
  - `ReadSkillMeta(skillPath string) (*SkillMeta, error)`
  - `WriteSkillMeta(skillPath string, meta *SkillMeta) error`
  - `ComputeContentHash(content string) string`
  - `ListSkillsWithMeta(scrollsDir string) ([]*Skill, error)` - Read all skills with their metadata

### Phase 4: Workspace System

**New files to create:**
- `internal/workspace.go`:
  - `ListWorkspaces() ([]*Workspace, error)`
  - `GetWorkspace(name string) (*Workspace, error)`
  - `CreateWorkspace(ws *Workspace) error`
  - `UpdateWorkspace(ws *Workspace) error`
  - `DeleteWorkspace(name string) error`
  - `GetActiveWorkspace() (*Workspace, error)`
  - `SetActiveWorkspace(name string) error`
  - `AddSkillToWorkspace(skillName, workspaceName string) error`
  - `RemoveSkillFromWorkspace(skillName, workspaceName string) error`
  - `SyncWorkspace(ws *Workspace) error` - Create/remove symlinks to match workspace

**Workspace switching flow:**
```
1. Load target workspace definition
2. Load current active workspace
3. Determine skills to remove (in current but not target)
4. Determine skills to add (in target but not current)
5. Remove symlinks for skills to remove
6. Create symlinks for skills to add
7. Update config.json with new active workspace
```

### Phase 5: CLI Commands

**Rename and update:**
- `cli/install.go` → `cli/add.go`:
  - Parse source (GitHub, local, URL, etc.)
  - Fetch content
  - Discover skills
  - Detect agents
  - Install to canonical location
  - Sync symlinks to all detected agents
  - Add to lock file
  - Add to active workspace

- `cli/uninstall.go` → `cli/remove.go`:
  - Remove from lock file
  - Remove from all workspaces
  - Remove symlinks from all agents
  - Optionally delete canonical copy

**Update:**
- `cli/list.go` - List installed skills with workspace info
- `cli/info.go` - Show skill details, source, agents

**New:**
- `cli/workspace.go`:
  - `scribe workspace list`
  - `scribe workspace create <name>`
  - `scribe workspace use <name>`
  - `scribe workspace add <skill-name>`
  - `scribe workspace remove <skill-name>`
  - `scribe workspace current`
  - `scribe workspace delete <name>`

- `cli/check.go` - Check for skill updates
- `cli/update.go` - Update all skills

### Phase 6: Frontend (Vue 3 / Wails)

The desktop app has a Vue 3 frontend that needs to be updated for the skills-only architecture.

**Current frontend structure:**
```
frontend/src/
├── App.vue                     # Main layout
├── components/
│   ├── PluginList.vue          # Lists plugins → rename to SkillList
│   ├── PluginCard.vue          # Plugin display → rename to SkillCard
│   └── EmptyState.vue          # No plugins state
├── composables/
│   └── usePlugins.ts           # Plugin data → rename to useSkills
├── types/
│   └── plugin.ts               # PluginInfo type → rename to skill.ts
└── bindings/
    └── scribe.ts               # Wails bindings (auto-generated)
```

**Files to rename/update:**

| Current | New | Changes |
|---------|-----|---------|
| `PluginList.vue` | `SkillList.vue` | Rename, add workspace selector |
| `PluginCard.vue` | `SkillCard.vue` | Show agents, remove version/author |
| `usePlugins.ts` | `useSkills.ts` | Update API calls |
| `types/plugin.ts` | `types/skill.ts` | New SkillInfo interface |
| - | `WorkspaceSelector.vue` | **NEW** - Workspace dropdown |
| - | `useWorkspaces.ts` | **NEW** - Workspace state |

**New TypeScript types (`types/skill.ts`):**
```typescript
export interface SkillInfo {
  name: string
  description: string
  source: string
  sourceType: string
  installedAt: string
  agents: string[]           // List of agent IDs with this skill
}

export interface WorkspaceInfo {
  name: string
  description: string
  skills: string[]           // Skill names in this workspace
  isActive: boolean
}
```

**New components:**

1. **WorkspaceSelector.vue** - Dropdown in header
   ```vue
   <template>
     <select v-model="activeWorkspace" @change="switchWorkspace">
       <option v-for="ws in workspaces" :value="ws.name">
         {{ ws.name }} ({{ ws.skills.length }} skills)
       </option>
     </select>
   </template>
   ```

2. **AgentStatusPanel.vue** - Visual display of installed coding agents
   ```vue
   <template>
     <div class="agent-status-panel">
       <div class="panel-header">
         <h3>Coding Agents</h3>
         <span class="summary">{{ installedCount }}/{{ totalCount }} installed</span>
       </div>
       <div class="agent-grid">
         <div
           v-for="agent in agents"
           :key="agent.id"
           :class="['agent-item', { installed: agent.installed, selected: agent.selected }]"
           @click="toggleAgent(agent)"
         >
           <div class="agent-icon">
             <span v-if="agent.installed" class="checkmark">✓</span>
             <span v-else class="empty">○</span>
           </div>
           <div class="agent-info">
             <span class="agent-name">{{ agent.displayName }}</span>
             <span v-if="agent.installed" class="skill-count">
               {{ agent.skillCount }} skills
             </span>
             <span v-else class="not-installed">Not installed</span>
           </div>
         </div>
       </div>
     </div>
   </template>
   ```

   **Features:**
   - Grid/list view of all 40+ supported agents
   - Visual checkmark indicator for installed agents
   - Grayed out styling for non-installed agents
   - Shows skill count per installed agent
   - Click to select/filter skills by agent
   - Collapsible panel to save space
   - Search/filter agents by name

3. **SkillCard.vue** - Updated design
   ```vue
   <template>
     <div class="skill-card">
       <h3>{{ skill.name }}</h3>
       <p>{{ skill.description }}</p>
       <div class="agents">
         <span v-for="agent in skill.agents" class="agent-badge">
           {{ agentDisplayName(agent) }}
         </span>
       </div>
       <div class="source">{{ skill.sourceType }}: {{ skill.source }}</div>
       <button @click="$emit('uninstall', skill.name)">Remove</button>
     </div>
   </template>
   ```

4. **App.vue** - Updated layout with workspace selector and agent panel
   ```vue
   <template>
     <div class="app">
       <header class="header">
         <h1>Scribe</h1>
         <WorkspaceSelector />
         <span class="version">v{{ version }}</span>
       </header>
       <div class="content">
         <aside class="sidebar">
           <AgentStatusPanel @agent-selected="filterByAgent" />
         </aside>
         <main class="main">
           <SkillList :agent-filter="selectedAgent" />
         </main>
       </div>
     </div>
   </template>
   ```

   **Layout:** Sidebar with agent status panel, main area with skill list. Agent panel can be collapsed on smaller screens.

**New composables:**

1. **useSkills.ts**
   ```typescript
   export function useSkills() {
     const skills = ref<SkillInfo[]>([])

     async function fetchSkills() {
       skills.value = await AppService.GetSkills()
     }

     async function uninstall(name: string) {
       await AppService.RemoveSkill(name)
       await fetchSkills()
     }

     return { skills, fetchSkills, uninstall }
   }
   ```

2. **useWorkspaces.ts**
   ```typescript
   export function useWorkspaces() {
     const workspaces = ref<WorkspaceInfo[]>([])
     const activeWorkspace = ref<string>('default')

     async function fetchWorkspaces() {
       workspaces.value = await AppService.GetWorkspaces()
       activeWorkspace.value = await AppService.GetActiveWorkspace()
     }

     async function switchWorkspace(name: string) {
       await AppService.SetActiveWorkspace(name)
       await fetchWorkspaces()
       Events.Emit('workspace-changed', name)
     }

     return { workspaces, activeWorkspace, fetchWorkspaces, switchWorkspace }
   }
   ```

3. **useAgents.ts** - Agent detection and status
   ```typescript
   export interface AgentStatus {
     id: string
     displayName: string
     installed: boolean
     skillCount: number
     globalSkillsDir: string
   }

   export function useAgents() {
     const agents = ref<AgentStatus[]>([])
     const selectedAgent = ref<string | null>(null)

     async function fetchAgents() {
       agents.value = await AppService.GetAgentStatus()
     }

     function selectAgent(agentId: string | null) {
       selectedAgent.value = agentId
       Events.Emit('agent-filter-changed', agentId)
     }

     const installedAgents = computed(() =>
       agents.value.filter(a => a.installed)
     )

     const installedCount = computed(() => installedAgents.value.length)
     const totalCount = computed(() => agents.value.length)

     return {
       agents,
       selectedAgent,
       installedAgents,
       installedCount,
       totalCount,
       fetchAgents,
       selectAgent
     }
   }
   ```

**Backend bindings required (Go → Wails):**

Update `internal/server.go` or create `internal/app_service.go`:

```go
// Skill management
func (s *AppService) GetSkills() ([]SkillInfo, error)
func (s *AppService) RemoveSkill(name string) error

// Workspace management
func (s *AppService) GetWorkspaces() ([]WorkspaceInfo, error)
func (s *AppService) GetActiveWorkspace() (string, error)
func (s *AppService) SetActiveWorkspace(name string) error
func (s *AppService) CreateWorkspace(name, description string) error
func (s *AppService) DeleteWorkspace(name string) error

// Agent status
func (s *AppService) GetAgentStatus() ([]AgentStatus, error)
```

**AgentStatus struct (Go):**
```go
type AgentStatus struct {
    ID             string `json:"id"`
    DisplayName    string `json:"displayName"`
    Installed      bool   `json:"installed"`
    SkillCount     int    `json:"skillCount"`
    GlobalSkillsDir string `json:"globalSkillsDir"`
}
```

### Phase 7: System Tray Update

**Update `cmd/scribe/gui_cgo.go`:**
- Show installed skills count
- Show current workspace
- Quick workspace switching submenu
- Status indicator

### Phase 7: Cleanup

**Remove:**
- `internal/marketplace.go` - No longer needed
- `internal/claude.go` - No longer managing Claude settings
- `internal/plugins.go` - Replaced by skills.go
- Old plugin types from `types.go`

**Update documentation:**
- `docs/cli-spec.md` - New command structure
- `docs/configuration.md` - New storage layout
- `docs/development.md` - Updated architecture

---

## Architecture Decisions (Finalized)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Migration** | Clean break | No migration of existing `~/.scribe` data |
| **Skill Scope** | **Global only** | No per-project skills - all skills managed globally |
| **Agent Management** | **Unified** | All coding agents managed globally by Scribe |
| **Workspace Scope** | Global only | Workspaces manage `~/.scribe/scrolls/` symlinks only |
| **Auto-add to Workspace** | Yes | Skills automatically added to active workspace on install |
| **Agent Selection** | All detected | Install to all agents with existing config directories |
| **URL Scheme** | Keep `agenthub://` | Maintain web integration compatibility |
| **GUI/System Tray** | Keep and update | Maintain GUI for non-technical users |

---

## File Changes Summary

### Backend: Files to Remove (pending cleanup)
```
internal/
├── marketplace.go          # No longer generating marketplace.json
├── claude.go               # No longer managing Claude settings
├── plugins.go              # Replaced by skills.go
└── registry.go             # Replaced by meta.go
```

### Backend: Files Added ✅
```
internal/
├── agents.go               ✅ 45 agent definitions with detection
├── skills.go               ✅ Skill discovery and SKILL.md parsing
├── installer.go            ✅ Symlink-based installation to agents
├── workspace.go            ✅ Workspace CRUD and switching
├── meta.go                 ✅ Sidecar .scribe-meta.json management
├── storage.go              ✅ Canonical storage paths
└── url_scheme.go           ✅ agenthub:// URL scheme handler

cli/
├── install.go              ✅ Updated for skills + zip URL download
├── uninstall.go            ✅ Updated for skills (replaced old plugin uninstall)
├── workspace.go            ✅ Workspace commands
├── check.go                ✅ Check for skill updates (content hash comparison)
└── update.go               ✅ Update installed skills (re-fetch and reinstall)
```

### Backend: Files Modified ✅
```
main.go                     ✅ Wails bindings for skills/workspaces/agents, system tray updated

internal/
├── types.go                ✅ New types (Skill, Agent, Workspace, etc.) + legacy types
├── skills.go               ✅ Added GetAgentsWithSkill export

cli/
├── root.go                 ✅ Added workspace, check, update commands
├── cli_test.go             ✅ Updated tests for skills system
├── list.go                 ✅ Updated for skills-only listing
└── info.go                 ✅ Updated for skill details
```

### Backend: Files to Modify (pending)
```
internal/
└── source.go               🚧 Add well-known parsing (optional)
```

### Frontend: Files Added ✅
```
frontend/src/
├── components/
│   ├── SkillList.vue           ✅ Main skill list component
│   ├── SkillCard.vue           ✅ Individual skill card with agent badges
│   ├── WorkspaceSelector.vue   ✅ Workspace dropdown in header
│   └── AgentStatusPanel.vue    ✅ Sidebar with agent status grid
├── composables/
│   ├── useSkills.ts            ✅ Skill state management
│   ├── useWorkspaces.ts        ✅ Workspace state management
│   └── useAgents.ts            ✅ Agent detection and status
└── types/
    └── skill.ts                ✅ SkillInfo, WorkspaceInfo, AgentStatus types
```

### Frontend: Files Modified ✅
```
frontend/src/
├── App.vue                 ✅ New layout with sidebar and workspace selector
├── components/
│   └── EmptyState.vue      ✅ Updated text for skills
└── bindings/
    └── scribe.ts           # (auto-generated from Go)
```

---

## Implementation Notes

### Symlink Handling

```go
func createSymlink(target, link string) error {
    // Use relative paths for portability
    relPath, err := filepath.Rel(filepath.Dir(link), target)
    if err != nil {
        return err
    }

    // On Windows, use junction for directories
    if runtime.GOOS == "windows" {
        return createJunction(target, link)
    }

    return os.Symlink(relPath, link)
}
```

### Path Sanitization

```go
func sanitizeName(name string) string {
    // Convert to lowercase kebab-case
    // Remove path traversal attempts (../, ./)
    // Limit to 255 characters
    // Replace special characters with hyphens
}
```

### Skill Discovery Priority

1. Direct SKILL.md in target path
2. `skills/` subdirectory
3. Agent-specific directories (`.claude/skills/`, etc.)
4. Recursive search with depth limit

---

## Implementation Order

### MVP (Minimum Viable Product) - Backend ✅ COMPLETED
1. ✅ **Types** - Define Skill, SkillMeta, Agent, Workspace, Config structs
2. ✅ **Agents** - Port 45 agents (agents.ts → agents.go)
3. ✅ **Skills** - SKILL.md parsing with frontmatter extraction
4. ✅ **Storage** - Canonical paths and directory creation
5. ✅ **Installer** - Symlink creation with copy fallback
6. ✅ **Meta** - Sidecar .scribe-meta.json read/write
7. ✅ **Workspaces** - Full workspace system
8. ✅ **CLI install** - Install skills from GitHub/local/GitLab
9. ✅ **CLI uninstall** - Remove skills
10. ✅ **CLI workspace** - All workspace commands

### MVP - Frontend ✅ COMPLETED
11. ✅ **Types** - SkillInfo, WorkspaceInfo TypeScript types
12. ✅ **Rename components** - PluginList → SkillList, PluginCard → SkillCard
13. ✅ **useSkills composable** - Replace usePlugins
14. ✅ **Wails bindings** - AppService methods for skills

### Full Feature Set - Backend ✅ COMPLETED
15. ✅ **Source parsing** - GitHub, GitLab, local paths implemented
16. ✅ **Source parsing** - Zip URLs implemented (well-known endpoints pending)
17. ✅ **Skill discovery** - Recursive SKILL.md discovery
18. ✅ **Check/Update** - `scribe check` and `scribe update` commands
19. ✅ **CLI list/info** - Updated for skills-only system

### Full Feature Set - Frontend ✅ COMPLETED
20. ✅ **WorkspaceSelector** - Dropdown component in header
21. ✅ **AgentStatusPanel** - Agent grid with status
22. ✅ **useWorkspaces** - Workspace state composable
23. ✅ **useAgents** - Agent detection composable
24. ✅ **SkillCard updates** - Show agent badges
25. ✅ **System tray** - Workspace switching menu

### Testing (Alongside Development)
26. ✅ **CLI tests** - Test command parsing and execution
27. ✅ **Backend unit tests** - 72.5% coverage (`internal/skills_system_test.go`)
28. ✅ **Docker test infrastructure** - `test.Dockerfile`, `docker-compose.test.yml`, Makefile targets
29. ✅ **Frontend component tests** - Vitest + Vue Test Utils (63 tests)
30. 🚧 **Integration tests** - Full install/remove/workspace flows

### Polish 🚧 PENDING
30. ✅ **URL scheme** - Updated agenthub:// handler (internal/url_scheme.go)
31. **Error handling** - Comprehensive error messages
32. **Documentation** - Update all docs
33. **Cleanup** - Remove deprecated code (after GUI testing)

---

## Estimated Effort

| Phase | Files | Complexity |
|-------|-------|------------|
| Phase 1: Core Infrastructure | 4 new Go files | Medium |
| Phase 2: Skill Discovery | 2 new, 1 modified | Medium |
| Phase 3: Installation | 2 new | Medium |
| Phase 4: Workspaces | 1 new | Medium |
| Phase 5: CLI Commands | 5 new, 3 modified | Low |
| Phase 6: Frontend | 3 renamed, 3 new Vue/TS | Medium |
| Phase 7: System Tray | 2 modified | Low |
| Phase 8: Cleanup | 4 removed | Low |

### Frontend Summary
| Component | Action | Complexity |
|-----------|--------|------------|
| PluginList.vue → SkillList.vue | Rename + update | Low |
| PluginCard.vue → SkillCard.vue | Rename + add agent badges | Medium |
| WorkspaceSelector.vue | New component | Medium |
| **AgentStatusPanel.vue** | **New component - agent grid with status** | **Medium** |
| useSkills.ts | New composable | Low |
| useWorkspaces.ts | New composable | Low |
| **useAgents.ts** | **New composable - agent detection** | **Low** |
| App.vue | Add sidebar with agent panel + workspace selector | Medium |

---

## Testing Strategy

### Backend Unit Tests (Go)

| Package | Test File | Coverage |
|---------|-----------|----------|
| `internal/agents` | `agents_test.go` | Agent detection, config paths |
| `internal/skills` | `skills_test.go` | SKILL.md parsing, frontmatter extraction |
| `internal/meta` | `meta_test.go` | Sidecar read/write, hash computation |
| `internal/installer` | `installer_test.go` | Symlink creation, copy fallback |
| `internal/workspace` | `workspace_test.go` | CRUD operations, switching logic |
| `internal/storage` | `storage_test.go` | Path resolution, directory creation |
| `internal/source` | `source_test.go` | URL parsing, source type detection |

**Key test cases:**

```go
// agents_test.go
func TestDetectInstalledAgents(t *testing.T)      // Mock filesystem, verify detection
func TestGetAgentByID(t *testing.T)               // Valid/invalid agent IDs
func TestAgentPaths(t *testing.T)                 // Project vs global paths

// skills_test.go
func TestParseSkillMd(t *testing.T)               // Valid frontmatter
func TestParseSkillMd_MissingName(t *testing.T)   // Error handling
func TestDiscoverSkills(t *testing.T)             // Recursive discovery
func TestSanitizeName(t *testing.T)               // Path traversal prevention

// installer_test.go
func TestCreateSymlink(t *testing.T)              // Unix symlinks
func TestCreateSymlink_Windows(t *testing.T)      // Windows junctions
func TestCopyFallback(t *testing.T)               // When symlink fails
func TestInstallToMultipleAgents(t *testing.T)    // Fan-out installation

// workspace_test.go
func TestSwitchWorkspace(t *testing.T)            // Symlink diff and sync
func TestAddSkillToWorkspace(t *testing.T)        // Skill addition
func TestRemoveSkillFromWorkspace(t *testing.T)   // Skill removal
```

### Integration Tests (Go)

```go
// integration_test.go
func TestFullInstallFlow(t *testing.T) {
    // 1. Parse GitHub source
    // 2. Clone/fetch skills
    // 3. Discover SKILL.md files
    // 4. Install to canonical location
    // 5. Create symlinks to all detected agents
    // 6. Write sidecar metadata
    // 7. Verify symlinks resolve correctly
}

func TestWorkspaceSwitching(t *testing.T) {
    // 1. Install multiple skills
    // 2. Create workspace with subset
    // 3. Switch workspace
    // 4. Verify correct symlinks exist/removed
}

func TestRemoveSkill(t *testing.T) {
    // 1. Install skill
    // 2. Remove skill
    // 3. Verify symlinks removed from all agents
    // 4. Verify canonical copy removed
}
```

### CLI Tests

```go
// cli/add_test.go
func TestAddCommand_GitHubShorthand(t *testing.T)
func TestAddCommand_LocalPath(t *testing.T)
func TestAddCommand_InvalidSource(t *testing.T)

// cli/workspace_test.go
func TestWorkspaceCreate(t *testing.T)
func TestWorkspaceUse(t *testing.T)
func TestWorkspaceList(t *testing.T)
```

### Frontend Tests (Vitest + Vue Test Utils)

| Component | Test File | Coverage |
|-----------|-----------|----------|
| `SkillList.vue` | `SkillList.spec.ts` | Rendering, filtering, empty state |
| `SkillCard.vue` | `SkillCard.spec.ts` | Display, agent badges, uninstall |
| `AgentStatusPanel.vue` | `AgentStatusPanel.spec.ts` | Grid render, click handling |
| `WorkspaceSelector.vue` | `WorkspaceSelector.spec.ts` | Dropdown, switching |

**Key test cases:**

```typescript
// SkillList.spec.ts
describe('SkillList', () => {
  it('renders skills from composable')
  it('shows empty state when no skills')
  it('filters by selected agent')
  it('emits uninstall event')
})

// AgentStatusPanel.spec.ts
describe('AgentStatusPanel', () => {
  it('renders all agents in grid')
  it('shows checkmark for installed agents')
  it('shows skill count for installed agents')
  it('grays out non-installed agents')
  it('emits agent-selected on click')
})

// useAgents.spec.ts
describe('useAgents', () => {
  it('fetches agent status on mount')
  it('computes installed count correctly')
  it('filters skills when agent selected')
})
```

### E2E Tests (Optional - Playwright)

```typescript
// e2e/install-skill.spec.ts
test('install skill from GitHub', async ({ page }) => {
  // Test full flow through GUI
})

// e2e/workspace-switch.spec.ts
test('switch workspace updates skill list', async ({ page }) => {
  // Test workspace switching through GUI
})
```

### Test Infrastructure

**Go test setup:**
```go
// testutil/testutil.go
func TempHome(t *testing.T) string           // Create temp ~/.scribe
func MockAgent(t *testing.T, id string)      // Create mock agent config dir
func CreateTestSkill(t *testing.T, name string) string  // Create test SKILL.md
```

**Frontend test setup:**
```typescript
// test/setup.ts
vi.mock('../bindings/scribe', () => ({
  AppService: {
    GetSkills: vi.fn(),
    GetAgentStatus: vi.fn(),
    GetWorkspaces: vi.fn(),
  }
}))
```

### Test Commands

```bash
# Backend (local)
go test ./...                    # All tests
go test ./internal/... -v        # Verbose internal tests
go test -race ./...              # Race detection
go test -cover ./...             # Coverage report

# Backend (Docker) - consistent isolated environment
make docker-test                 # Run all tests in Docker
make docker-test-coverage        # Tests with coverage report
make docker-test-race            # Tests with race detector
make docker-test-filter TEST_PATTERN=TestSkill  # Filter tests

# Frontend
pnpm test                        # Run Vitest
pnpm test:coverage               # Coverage report
pnpm test:e2e                    # Playwright E2E (if added)
```

---

## Current Status

**MVP Backend: COMPLETE** ✅

The core backend infrastructure is implemented and working:
- Skills can be installed from GitHub repos, local paths, GitLab, and **zip URLs**
- Skills are stored in `~/.scribe/scrolls/` with sidecar metadata
- Symlinks are created in all detected agent directories (45 agents supported)
- Workspace system allows organizing skills into named sets
- CLI commands: `install`, `uninstall`, `workspace list/create/use/add/remove/current/delete`
- **Test coverage at 72.5%** with Docker test infrastructure for CI consistency
- **CI workflow updated** to Go 1.25 to match go.mod requirements

**MVP Frontend: COMPLETE** ✅

The Vue 3 frontend has been fully migrated to the skills-only architecture:
- **Wails bindings** expose skills, workspaces, and agents APIs to frontend
- **New components**: SkillList, SkillCard, WorkspaceSelector, AgentStatusPanel
- **New composables**: useSkills, useWorkspaces, useAgents
- **Updated layout**: App.vue now has sidebar with agent panel and workspace selector
- **System tray** shows skills count and active workspace
- **Frontend builds successfully** with TypeScript type checking

**CLI Migration: COMPLETE** ✅

All CLI commands have been migrated to the skills-only system:
- **`scribe list`** - Now shows skills with description, source, agents
- **`scribe info`** - Now shows skill details including content hash
- **`scribe check`** - Check installed skills for updates (compares content hashes)
- **`scribe update`** - Update outdated skills from their sources
- **Zip URL support** - `scribe install https://example.com/skills.zip` now works
- **Tests updated** - CLI tests migrated from legacy plugin system to skills

**Legacy Cleanup: COMPLETE** ✅

All deprecated plugin code has been removed:
- **Removed Go files**: marketplace.go, claude.go, plugins.go, registry.go, url_scheme.go, source.go, server.go
- **Removed legacy types**: PluginSource, Plugin, Author, RegistryEntry, PluginProvides
- **Removed Vue components**: PluginCard.vue, PluginList.vue
- **Removed composables/types**: usePlugins.ts, plugin.ts
- **Updated main.go**: Skills-only API, no legacy plugin code
- **Updated CLI**: root.go no longer uses Server
- **All tests passing**: 54 internal tests, 12 CLI tests

**Next Steps:**
1. **End-to-end testing** - Test full GUI workflow with skills system
2. **Additional source types** - Well-known endpoints (/.well-known/skills/)
