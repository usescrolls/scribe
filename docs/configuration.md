# Configuration

This guide covers Scribe's storage layout, source types, and settings.

## Data Storage

All Scribe data is stored in the user's home directory:
- **macOS/Linux:** `~/.scribe/`
- **Windows:** `%USERPROFILE%\.scribe\`

```
~/.scribe/
├── scrolls/                    # Canonical skill storage
│   ├── react-best-practices/
│   │   ├── SKILL.md            # Skill content
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

### Skill Storage (`~/.scribe/scrolls/`)

Each skill is stored in its own directory containing:
- **SKILL.md** - The skill content with YAML frontmatter
- **.scribe-meta.json** - Sidecar metadata tracking the source

Skills are symlinked to each detected agent's skills directory:
```
~/.claude/skills/react-best-practices -> ~/.scribe/scrolls/react-best-practices
~/.cursor/skills/react-best-practices -> ~/.scribe/scrolls/react-best-practices
```

### Sidecar Metadata (`.scribe-meta.json`)

Each skill has its own metadata file for source tracking:

```json
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

### Workspace Files (`~/.scribe/workspaces/`)

Workspaces define which skills are active:

```json
{
  "name": "web-dev",
  "description": "Web development skills only",
  "skills": ["react-best-practices", "typescript-patterns"]
}
```

When switching workspaces, Scribe:
1. Removes symlinks for skills NOT in the target workspace
2. Creates symlinks for skills IN the target workspace

### Config File (`~/.scribe/config.json`)

Global configuration:

```json
{
  "activeWorkspace": "default"
}
```

---

## Source Types

| Type | Example | Description |
|------|---------|-------------|
| GitHub Shorthand | `owner/repo` | Clone repo, discover skills |
| GitHub URL | `https://github.com/owner/repo` | Clone repo, discover skills |
| GitHub Subpath | `owner/repo/path/to/skills` | Clone repo, use subpath |
| GitHub Branch | `owner/repo#branch` | Clone specific branch |
| GitLab | `https://gitlab.com/owner/repo` | Clone repo |
| Local Path | `./local/path` | Read directly |
| Zip URL | `https://example.com/skills.zip` | Download, extract, discover skills |

### Skill Discovery

When fetching from a source, Scribe discovers skills by:

1. Looking for `SKILL.md` in root (single-skill repo)
2. Searching common directories: `skills/`, `.claude/skills/`, etc.
3. Recursive search (max depth 5, excludes: `node_modules`, `.git`, `dist`, `build`)

A valid skill is a directory containing a `SKILL.md` file with valid frontmatter (name + description).

---

## Skill Format (SKILL.md)

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

---

## Supported Agents

Scribe distributes skills by creating symlinks in each agent's skills directory. It only creates symlinks for agents that are already installed on your system (detected by their config directory existing).

| Agent | Config Dir | Skills Dir |
|-------|-----------|------------|
| Claude Code | `~/.claude` | `~/.claude/skills/` |
| Cursor | `~/.cursor` | `~/.cursor/skills/` |
| GitHub Copilot | `~/.copilot` | `~/.copilot/skills/` |
| Cline | `~/.cline` | `~/.cline/skills/` |
| Continue | `~/.continue` | `~/.continue/skills/` |
| Windsurf | `~/.codeium/windsurf` | `~/.codeium/windsurf/skills/` |
| OpenCode | `~/.config/opencode` | `~/.config/opencode/skills/` |
| Codex | `~/.codex` | `~/.codex/skills/` |
| Gemini CLI | `~/.gemini` | `~/.gemini/skills/` |
| Goose | `~/.config/goose` | `~/.config/goose/skills/` |
| + 35 more... | | |

---

## System Tray

When running with GUI (default), Scribe shows a system tray icon with:

- **Status**: Shows version and running state
- **Skills**: Shows count of installed skills
- **Workspace**: Shows active workspace with submenu to switch
- **Agents**: Shows count of detected agents
- **Quit**: Stops Scribe

---

## Security

- **URL Scheme**: Skill installation uses `agenthub://` URLs
- **IPC Security**: Linux socket uses `0600` permissions; Windows pipe uses current user security
- **Symlinks**: Skills are symlinked, not copied, so all agents share the same source
