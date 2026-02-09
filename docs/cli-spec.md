# Scribe CLI Specification

## Overview

Scribe CLI distributes AI coding skills to 39 coding agents. Built with [Cobra](https://github.com/spf13/cobra).

## Command Structure

```bash
scribe <command> [flags] [arguments]
```

## Commands

### `scribe install`

Install skills from a source.

```bash
scribe install <source> [flags]
```

**Sources:**
| Format | Example |
|--------|---------|
| GitHub shorthand | `owner/repo` |
| GitHub URL | `https://github.com/owner/repo` |
| GitHub with branch | `owner/repo#branch` |
| GitHub with path | `owner/repo/path/to/skills` |
| GitLab URL | `https://gitlab.com/owner/repo` |
| Bitbucket URL | `https://bitbucket.org/owner/repo` |
| Local path | `./local/path` or `/absolute/path` |
| Zip URL | `https://example.com/skills.zip` |

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--agent` | `-a` | Target specific agents (comma-separated) |
| `--skill` | `-s` | Select specific skills to install |
| `--list` | `-l` | List available skills without installing |
| `--yes` | `-y` | Skip interactive prompts |
| `--all` | | Install all skills to all detected agents |

**Examples:**
```bash
scribe install vercel-labs/agent-skills
scribe install https://github.com/owner/repo
scribe install ./my-local-skills
scribe install https://example.com/skills.zip
scribe install owner/repo --skill react-patterns --agent claude-code,cursor
scribe install owner/repo --list
```

**Multi-skill sources:** When a source contains multiple skills, Scribe discovers all of them. Use `--list` to preview, `--skill` to pick specific ones (comma-separated), or `--all` to install everything.

```bash
scribe install owner/multi-skill-repo --list                             # Preview available skills
scribe install owner/multi-skill-repo --skill react-patterns,go-idioms   # Install two specific skills
scribe install owner/multi-skill-repo --all                              # Install all discovered skills
```

### `scribe uninstall`

Remove an installed skill.

```bash
scribe uninstall <skill-name> [flags]
```

**Aliases:** `remove`, `rm`

**Flags:**
| Flag | Description |
|------|-------------|
| `--all` | Remove all installed skills |

**Examples:**
```bash
scribe uninstall react-patterns
scribe rm typescript-best-practices
scribe uninstall --all
```

### `scribe list`

List installed skills.

```bash
scribe list [flags]
```

**Aliases:** `ls`

**Flags:**
| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format |
| `--names-only` | Print only skill names, one per line |

**Examples:**
```bash
scribe list
scribe ls --json
scribe list --names-only
```

**Output (default):**
```
NAME                    DESCRIPTION                              SOURCE                    AGENTS
react-best-practices    React patterns and best practices        github:vercel-labs/...    claude-code, cursor
typescript-patterns     TypeScript coding standards              github:owner/repo         claude-code, cline

2 skill(s) installed
```

**Output (JSON):**
```json
{
  "skills": [
    {
      "name": "react-best-practices",
      "description": "React patterns and best practices",
      "source": "vercel-labs/agent-skills",
      "sourceType": "github",
      "installedAt": "2025-01-15T10:30:00Z",
      "agents": ["claude-code", "cursor", "cline"]
    }
  ],
  "count": 1
}
```

### `scribe info`

Show detailed information about an installed skill.

```bash
scribe info <skill-name>
```

**Examples:**
```bash
scribe info react-best-practices
```

**Output:**
```
Name:         react-best-practices
Description:  React patterns and best practices
Source:       github:vercel-labs/agent-skills
Source URL:   https://github.com/vercel-labs/agent-skills
Skill Path:   skills/react-best-practices
Content Hash: sha256:abc123...
Installed:    2025-01-15 10:30:00
Updated:      2025-01-20 14:00:00
Agents:       claude-code, cursor, cline, windsurf
```

### `scribe check`

Check installed skills for available updates. Without arguments, checks all installed skills. With a skill name, checks only that skill.

```bash
scribe check [skill-name] [flags]
```

**Flags:**
| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format |

**Examples:**
```bash
scribe check                    # Check all skills
scribe check react-patterns     # Check a specific skill
```

**Output:**
```
Checking 5 skills for updates...

SKILL                   STATUS
react-best-practices    Update available (content changed)
typescript-patterns     Up to date
go-patterns             Update available (content changed)

2 skill(s) have updates available
Run 'scribe update' to update all skills
```

### `scribe update`

Update installed skills to their latest versions. Without arguments, updates all outdated skills. With a skill name, updates only that skill.

```bash
scribe update [skill-name] [flags]
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Force update even if up-to-date |

**Examples:**
```bash
scribe update                    # Update all outdated skills
scribe update react-patterns     # Update specific skill
scribe update --force            # Force update all skills
```

### `scribe workspace`

Manage workspaces for organizing skills.

#### `scribe workspace list`

List all workspaces.

```bash
scribe workspace list
```

**Output:**
```
NAME        SKILLS  DESCRIPTION
default     5       All installed skills
web-dev     3       Web development skills
backend     2       Backend development skills

* Active: web-dev
```

#### `scribe workspace create`

Create a new workspace.

```bash
scribe workspace create <name> [--description "..."]
```

**Examples:**
```bash
scribe workspace create mobile-dev --description "Mobile development skills"
```

#### `scribe workspace use`

Switch to a different workspace.

```bash
scribe workspace use <name>
```

**Examples:**
```bash
scribe workspace use backend
```

**Output:**
```
Switching from 'web-dev' to 'backend'...
- Removed: react-best-practices
- Removed: typescript-patterns
+ Added: go-patterns
+ Added: api-design

Active workspace: backend (2 skills)
```

#### `scribe workspace add`

Add a skill to the current workspace.

```bash
scribe workspace add <skill-name>
```

#### `scribe workspace remove`

Remove a skill from the current workspace.

```bash
scribe workspace remove <skill-name>
```

#### `scribe workspace current`

Show the current active workspace.

```bash
scribe workspace current
```

#### `scribe workspace delete`

Delete a workspace.

```bash
scribe workspace delete <name>
```

### `scribe setup`

Run the first-time onboarding wizard. Detects installed agents, optionally imports existing skills from agent directories, and installs a demo skill.

```bash
scribe setup
```

This runs automatically on first use if onboarding hasn't been completed. You can also run it manually to re-run the setup process.

### `scribe cache`

Manage the local clone cache. Scribe caches cloned repositories to speed up subsequent installs, checks, and updates.

#### `scribe cache path`

Print the cache directory path.

```bash
scribe cache path
```

#### `scribe cache clear`

Clear the entire clone cache.

```bash
scribe cache clear
```

### `scribe help`

Show help for any command.

```bash
scribe help [command]
scribe --help
scribe <command> --help
```

### `scribe version`

Show version information.

```bash
scribe version
scribe --version
```

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--debug` | | Enable debug logging |
| `--json` | | Output in JSON format (where applicable) |
| `--quiet` | `-q` | Suppress non-essential output |
| `--help` | `-h` | Show help |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid usage / bad arguments |
| 3 | Skill not found |
| 4 | Source resolution failed |
| 5 | Filesystem error |

## URL Scheme Handling

The `agenthub://` URL scheme is supported for one-click installs:

```bash
scribe "agenthub://install?source=github&repo=owner/repo"
```

## GUI Mode

Running without arguments launches the desktop GUI:

```bash
scribe              # Launch GUI with system tray
scribe --no-gui     # Run in headless mode
```

## Detection Order

When scribe is invoked:

1. If first argument starts with `agenthub://` → URL scheme handler
2. If first argument is a known command → CLI mode
3. If no arguments → GUI mode (or headless with `--no-gui`)
