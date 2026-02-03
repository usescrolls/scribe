# Upcoming Features

This document tracks planned features and improvements for Scribe.

## User Experience

### First Run Setup Instructions
Add onboarding flow for new users:
- Welcome screen explaining what Scribe does
- Guide to install first skill
- Show detected agents

### Workspace Selector in Sidebar
Move workspace selection from header to left sidebar for better UX:
- Workspace list always visible
- Quick switching without dropdown
- Visual indicator of active workspace

### Global Hotkeys for Workspace Switching
Add keyboard shortcuts to switch workspaces system-wide:
- Configurable hotkey bindings
- Quick switch between recent workspaces

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

## Release & Distribution

### Release Checklist
- Version bump in `wails.json`
- Update changelog
- Build artifacts for all platforms
- Tag release in git

### Homebrew Formula Update
Update the Homebrew tap formula for new releases:
- `usescrolls/tap/scribe`

### GitHub Release Workflow
Automate release artifact publishing:
- Build binaries for macOS (arm64, amd64), Linux, Windows
- Create DMG installer for macOS
- Publish to GitHub Releases

### Code Coverage Badges
Add coverage badges to README:
- Backend coverage (Go)
- Frontend coverage (Vitest)
- Integration with codecov or similar

## Developer Experience

### Code Formatter and Linter (Go)
Add consistent code formatting and linting:
- `gofmt` / `goimports` for formatting
- `golangci-lint` for linting
- Editor integration configs

### Oxy Tools and Pre-commit Hooks
Configure development tooling:
- Pre-commit hooks for formatting and linting
- Automated checks before commit
- CI integration

## Documentation

### Marketing Comparison
Add intro comparing Scribe with similar tools:
- Comparison with Vercel Skills CLI
- Unique value proposition (multi-agent support, workspaces)
- When to use Scribe vs alternatives

### Install Button for GitHub READMEs
Showcase how skill authors can add an "Install with Scribe" button to their repositories:

```markdown
[![Install with Scribe](https://usescrolls.com/badge/install.svg)](agenthub://install?source=github&repo=owner/repo)
```

Features to implement:
- Badge SVG hosted on usescrolls.com
- Documentation for skill authors
- Example README templates
- Support for branch/tag specification in URL
