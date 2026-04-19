# Scribe

![Go Coverage](https://img.shields.io/endpoint?url=https%3A%2F%2Fgitlab.com%2Fusescrolls%2Fscribe%2F-%2Fjobs%2Fartifacts%2Fmain%2Fraw%2Fbadges%2Fgo-coverage.json%3Fjob%3Dgo-test)
![Frontend Coverage](https://img.shields.io/endpoint?url=https%3A%2F%2Fgitlab.com%2Fusescrolls%2Fscribe%2F-%2Fjobs%2Fartifacts%2Fmain%2Fraw%2Fbadges%2Ffrontend-coverage.json%3Fjob%3Dfrontend)

A skill distribution tool that syncs AI coding skills to 39 coding agents.

```bash
curl -fsSL https://gitlab.com/usescrolls/scribe/-/raw/main/scripts/install.sh | bash
```

![Scribe Demo](https://cdn.usescrolls.com/images/demo_scribe.gif)

## Overview

Scribe lets you install AI coding skills once and automatically distributes them to all your coding agents (Claude Code, Cursor, Copilot, Cline, etc.) via symlinks. Instead of manually copying skills to each agent's directory, Scribe maintains a central skill library and keeps all your agents in sync.

On first launch, Scribe runs an onboarding wizard that detects your installed agents, optionally imports existing skills, and installs a demo skill to get you started.

### Why Scribe?

If you've seen `npx skills add`, you might wonder why Scribe exists. Both tools install the same `SKILL.md` files to coding agents, but they make different trade-offs:

- **No runtime dependency**: Scribe is a standalone Go binary — no Node.js required. Install once via `curl`, runs offline after that
- **Fast**: Native binary + local git clone cache. No npm registry roundtrip on every invocation
- **Workspaces**: Organize skills into named sets (e.g. `web-dev`, `data-eng`) and switch between them
- **Desktop GUI**: Vue 3 app with skill browser, workspace manager, and install wizard. You can paste `npx skills add owner/repo` commands directly into the Install tab — Scribe parses them automatically
- **Guided onboarding**: First-run wizard detects your agents, imports existing skills, and installs a demo skill
- **39 agents supported**: Claude Code, Cursor, GitHub Copilot, Cline, Windsurf, Continue, and more
- **Private repositories**: Automatic credential resolution via git credential helpers and SSH agent
- **Multiple sources**: GitHub, GitLab, Bitbucket, local paths, zip URLs, SSH URLs
- **Version tracking**: Git commit hash and date tracking with update detection
- **No telemetry**: Scribe collects nothing

Scribe is fully compatible with the [Agent Skills specification](https://agentskills.io) and installs the same skills that `npx skills` does.

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

## Other Installation Methods

For Windows and building from source, see [Installation](docs/installation.md).

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
