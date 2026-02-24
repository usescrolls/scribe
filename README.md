# Scribe

![Go Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/nunomen/f7a526db56e4e8869e1a3ff5bae38b38/raw/go-coverage.json)
![Frontend Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/nunomen/f7a526db56e4e8869e1a3ff5bae38b38/raw/frontend-coverage.json)

A skill distribution tool that syncs AI coding skills to 39 coding agents.

## Overview

Scribe lets you install AI coding skills once and automatically distributes them to all your coding agents (Claude Code, Cursor, Copilot, Cline, etc.) via symlinks. Instead of manually copying skills to each agent's directory, Scribe maintains a central skill library and keeps all your agents in sync.

On first launch, Scribe runs an onboarding wizard that detects your installed agents, optionally imports existing skills, and installs a demo skill to get you started.

### Key Features

- **Install Once, Use Everywhere**: Skills are automatically symlinked to all detected agents
- **39 Agents Supported**: Claude Code, Cursor, GitHub Copilot, Cline, Windsurf, Continue, and more
- **Workspaces**: Organize skills into named sets and switch between them
- **Private Repositories**: Automatic credential resolution via git credential helpers and SSH agent
- **Multiple Sources**: GitHub, GitLab, Bitbucket, local paths, zip URLs, SSH URLs
- **Desktop GUI**: Vue 3 frontend with workspace management, skill browser, and multi-step install wizard (you can paste CLI commands like `npx skills add owner/repo` directly into the Install tab)
- **Version Tracking**: Git commit hash and date tracking for installed skills with update detection
- **Local Cache**: Git clone cache for fast installs, checks, and updates

## Supported Agents

Scribe distributes skills to 39 coding agents (when installed):

| Agent | Skills Directory |
|-------|-----------------|
| Claude Code | `~/.claude/skills/` |
| Amp | `~/.config/agents/skills/` |
| Augment | `~/.augment/rules/` |
| Cline | `~/.cline/skills/` |
| Codex | `~/.codex/skills/` |
| Continue | `~/.continue/skills/` |
| Cursor | `~/.cursor/skills/` |
| Gemini CLI | `~/.gemini/skills/` |
| GitHub Copilot | `~/.copilot/skills/` |
| Goose | `~/.config/goose/skills/` |
| Kilo Code | `~/.kilocode/skills/` |
| Kiro CLI | `~/.kiro/skills/` |
| OpenCode | `~/.config/opencode/skills/` |
| Roo Code | `~/.roo/skills/` |
| Trae | `~/.trae/skills/` |
| Windsurf | `~/.codeium/windsurf/skills/` |
| Zencoder | `~/.zencoder/skills/` |
| + 22 more... | |

Scribe detects which agents you have installed (by checking for their config directories) and only creates symlinks for those agents. See [Configuration](docs/configuration.md) for the full agent list.

## Quick Start

### macOS (Homebrew)

```bash
brew install usescrolls/tap/scribe
```

### macOS (DMG)

Download the DMG from the [latest GitHub release](https://github.com/usescrolls/scribe/releases/latest), open it, and drag Scribe to your Applications folder.

### Other Platforms

For Linux, Windows, and other installation methods, see [Installation](docs/installation.md).

## CLI Usage

```bash
# Install skills from various sources
scribe install owner/repo                          # GitHub shorthand
scribe install https://github.com/owner/repo       # Full GitHub URL
scribe install https://gitlab.com/owner/repo       # GitLab URL
scribe install https://bitbucket.org/owner/repo    # Bitbucket URL
scribe install ./local/path                        # Local directory
scribe install https://example.com/skills.zip      # Zip URL
scribe install git@github.com:owner/repo.git       # SSH URL (private repos)
scribe install owner/repo --all                    # Install all skills to all agents

# List installed skills
scribe list
scribe ls --json
scribe list --names-only

# Show skill details
scribe info <skill-name>

# Check for updates
scribe check                    # Check all skills
scribe check <skill-name>       # Check a specific skill

# Update skills
scribe update                    # Update all outdated skills
scribe update <skill-name>       # Update a specific skill
scribe update --force            # Force update even if up-to-date

# Remove a skill
scribe uninstall <skill-name>
scribe rm <skill-name>
scribe uninstall --all           # Remove all installed skills

# First-run onboarding
scribe setup

# Cache management
scribe cache path                # Print cache directory path
scribe cache clear               # Clear the local clone cache
```

### Workspace Commands

```bash
# List workspaces
scribe workspace list

# Create a workspace
scribe workspace create <name> --description "..."

# Switch workspaces
scribe workspace use <name>

# Add/remove skills from workspace
scribe workspace add <skill-name>
scribe workspace remove <skill-name>

# Show current workspace
scribe workspace current

# Delete a workspace
scribe workspace delete <name>
```

For complete CLI documentation, see [CLI Specification](docs/cli-spec.md).

## Private Repositories

Scribe supports private repositories out of the box by leveraging your existing git credentials. No Scribe-specific configuration is needed.

**HTTPS (recommended):** Scribe uses the system `git credential` helper, which automatically picks up credentials from:
- `gh auth login` (GitHub CLI)
- macOS Keychain
- Windows Credential Manager
- `git-credential-store` or any configured credential helper

**SSH:** Use an SSH URL to clone via your SSH key:
```bash
scribe install git@github.com:owner/private-repo.git
scribe install git@gitlab.com:org/private-repo.git
```

Scribe reads keys from your running SSH agent. If authentication fails, Scribe shows a hint with setup instructions.

## Architecture

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
├── cache/                      # Local git clone cache
└── config.json                 # Active workspace, onboarding state

Skills are symlinked to each agent's skills directory:
~/.claude/skills/react-best-practices -> ~/.scribe/scrolls/react-best-practices
~/.cursor/skills/react-best-practices -> ~/.scribe/scrolls/react-best-practices
```

## Documentation

- [Installation](docs/installation.md) - All installation methods and background service setup
- [CLI Specification](docs/cli-spec.md) - Complete CLI command reference
- [Install Badge](docs/install-badge.md) - One-click install badges and `agenthub://` URL scheme
- [Configuration](docs/configuration.md) - Storage layout and settings
- [Troubleshooting](docs/troubleshooting.md) - Common issues and solutions
- [Development](docs/development.md) - Building, testing, and architecture reference

## Created By

**Nuno L Mendes** — creator and maintainer of Scribe.

- GitHub: [@nunomen](https://github.com/nunomen)

## Sponsors

Scribe is free and open source. If you or your organization find it useful, consider sponsoring the project to support ongoing development.

<a href="https://github.com/sponsors/nunomen">
  <img src="https://img.shields.io/badge/Sponsor-Scribe-EA4AAA?logo=github-sponsors&logoColor=white&style=for-the-badge" alt="Sponsor Scribe" />
</a>

Sponsors help fund new features, agent integrations, and long-term maintenance. Thank you to all our supporters!

## License

MIT

---

<p align="center">
  <sub>🇨🇭 Made with precision in Switzerland</sub>
</p>
