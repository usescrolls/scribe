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
├── cache/                      # Local git clone cache
│   └── github.com/
│       └── owner/
│           └── repo/           # Cached shallow clones
└── config.json                 # Global config (active workspace, onboarding)
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
  "commitHash": "a1b2c3d",
  "commitDate": "2025-01-28T15:00:00Z",
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
  "activeWorkspace": "default",
  "onboardingCompleted": true
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
| Bitbucket | `https://bitbucket.org/owner/repo` | Clone repo |
| SSH URL | `git@github.com:owner/repo.git` | Clone via SSH (private repos) |
| Local Path | `./local/path` | Read directly |
| Zip URL | `https://example.com/skills.zip` | Download, extract, discover skills |

### Authentication

Scribe supports private repositories through two mechanisms:

- **Git credential helper (HTTPS):** Scribe calls `git credential fill` to resolve credentials from the system git credential store. This automatically picks up credentials from `gh auth login`, macOS Keychain, Windows Credential Manager, and any other configured credential helper.
- **SSH agent:** When using SSH URLs (`git@host:owner/repo.git`), Scribe reads keys from the running SSH agent.

No Scribe-specific token or credential configuration is needed. If authentication fails, Scribe displays a hint with setup instructions.

### Skill Discovery

When fetching from a source, Scribe discovers skills by:

1. Looking for `SKILL.md` in root (single-skill repo)
2. Searching common directories: `skills/`, `.claude/skills/`, etc.
3. Recursive search (max depth 5, excludes: `node_modules`, `.git`, `dist`, `build`)

A valid skill is a directory containing a `SKILL.md` file with valid frontmatter (name + description).

### Multi-Skill Sources

A single git repository or zip URL can contain multiple skills. Scribe discovers all `SKILL.md` files in the source and installs each one individually.

**Example repo structure:**
```
owner/my-skills/
├── skills/
│   ├── react-patterns/
│   │   └── SKILL.md
│   ├── typescript-tips/
│   │   └── SKILL.md
│   └── go-idioms/
│       └── SKILL.md
└── README.md
```

Running `scribe install owner/my-skills` discovers all 3 skills. Each is installed to its own directory under `~/.scribe/scrolls/` and tracked independently via `.scribe-meta.json` (the `skillPath` field records where the skill lives within the source).

**Selective install:**
```bash
scribe install owner/my-skills --list                          # List discovered skills
scribe install owner/my-skills --skill react-patterns,go-idioms  # Install specific skills
scribe install owner/my-skills --all                            # Install all to all agents
```

This works the same for all source types: GitHub, GitLab, local paths, and zip URLs.

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
| AdaL | `~/.adal` | `~/.adal/skills/` |
| Amp | `~/.config/amp` | `~/.config/agents/skills/` |
| Antigravity | `~/.gemini/antigravity` | `~/.gemini/antigravity/skills/` |
| Augment | `~/.augment` | `~/.augment/rules/` |
| Claude Code | `~/.claude` | `~/.claude/skills/` |
| Cline | `~/.cline` | `~/.cline/skills/` |
| CodeBuddy | `~/.codebuddy` | `~/.codebuddy/skills/` |
| Codex | `~/.codex` | `~/.codex/skills/` |
| Command Code | `~/.commandcode` | `~/.commandcode/skills/` |
| Continue | `~/.continue` | `~/.continue/skills/` |
| Crush | `~/.config/crush` | `~/.config/crush/skills/` |
| Cursor | `~/.cursor` | `~/.cursor/skills/` |
| Droid | `~/.factory` | `~/.factory/skills/` |
| Gemini CLI | `~/.gemini` | `~/.gemini/skills/` |
| GitHub Copilot | `~/.copilot` | `~/.copilot/skills/` |
| Goose | `~/.config/goose` | `~/.config/goose/skills/` |
| iFlow CLI | `~/.iflow` | `~/.iflow/skills/` |
| Junie | `~/.junie` | `~/.junie/skills/` |
| Kilo Code | `~/.kilocode` | `~/.kilocode/skills/` |
| Kimi CLI | `~/.kimi` | `~/.config/agents/skills/` |
| Kiro CLI | `~/.kiro` | `~/.kiro/skills/` |
| Kode | `~/.kode` | `~/.kode/skills/` |
| MCPJam | `~/.mcpjam` | `~/.mcpjam/skills/` |
| Mistral Vibe | `~/.vibe` | `~/.vibe/skills/` |
| Mux | `~/.mux` | `~/.mux/skills/` |
| Neovate | `~/.neovate` | `~/.neovate/skills/` |
| OpenClaw | `~/.openclaw` | `~/.openclaw/skills/` |
| OpenCode | `~/.config/opencode` | `~/.config/opencode/skills/` |
| OpenHands | `~/.openhands` | `~/.openhands/skills/` |
| Pi | `~/.pi/agent` | `~/.pi/agent/skills/` |
| Pochi | `~/.pochi` | `~/.pochi/skills/` |
| Qoder | `~/.qoder` | `~/.qoder/skills/` |
| Qwen Code | `~/.qwen` | `~/.qwen/skills/` |
| Replit | `~/.config/agents` | `~/.config/agents/skills/` |
| Roo Code | `~/.roo` | `~/.roo/skills/` |
| Trae | `~/.trae` | `~/.trae/skills/` |
| Trae CN | `~/.trae-cn` | `~/.trae-cn/skills/` |
| Windsurf | `~/.codeium/windsurf` | `~/.codeium/windsurf/skills/` |
| Zencoder | `~/.zencoder` | `~/.zencoder/skills/` |

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
