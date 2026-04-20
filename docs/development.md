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

**Node.js 20+ and pnpm 10.30.0+**

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

# Build the app for your current platform
make build

# Run the binary
./build/bin/scribe
```

---

## Project Structure

```
scribe/
├── main.go                     # Entry point — dual-mode (CLI + GUI)
├── internal/                   # Core business logic
│   ├── types.go                # All core data structures
│   ├── agents.go               # 39 agent definitions with detection
│   ├── skills.go               # Skill discovery, listing, querying
│   ├── installer.go            # Symlink-based installation
│   ├── workspace.go            # Workspace CRUD and switching
│   ├── storage.go              # Persistent storage (JSON-based)
│   ├── meta.go                 # Sidecar .scribe-meta.json management
│   ├── naming.go               # Skill naming conventions
│   ├── source.go               # Source string parsing (GitHub, GitLab, etc.)
│   ├── fetch.go                # Git clone, source fetching, ZIP extraction
│   ├── gitauth.go              # Git authentication (SSH, HTTPS, credentials)
│   ├── gitcache.go             # Git clone/fetch caching
│   ├── marketplace.go          # Marketplace interface
│   ├── marketplace_github.go   # GitHub marketplace search
│   ├── onboarding.go           # Onboarding wizard logic
│   ├── system_skill.go         # Built-in agent detection skill
│   ├── url_scheme.go           # agenthub:// URL scheme handler
│   ├── update_checker.go       # Check for app updates from the release manifest
│   ├── update_config.go        # App update notification configuration
│   ├── updater.go              # Perform skill updates
│   ├── self_update.go          # Self-update (binary upgrade)
│   ├── config.go               # Logger initialization
│   ├── logwriter.go            # Rotating log writer
│   └── *_test.go               # Unit tests (per-file)
├── cli/                        # CLI commands (Cobra)
│   ├── root.go                 # Root command, global flags (--debug, --json, --quiet)
│   ├── install.go              # scribe install
│   ├── uninstall.go            # scribe uninstall (aliases: remove, rm)
│   ├── list.go                 # scribe list (alias: ls)
│   ├── info.go                 # scribe info
│   ├── check.go                # scribe check
│   ├── update.go               # scribe update
│   ├── upgrade.go              # scribe upgrade (alias for update)
│   ├── version.go              # scribe version
│   ├── workspace.go            # scribe workspace subcommands
│   ├── cache.go                # scribe cache subcommands
│   ├── onboarding.go           # First-run onboarding flow
│   ├── prompt.go               # Interactive prompts and user input
│   ├── format.go               # Output formatting (tables, JSON)
│   ├── exitcodes.go            # Exit code constants
│   └── *_test.go               # CLI tests (per-file)
├── frontend/                   # Vue 3 + TypeScript + Tailwind CSS
│   ├── bindings/
│   │   └── gitlab.com/usescrolls/scribe/
│   │       ├── index.js        # Checked-in generated Wails entrypoint
│   │       ├── appservice.js   # Checked-in generated Wails service bindings
│   │       └── internal/       # Checked-in generated Wails model bindings
│   ├── src/
│   │   ├── main.ts             # Vue app entry with Wails runtime
│   │   ├── App.vue             # Main layout with onboarding gate
│   │   ├── components/
│   │   │   ├── SkillList.vue              # Workspace skills list
│   │   │   ├── SkillCard.vue              # Skill display card
│   │   │   ├── SkillDetailModal.vue       # Skill detail modal
│   │   │   ├── BrowseSkills.vue           # All skills browser
│   │   │   ├── MarketplaceSkills.vue      # Marketplace browser
│   │   │   ├── InstallSkills.vue          # Multi-step install wizard
│   │   │   ├── SidebarWorkspaceList.vue   # Workspace list in sidebar
│   │   │   ├── WorkspaceSelector.vue      # Workspace selector
│   │   │   ├── WorkspaceDropdown.vue      # Workspace dropdown menu
│   │   │   ├── WorkspaceSwitchInfoModal.vue # Workspace switch info
│   │   │   ├── AgentStatusPanel.vue       # Agent status display
│   │   │   ├── OnboardingWizard.vue       # First-run onboarding
│   │   │   ├── SettingsModal.vue          # Settings panel
│   │   │   ├── RepoReadmeModal.vue        # Repository README display
│   │   │   ├── ConfirmDialog.vue          # Confirmation dialogs
│   │   │   ├── ToastNotification.vue      # Notification system
│   │   │   ├── EmptyState.vue             # Empty state placeholder
│   │   │   ├── SourceAvatar.vue           # Source icon/avatar
│   │   │   └── onboarding/               # Onboarding step components
│   │   │       ├── WelcomeStep.vue
│   │   │       ├── AgentDetectionStep.vue
│   │   │       ├── ExistingSkillsStep.vue
│   │   │       ├── InstallDemoStep.vue
│   │   │       ├── TermsStep.vue
│   │   │       └── CompleteStep.vue
│   │   ├── composables/
│   │   │   ├── useSkills.ts
│   │   │   ├── useWorkspaces.ts
│   │   │   ├── useAgents.ts
│   │   │   ├── useOnboarding.ts
│   │   │   ├── useLogger.ts
│   │   │   ├── useUpdateChecker.ts
│   │   │   └── useSkillUpdateChecker.ts
│   │   ├── types/
│   │   │   └── skill.ts
│   │   └── bindings/
│   │       └── scribe.ts         # Auto-generated Wails bindings
│   └── **/*.test.ts              # Frontend tests (Vitest)
├── docs/                        # Documentation
├── scripts/
│   ├── install.sh               # Universal installer (macOS, Linux, WSL)
│   ├── uninstall.sh             # Universal uninstaller
│   └── hooks/                   # Git hooks (pre-commit, commit-msg)
├── icons/                       # App icons (SVG, PNG, ICO, ICNS)
├── assets/                      # Badge assets for documentation
├── .gitlab-ci.yml               # CI/CD pipeline
└── build/                       # Build outputs (generated)
```

---

## Testing

### Go Tests

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run tests with coverage
make coverage

# Generate HTML coverage report
make coverage-html
open build/coverage.html
```

### Frontend Tests

```bash
cd frontend

# Run Vitest tests
pnpm test

# Run tests once (no watch)
pnpm test:run

# Run with coverage
pnpm test:coverage
```

---

## Building

### Development Build

```bash
# Run with Wails dev server (hot reload)
make dev
```

### Production Build

```bash
# Build binary for current platform
make build

# Build and install to ~/.local/bin
make install
```

### Makefile Targets

| Target | Description |
|--------|-------------|
| `deps` | Install wails3, download Go modules, install frontend packages |
| `build` | Build frontend + Go binary for current platform |
| `build-frontend` | Build Vue frontend only (refreshes bindings if missing) |
| `dev` | Development mode with hot reload (`wails3 dev`) |
| `run` | Quick test: `go run . list` |
| `clean` | Remove build artifacts, local caches, node_modules, coverage |
| `install` | Build and install to `~/.local/bin` |
| `test` | Run Go tests |
| `test-verbose` | Run Go tests with verbose output |
| `coverage` | Run tests with coverage stats |
| `coverage-html` | Generate HTML coverage report |
| `lint` | Run `golangci-lint` |
| `lint-fix` | Run `golangci-lint` with auto-fixes |
| `wails-generate` | Force regenerate all Wails v3 bindings |
| `wails-sync-bindings` | Regenerate the checked-in frontend bindings used by CI |
| `wails-ensure-bindings` | Generate bindings only if missing |
| `install-hooks` | Install pre-commit and commit-msg git hooks |

---

## Wails Bindings

Scribe commits the generated files under `frontend/bindings/gitlab.com/usescrolls/scribe/`.

That is intentional:

- frontend tests import those bindings directly
- the GitLab frontend job does not run Wails code generation anymore
- CI stays faster and less fragile when these files are already in the repo

If you change the Wails service surface in `app_service.go`, `main.go`, or exported model types under `internal/`, refresh the tracked bindings with:

```bash
make wails-sync-bindings
```

The sync script expects the locally installed `wails3` version to match the pinned `WAILS_VERSION` in the `Makefile`. If it does not, rerun `make deps`.

If you have the repo hooks installed with `make install-hooks`, the pre-commit hook will do that automatically and stage the updated binding files for you.

---

## Architecture Reference

### Core Types

```go
type Skill struct {
    Name        string         `json:"name"`
    Description string         `json:"description"`
    Path        string         `json:"path,omitempty"`
    Content     string         `json:"content,omitempty"`
    Metadata    map[string]any `json:"metadata,omitempty"`
    Meta        *SkillMeta     `json:"meta,omitempty"`
}

type SkillMeta struct {
    Source      string `json:"source"`
    SourceType  string `json:"sourceType"`
    SourceURL   string `json:"sourceUrl,omitempty"`
    SkillPath   string `json:"skillPath,omitempty"`
    ContentHash string `json:"contentHash"`
    CommitHash  string `json:"commitHash,omitempty"`
    CommitDate  string `json:"commitDate,omitempty"`
    IsPrivate   bool   `json:"isPrivate,omitempty"`
    InstalledAt string `json:"installedAt"`
    UpdatedAt   string `json:"updatedAt"`
}

type Agent struct {
    ID              string `json:"id"`
    DisplayName     string `json:"displayName"`
    GlobalSkillsDir string `json:"globalSkillsDir"`
    GlobalConfigDir string `json:"globalConfigDir"`
}

type Workspace struct {
    Name        string   `json:"name"`
    Description string   `json:"description,omitempty"`
    Skills      []string `json:"skills"`
}

type SourceInfo struct {
    Type      string // github, gitlab, bitbucket, git, local, url, well-known, zip
    Owner     string
    Repo      string
    Ref       string
    Subpath   string
    URL       string
    LocalPath string
}

type Config struct {
    ActiveWorkspace             string `json:"activeWorkspace"`
    OnboardingCompleted         bool   `json:"onboardingCompleted"`
    TermsAcceptedAt             string `json:"termsAcceptedAt,omitempty"`
    TermsAcceptedVersion        int    `json:"termsAcceptedVersion,omitempty"`
    UpdateNotificationsDisabled bool   `json:"updateNotificationsDisabled,omitempty"`
    LastUpdateCheck             string `json:"lastUpdateCheck,omitempty"`
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

Linux and Windows release artifacts are built in GitLab CI. The macOS release
artifact is built and published locally from a tagged macOS checkout because the
GUI build requires the Apple SDK.

The macOS binary is built with a pinned minimum deployment target of `11.0` by
default so releases stay compatible across build machines. Override it with
`MACOS_MIN_VERSION` only when intentionally raising the support floor.

## Release Publishing

Tagged releases are split across CI and a local macOS publish step:

- GitLab CI runs tests, builds the Linux and Windows binaries, uploads immutable release files under `scribe/releases/<tag>/`, refreshes the moving latest aliases, and creates the GitLab release entry.
- A local macOS checkout builds `scribe-darwin-arm64` with `MACOS_MIN_VERSION=11.0` by default, uploads it under the same `scribe/releases/<tag>/` prefix, refreshes the moving macOS alias, and writes the final release manifest consumed by auto-update.

Run the macOS publish step from a clean macOS checkout where `HEAD` has exactly one `v*` tag:

```bash
./scripts/release-macos-local.sh
```

The script auto-loads `.env` from the repo root. Start from [`.env.example`](../.env.example) and set at least:

- `R2_ACCESS_KEY_ID`
- `R2_SECRET_ACCESS_KEY`
- `R2_ENDPOINT`

Optional overrides:

- `CDN_BUCKET`
- `CDN_PREFIX` (defaults to `scribe`)
- `PUBLIC_DOWNLOAD_BASE` (must end with `/scribe` when `CDN_PREFIX=scribe`)
- `GITLAB_PROJECT_URL`

The script waits for the CI-published Linux and Windows assets under `scribe/releases/<tag>/`, uploads the macOS binary under the same versioned prefix, writes `scribe/releases/latest`, and also refreshes the legacy root `releases/latest` manifest for older installed builds.

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
