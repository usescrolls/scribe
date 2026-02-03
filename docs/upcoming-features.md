# Upcoming Features

This document tracks planned features and improvements for Scribe.

## User Experience

### First Run Setup Instructions
Add onboarding flow for new users:
- Welcome screen explaining what Scribe does
- Guide to install first skill
- Show detected agents


## Source Types

### Well-Known Endpoint
**Priority:** Low

Support for `/.well-known/skills/` discovery endpoint:
```bash
scribe install https://example.com
# Fetches https://example.com/.well-known/skills/ to discover available skills
```

### Direct URL (Single SKILL.md)
**Priority:** Low

Install a single skill directly from a URL:
```bash
scribe install https://example.com/my-skill/SKILL.md
```

## Quality Improvements

### Comprehensive Error Messages
Improve error handling with:
- Clear, actionable error messages for common failures
- Suggestions for how to fix issues
- Better network error handling with retry logic

### End-to-End GUI Testing
Full workflow testing for the desktop application:
- Install skill via GUI
- Workspace switching
- Agent detection display
- Skill removal

### Cross-Platform Testing
Verify functionality across platforms:
- macOS: Symlink behavior
- Linux: Symlink permissions
- Windows: Junction creation and permissions

## Documentation

### Marketing Comparison
Add intro comparing Scribe with similar tools:
- Comparison with Vercel Skills CLI
- Unique value proposition (multi-agent support, workspaces)
- When to use Scribe vs alternatives

### Install Button for GitHub READMEs
**Status:** Implemented

Skill authors can add an "Install with Scribe" button to their repositories:

```markdown
[![Install with Scribe](https://raw.githubusercontent.com/usescrolls/scribe/main/assets/badge/install.svg)](agenthub://install?source=github&repo=owner/repo)
```

See [install-badge.md](./install-badge.md) for full documentation including:
- Badge style variants (default, flat, dark)
- URL parameters (ref, name, source)
- Example README templates 