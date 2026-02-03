# Scribe

A skill distribution tool that syncs AI coding skills to 45+ coding agents.

## Overview

Scribe lets you install AI coding skills once and automatically distributes them to all your coding agents (Claude Code, Cursor, Copilot, Cline, etc.) via symlinks. Instead of manually copying skills to each agent's directory, Scribe maintains a central skill library and keeps all your agents in sync.

### Architecture

```
~/.scribe/
├── scrolls/                    # Canonical skill storage
│   ├── react-best-practices/
│   │   ├── SKILL.md
│   │   └── .scribe-meta.json   # Source tracking
│   └── typescript-patterns/
│       ├── SKILL.md
│       └── .scribe-meta.json
├── workspaces/                 # Workspace definitions
│   ├── default.json
│   └── web-dev.json
└── config.json                 # Active workspace

Skills are symlinked to each agent's skills directory:
~/.claude/skills/react-best-practices -> ~/.scribe/scrolls/react-best-practices
~/.cursor/skills/react-best-practices -> ~/.scribe/scrolls/react-best-practices
```

### Key Features

- **Install Once, Use Everywhere**: Skills are automatically symlinked to all detected agents
- **45+ Agents Supported**: Claude Code, Cursor, GitHub Copilot, Cline, Windsurf, Continue, and more
- **Workspaces**: Organize skills into named sets and switch between them
- **Multiple Sources**: GitHub, GitLab, local paths, zip URLs
- **Desktop GUI**: Vue 3 frontend with workspace selector and skill browser

## Quick Start

### macOS (Homebrew)

```bash
brew install usescrolls/tap/scribe
```

### macOS (DMG)

Download the DMG from [usescrolls.com/releases](https://usescrolls.com/releases), open it, and drag Scribe to your Applications folder.

### Other Platforms

For Linux, Windows, and other installation methods, see [Installation](docs/installation.md).

## CLI Usage

```bash
# Install skills from various sources
scribe install owner/repo                    # GitHub shorthand
scribe install https://github.com/owner/repo # Full GitHub URL
scribe install ./local/path                  # Local directory
scribe install https://example.com/skills.zip # Zip URL

# List installed skills
scribe list
scribe ls --json

# Show skill details
scribe info <skill-name>

# Check for updates
scribe check

# Update all skills
scribe update

# Remove a skill
scribe uninstall <skill-name>
scribe rm <skill-name>
```

### Workspace Commands

```bash
# List workspaces
scribe workspace list

# Create a workspace
scribe workspace create <name>

# Switch workspaces
scribe workspace use <name>

# Add/remove skills from workspace
scribe workspace add <skill-name>
scribe workspace remove <skill-name>

# Show current workspace
scribe workspace current
```

For complete CLI documentation, see [CLI Specification](docs/cli-spec.md).

## URL Scheme

Scribe supports the `agenthub://` URL scheme for one-click installs from websites:

```
agenthub://install?source=github&repo=owner/repo
```

1. Click an install link on [useScrolls.com](https://usescrolls.com)
2. OS launches Scribe with the URL
3. Scribe installs the skill and symlinks to all detected agents

## Supported Agents

Scribe can distribute skills to these coding agents (when installed):

| Agent | Skills Directory |
|-------|-----------------|
| Claude Code | `~/.claude/skills/` |
| Cursor | `~/.cursor/skills/` |
| GitHub Copilot | `~/.copilot/skills/` |
| Cline | `~/.cline/skills/` |
| Continue | `~/.continue/skills/` |
| Windsurf | `~/.codeium/windsurf/skills/` |
| Codex | `~/.codex/skills/` |
| Gemini CLI | `~/.gemini/skills/` |
| Goose | `~/.config/goose/skills/` |
| + 35 more... | |

Scribe detects which agents you have installed (by checking for their config directories) and only creates symlinks for those agents.

## Documentation

- [Installation](docs/installation.md) - All installation methods and background service setup
- [CLI Specification](docs/cli-spec.md) - Complete CLI command reference
- [Configuration](docs/configuration.md) - Storage layout and settings
- [Troubleshooting](docs/troubleshooting.md) - Common issues and solutions
- [Development](docs/development.md) - Building, testing, and architecture reference

## License

MIT
