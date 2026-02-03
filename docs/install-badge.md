# Install Badge for GitHub READMEs

Add an "Install with Scribe" button to your repository to make it easy for users to install your skills with a single click.

## Quick Start

Add this markdown to your README:

```markdown
[![Install with Scribe](https://raw.githubusercontent.com/usescrolls/scribe/main/assets/badge/install.svg)](agenthub://install?source=github&repo=OWNER/REPO)
```

Replace `OWNER/REPO` with your GitHub username and repository name.

## Badge Styles

### Default (with gradient)
```markdown
[![Install with Scribe](https://raw.githubusercontent.com/usescrolls/scribe/main/assets/badge/install.svg)](agenthub://install?source=github&repo=OWNER/REPO)
```

### Flat
```markdown
[![Install with Scribe](https://raw.githubusercontent.com/usescrolls/scribe/main/assets/badge/install-flat.svg)](agenthub://install?source=github&repo=OWNER/REPO)
```

### Dark
```markdown
[![Install with Scribe](https://raw.githubusercontent.com/usescrolls/scribe/main/assets/badge/install-dark.svg)](agenthub://install?source=github&repo=OWNER/REPO)
```

## URL Parameters

The `agenthub://install` URL supports the following parameters:

| Parameter | Required | Description |
|-----------|----------|-------------|
| `source` | No | Source type: `github` (default), `gitlab`, or `zip` |
| `repo` | Yes | Repository path (e.g., `owner/repo`) or full URL for zip sources |
| `name` | No | Install only a specific skill by name |
| `ref` | No | Git branch, tag, or commit to install from |

### Examples

**Basic installation:**
```
agenthub://install?source=github&repo=myuser/my-skills
```

**Install from a specific branch:**
```
agenthub://install?source=github&repo=myuser/my-skills&ref=develop
```

**Install from a specific tag:**
```
agenthub://install?source=github&repo=myuser/my-skills&ref=v1.0.0
```

**Install a specific skill from a multi-skill repo:**
```
agenthub://install?source=github&repo=myuser/my-skills&name=my-awesome-skill
```

**Install from a subdirectory:**
```
agenthub://install?source=github&repo=myuser/monorepo/skills/my-skill
```

**Install from GitLab:**
```
agenthub://install?source=gitlab&repo=myuser/my-skills
```

## Complete README Example

Here's a complete example of how to add the badge to your skill repository:

```markdown
# My Awesome Skill

[![Install with Scribe](https://raw.githubusercontent.com/usescrolls/scribe/main/assets/badge/install.svg)](agenthub://install?source=github&repo=myuser/my-awesome-skill)

A skill that does amazing things for your AI coding agents.

## Installation

### One-Click Install
Click the badge above to install with [Scribe](https://github.com/usescrolls/scribe).

### Manual Install
```bash
scribe install github:myuser/my-awesome-skill
```

## Usage

After installation, the skill will be available in all your AI coding agents.
```

## How It Works

When a user clicks the badge:

1. The `agenthub://` URL scheme is handled by Scribe
2. Scribe clones your repository (shallow clone for speed)
3. Skills are discovered by looking for `SKILL.md` files
4. Skills are installed to `~/.scribe/scrolls/`
5. Symlinks are created for all detected AI agents

## Requirements

Users must have [Scribe](https://github.com/usescrolls/scribe) installed for the badge to work. The `agenthub://` URL scheme is registered during Scribe installation.

## Tips

- Keep your skill repository focused on a single skill for the cleanest install experience
- If you have multiple skills in one repo, consider using the `name` parameter to let users install specific ones
- Use the `ref` parameter to link to stable releases rather than the main branch
- Add installation instructions for users who don't have Scribe installed yet
