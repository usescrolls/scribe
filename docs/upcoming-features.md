# Upcoming Features

This document tracks planned features and improvements for Scribe.

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
