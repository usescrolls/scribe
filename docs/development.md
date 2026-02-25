# Development

This guide covers building, testing, and contributing to Scribe.

## Prerequisites

**Go 1.26+**

```bash
# macOS
brew install go

# Linux / WSL
curl -LO https://go.dev/dl/go1.26.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.26.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
```

**Node.js 20+ and pnpm**

```bash
# macOS
brew install node pnpm

# Linux / WSL
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo bash -
sudo apt install -y nodejs
npm install -g pnpm
```

**Linux only — native dependencies**

```bash
# Debian/Ubuntu
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev libayatana-appindicator3-dev pkg-config
```

## Quick Start

```bash
# Install Go and frontend dependencies
make deps

# Run in development mode with hot reload
wails3 dev

# Run tests
go test ./...

# Build for current platform
make build
```

---

## Project Structure

```
scribe/
├── main.go                     # Wails app entry, bindings
├── internal/                   # Core business logic
│   ├── types.go                # Data structures (Skill, Agent, Workspace)
│   ├── agents.go               # 39 agent definitions with detection
│   ├── skills.go               # SKILL.md parsing and discovery
│   ├── installer.go            # Symlink-based installation
│   ├── workspace.go            # Workspace CRUD and switching
│   ├── meta.go                 # Sidecar .scribe-meta.json management
│   ├── storage.go              # Canonical storage paths
│   ├── fetcher.go              # Git clone and source fetching (go-git)
│   ├── source.go               # Source string parsing
│   ├── url_scheme.go           # agenthub:// URL scheme handler
│   ├── onboarding.go           # Onboarding logic (agent detection, skill import)
│   └── *_test.go               # Backend unit tests (per-file)
├── cli/                        # CLI commands
│   ├── root.go                 # Root command setup
│   ├── install.go              # Install command
│   ├── uninstall.go            # Uninstall command
│   ├── list.go                 # List command
│   ├── info.go                 # Info command
│   ├── check.go                # Check for updates
│   ├── update.go               # Update skills
│   ├── workspace.go            # Workspace commands
│   ├── cache.go                # Cache management commands
│   ├── onboarding.go           # CLI onboarding/setup wizard
│   └── *_test.go               # CLI tests (per-file)
├── frontend/                   # Vue 3 frontend
│   ├── src/
│   │   ├── App.vue             # Main layout with onboarding gate
│   │   ├── components/
│   │   │   ├── SkillList.vue           # Workspace skills list
│   │   │   ├── SkillCard.vue           # Skill display card
│   │   │   ├── BrowseSkills.vue        # All skills browser with update support
│   │   │   ├── InstallSkills.vue       # Multi-step install wizard
│   │   │   ├── WorkspaceDropdown.vue   # Workspace selector
│   │   │   ├── AgentStatusPanel.vue    # Agent status display
│   │   │   ├── OnboardingWizard.vue    # First-run onboarding
│   │   │   ├── SettingsModal.vue       # Settings panel
│   │   │   ├── ToastNotification.vue   # Notification system
│   │   │   ├── ConfirmDialog.vue       # Confirmation dialogs
│   │   │   └── onboarding/             # Onboarding step components
│   │   │       ├── WelcomeStep.vue
│   │   │       ├── AgentDetectionStep.vue
│   │   │       ├── ExistingSkillsStep.vue
│   │   │       ├── InstallDemoStep.vue
│   │   │       └── CompleteStep.vue
│   │   ├── composables/
│   │   │   ├── useSkills.ts
│   │   │   ├── useWorkspaces.ts
│   │   │   ├── useAgents.ts
│   │   │   └── useOnboarding.ts
│   │   └── types/
│   │       └── skill.ts
│   └── src/**/*.test.ts        # Frontend tests (Vitest)
├── docs/                       # Documentation
├── packaging/                  # Platform installers
└── build/                      # Build outputs
```

---

## Testing

### Local Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
make test-verbose

# Run tests with coverage
make coverage

# Generate HTML coverage report
make coverage-html
```

### Test Coverage

```bash
# HTML report
make coverage-html
open build/coverage.html
```

### Frontend Testing

```bash
cd frontend

# Run Vitest tests
pnpm test

# Run with coverage
pnpm test:coverage

# Run in watch mode
pnpm test:watch
```

---

## Building

### Development Build

```bash
# Run with Wails dev server (hot reload)
wails3 dev

# Or build once
wails3 build
```

### Production Build

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build macOS DMG installer
make dmg
```

### Building the macOS DMG Installer

```bash
# Prerequisites
brew install create-dmg

# Build the DMG (requires a binary in build/)
make app    # Creates Scribe.app bundle
make dmg    # Creates Scribe-Installer.dmg

# Or build everything at once
make release
```

The DMG will be created at `build/Scribe-Installer.dmg`.

**Note:** Building the DMG requires granting Finder automation permissions to your terminal app (System Settings > Privacy & Security > Automation).

---

## Architecture Reference

### Core Types

```go
type Skill struct {
    Name        string
    Description string
    Path        string         // Local path to SKILL.md directory
    Content     string         // Raw SKILL.md content
    Metadata    map[string]any // Additional frontmatter fields
    Meta        *SkillMeta     // Source tracking (from .scribe-meta.json)
}

type SkillMeta struct {
    Source      string `json:"source"`
    SourceType  string `json:"sourceType"`
    SourceURL   string `json:"sourceUrl"`
    SkillPath   string `json:"skillPath,omitempty"`
    ContentHash string `json:"contentHash"`
    CommitHash  string `json:"commitHash,omitempty"`
    CommitDate  string `json:"commitDate,omitempty"`
    InstalledAt string `json:"installedAt"`
    UpdatedAt   string `json:"updatedAt"`
}

type Agent struct {
    ID              string
    DisplayName     string
    GlobalSkillsDir string // e.g., "~/.claude/skills"
    GlobalConfigDir string // For detection, e.g., "~/.claude"
}

type Workspace struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Skills      []string `json:"skills"`
}
```

### URL Scheme IPC Architecture

When a user clicks an `agenthub://` link, the OS behavior differs by platform:

| Platform | App Not Running | App Already Running |
|----------|-----------------|---------------------|
| macOS | Launches app with URL as CLI arg | Sends `kAEGetURL` Apple Event |
| Linux | Launches new process with URL arg | Must forward via IPC (new process starts) |
| Windows | Launches new process with URL arg | Must forward via IPC (new process starts) |

### IPC Flow (Linux/Windows)

```mermaid
sequenceDiagram
    participant New as New Instance<br/>(from URL)
    participant Running as Running Instance<br/>(with systray)

    New->>Running: 1. Try connect to socket/pipe
    New->>Running: 2. Send URL
    Running-->>New: 3. Receive "OK"
    Note over New: 4. Exit
    Note over Running: 5. Process URL
```

### IPC Protocol

Simple newline-delimited text with acknowledgment:

```
Client → Server: agenthub://install?source=github&repo=user/repo\n
Server → Client: OK\n
```

**Timeouts:**
- Connection: 2 seconds
- Read/write: 5 seconds

**Security:**
- Linux socket permissions: `0600` (owner only)
- Windows named pipe: Default security (current user)

---

## Platform-Specific Details

### macOS

- URL scheme registered via `Info.plist` in the `.app` bundle
- Apple Events handled by Objective-C code
- CGO required for Cocoa/Objective-C integration
- Must run as `.app` bundle for URL scheme to work (not raw binary)

### Linux

- URL scheme registered via XDG desktop entry (`~/.local/share/applications/scribe.desktop`)
- IPC via Unix domain socket at `~/.scribe/ipc.sock`
- CGO required for GTK3 systray bindings
- Must run `update-desktop-database` after installing desktop entry

### Windows

- URL scheme registered in Windows Registry (`HKEY_CLASSES_ROOT\agenthub`)
- Single-instance detection via named mutex (`Global\Scribe`)
- IPC via named pipe (`\\.\pipe\Scribe`)
- No CGO required for IPC (uses `go-winio` library)
- Requires admin for system-wide install, or use `-UserInstall` for user-only

---

## Cross-Compilation

The Wails framework requires building on the target platform for GUI applications:

```bash
# Build on macOS for macOS
make build

# Build on Linux for Linux
make build

# Build on Windows for Windows
make build
```

For CI/CD, use platform-specific runners or Docker containers.

---

## Testing IPC

### macOS
```bash
# Terminal 1: Start Scribe
./build/scribe -debug

# Terminal 2: Test URL scheme
open "agenthub://install?source=github&repo=user/repo"

# Verify
ls ~/.scribe/scrolls/
```

### Linux
```bash
./build/scribe -debug &
xdg-open "agenthub://install?source=github&repo=user/repo"
```

### Windows
```powershell
# Terminal 1: Start Scribe
.\scribe.exe -debug

# Terminal 2: Test URL scheme
Start-Process "agenthub://install?source=github&repo=user/repo"

# Verify
dir $env:USERPROFILE\.scribe\scrolls\
```

---

## Common Pitfalls

| Platform | Issue | Solution |
|----------|-------|----------|
| Linux | Stale socket file | Remove on startup (`os.Remove(socketPath)`) |
| Linux | Desktop database not updated | Run `update-desktop-database` |
| Windows | Pipe naming | Must use `\\.\pipe\` prefix |
| Windows | Registry permissions | Use `-UserInstall` or run as admin |
| All | Race condition on startup | Add connection timeout, retry logic |
| All | Symlink failures | Fall back to directory copy |
