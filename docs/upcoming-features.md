# Upcoming Features

This document tracks planned features and improvements for Scribe.

## Recently Completed

These features have been implemented and are available in the current build:

- **Bitbucket support** - Full support for Bitbucket repositories as a skill source
- **Onboarding wizard** - GUI and CLI first-run experience with agent detection, skill import, and demo skill installation
- **Skill update button** - Update skills directly from the Browse All tab in the GUI
- **Version tracking** - Git commit hash and date tracking for installed skills
- **Git clone cache** - Local cache for faster installs, checks, and updates

---

## Planned Features

### Marketplace Browser
**Priority:** Medium

Browse skills from within Scribe without leaving the app. Should support switching between the Vercel Skills marketplace (skills.sh) and GitHub search.

### Markdown Rendering
**Priority:** Medium

Render SKILL.md content as formatted markdown when clicking on a skill in the GUI, instead of showing plain text.

### Configuration Sync
**Priority:** Medium

Sync Scribe configuration to a GitHub repository:
- Default repository name: `skills` (configurable)
- Support private or public repositories
- Optional step during onboarding, configurable later in settings
- CLI support (`scribe sync pull` / `scribe sync push`)
- Conflict-resilient pull/push mechanism

### Launch at Startup
**Priority:** Low

Use the Wails v3 `start_at_login` plugin to register Scribe as a login item. Cross-platform support via:
- macOS: AppleScript Login Items
- Linux: XDG Autostart `.desktop` file
- Windows: Registry `Run` key

Add a toggle in the tray menu and/or settings UI.

### Private Repository Support
**Priority:** Low

Support authenticated access to private GitHub/GitLab/Bitbucket repositories via personal access tokens or SSH keys.

### WSL Support
**Priority:** Low

Investigate and document whether the Windows build works correctly under WSL.

### Sponsor Section
**Priority:** Low

Add a sponsor/funding section to the project README with tiers and a "Made in Switzerland" badge.

---

## Distribution Gaps

The macOS distribution story is complete (DMG + Homebrew). Windows and Linux have known gaps:

| Gap | Impact |
|-----|--------|
| No macOS Intel (amd64) build | Intel Mac users can't install via binary download |
| No Linux arm64 build | Raspberry Pi / ARM server users can't install |
| Install scripts not published to releases | `install.ps1` and `install.sh` exist in repo but aren't distributed |
| No Windows installer (MSI) | No URL scheme registration, PATH setup, or Start Menu entry |
| No Linux packaging (AppImage/.deb) | No automatic dependency resolution |
| No package managers (Scoop/WinGet/AUR) | No auto-updates, manual download only |

### Documentation Fixes Applied

The following documentation-reality mismatches have been addressed:
- Removed non-existent macOS Intel binary download from installation.md
- Added Linux native dependency requirements (libgtk-3, libwebkit2gtk-4.1, libayatana-appindicator3)
- Fixed install.ps1 instructions to reference the repo instead of releases
- Added note about `.app` bundle requirement for macOS URL scheme support

---

## Cross-Platform Testing

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
