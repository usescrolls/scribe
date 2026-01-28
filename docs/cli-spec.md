# Scribe CLI Specification

## Overview

Scribe CLI provides intuitive command-line package management for Claude Code plugins, using [Cobra](https://github.com/spf13/cobra) as the CLI framework.

## Command Structure

```bash
scribe <command> [flags] [arguments]
```

## Commands

### `scribe install`

Install a plugin from a source.

```bash
scribe install [name] --github|--npm|--url|--zip <source> [flags]
```

**Flags:**
| Flag | Description |
|------|-------------|
| `--github` | GitHub repository (owner/repo) |
| `--npm` | NPM package name |
| `--url` | Git URL |
| `--zip` | Zip file URL |
| `--ref` | Branch or tag reference |
| `--no-enable` | Don't auto-enable in Claude settings |

**Examples:**
```bash
scribe install prettier --github usescrolls/prettier-skill
scribe install eslint --npm @anthropic/claude-eslint
scribe install custom --url https://github.com/user/plugin.git
scribe install tool --zip https://example.com/plugin.zip
scribe install prettier --github usescrolls/prettier-skill --ref v1.0.0
```

### `scribe uninstall`

Remove an installed plugin.

```bash
scribe uninstall <name> [flags]
```

**Aliases:** `remove`, `rm`

**Flags:**
| Flag | Description |
|------|-------------|
| `--all` | Remove all installed plugins |

**Examples:**
```bash
scribe uninstall prettier
scribe rm prettier
scribe uninstall --all
```

### `scribe list`

List installed plugins.

```bash
scribe list [flags]
```

**Aliases:** `ls`

**Flags:**
| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format |
| `--names-only` | Print only plugin names, one per line |

**Examples:**
```bash
scribe list
scribe ls --json
scribe list --names-only
```

**Output (default):**
```
NAME         SOURCE                              VERSION   INSTALLED
prettier     github:usescrolls/prettier-skill    1.2.0     2024-01-15
eslint       npm:@anthropic/claude-eslint        2.0.1     2024-01-10

2 plugin(s) installed
```

**Output (JSON):**
```json
{
  "plugins": [
    {
      "name": "prettier",
      "source": {"source": "github", "repo": "usescrolls/prettier-skill"},
      "version": "1.2.0",
      "installedAt": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1
}
```

### `scribe info`

Show detailed information about an installed plugin.

```bash
scribe info <name>
```

**Examples:**
```bash
scribe info prettier
```

**Output:**
```
Name:        prettier
Source:      github:usescrolls/prettier-skill
Version:     1.2.0
Category:    formatting
Description: Code formatting skill for Claude Code
Author:      Scribe Team
Installed:   2024-01-15 10:30:00
Tags:        formatter, code-style
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
| 3 | Plugin not found |
| 4 | Source resolution failed |
| 5 | Registry/filesystem error |

## Backward Compatibility

The CLI maintains full backward compatibility with existing functionality:

1. **URL Scheme Handling** - `agenthub://` URLs continue to work:
   ```bash
   scribe "agenthub://install?name=test&source=github&repo=user/repo"
   ```

2. **GUI Mode** - Running without arguments launches the system tray:
   ```bash
   scribe
   ```

3. **Headless Mode** - The `--no-gui` flag works as before:
   ```bash
   scribe --no-gui
   ```

## Detection Order

When scribe is invoked:

1. If first argument starts with `agenthub://` → URL scheme handler
2. If first argument is a known command → CLI mode
3. If no arguments → GUI mode (or headless with `--no-gui`)

## Future Enhancements

- `scribe update [name]` - Update installed plugins
- `scribe search <query>` - Search the useScrolls registry
- Shell completions (bash, zsh, fish, powershell)
- `@scribe/name` shorthand for registry lookups
