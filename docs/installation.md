# Installation

This guide covers all installation methods for Scribe.

## Option 1: macOS DMG Installer (Recommended)

Download the DMG from [usescrolls.com/releases](https://usescrolls.com/releases), open it, and drag Scribe to your Applications folder.

## Option 2: Homebrew (macOS)

```bash
brew install usescrolls/tap/scribe
```

## Option 3: Download Binary

```bash
# macOS (Apple Silicon)
curl -fsSL https://usescrolls.com/releases/scribe-darwin-arm64 -o scribe
chmod +x scribe
./scribe

# macOS (Intel)
curl -fsSL https://usescrolls.com/releases/scribe-darwin-amd64 -o scribe
chmod +x scribe
./scribe

# Linux
curl -fsSL https://usescrolls.com/releases/scribe-linux-amd64 -o scribe
chmod +x scribe
./scribe

# Windows (PowerShell)
Invoke-WebRequest -Uri https://usescrolls.com/releases/scribe-windows-amd64.exe -OutFile scribe.exe
.\scribe.exe
```

## Option 4: Windows PowerShell Installer

For Windows, use the PowerShell installer script for full setup including URL scheme registration:

```powershell
# Download the binary and installer
Invoke-WebRequest -Uri https://usescrolls.com/releases/scribe-windows-amd64.exe -OutFile scribe.exe
Invoke-WebRequest -Uri https://usescrolls.com/releases/install.ps1 -OutFile install.ps1

# System-wide install (requires admin)
.\install.ps1

# Or user-only install (no admin required)
.\install.ps1 -UserInstall
```

The installer:
- Copies the binary to `Program Files\Scribe` (or `%LOCALAPPDATA%\Scribe` for user install)
- Registers the `agenthub://` URL scheme in the Windows Registry
- Optionally creates a startup shortcut

## Option 5: Build from Source

```bash
# Clone the repo
git clone https://github.com/usescrolls/scribe.git
cd scribe

# Requires Go 1.25+ and Wails v3
make deps
make build
./build/scribe
```

---

## Running as a Background Service

### macOS (launchd)

```bash
# Create launch agent
cat > ~/Library/LaunchAgents/dev.scribe.plist << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>dev.scribe</string>
    <key>ProgramArguments</key>
    <array>
        <string>/Users/YOU/.local/bin/scribe</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/scribe.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/scribe.log</string>
</dict>
</plist>
EOF

# Load the service
launchctl load ~/Library/LaunchAgents/dev.scribe.plist
```

### Linux (systemd)

```bash
# Create user service
mkdir -p ~/.config/systemd/user

cat > ~/.config/systemd/user/scribe.service << 'EOF'
[Unit]
Description=Scribe
After=network.target

[Service]
ExecStart=%h/.local/bin/scribe
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
EOF

# Enable and start
systemctl --user daemon-reload
systemctl --user enable scribe
systemctl --user start scribe
```

### Windows (Startup Folder)

The PowerShell installer can create a startup shortcut automatically. To do it manually:

```powershell
# Create a shortcut in the Startup folder
$WshShell = New-Object -ComObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut("$env:APPDATA\Microsoft\Windows\Start Menu\Programs\Startup\Scribe.lnk")
$Shortcut.TargetPath = "$env:ProgramFiles\Scribe\scribe.exe"
$Shortcut.Save()
```

Or use Task Scheduler for more control:

```powershell
# Create a scheduled task to run at login
$Action = New-ScheduledTaskAction -Execute "$env:ProgramFiles\Scribe\scribe.exe"
$Trigger = New-ScheduledTaskTrigger -AtLogon
$Settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries
Register-ScheduledTask -TaskName "Scribe" -Action $Action -Trigger $Trigger -Settings $Settings
```

---

## Command Line Interface

Scribe provides a full CLI for managing skills.

### Install Skills

```bash
scribe install <source> [flags]
```

**Sources:**
```bash
scribe install owner/repo                    # GitHub shorthand
scribe install https://github.com/owner/repo # Full GitHub URL
scribe install ./local/path                  # Local directory
scribe install https://example.com/skills.zip # Zip URL
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--agent` | `-a` | Target specific agents (comma-separated) |
| `--skill` | `-s` | Select specific skills to install |
| `--list` | `-l` | List available skills without installing |
| `--yes` | `-y` | Skip interactive prompts |

### Uninstall Skills

```bash
scribe uninstall <skill-name>
scribe rm <skill-name>
scribe uninstall --all
```

### List Installed Skills

```bash
scribe list
scribe ls --json
scribe list --names-only
```

### Show Skill Info

```bash
scribe info <skill-name>
```

### Check for Updates

```bash
scribe check
```

### Update Skills

```bash
scribe update              # Update all
scribe update <skill-name> # Update specific skill
```

### Workspace Commands

```bash
scribe workspace list              # List workspaces
scribe workspace create <name>     # Create workspace
scribe workspace use <name>        # Switch workspace
scribe workspace add <skill>       # Add skill to workspace
scribe workspace remove <skill>    # Remove skill from workspace
scribe workspace current           # Show active workspace
scribe workspace delete <name>     # Delete workspace
```

### Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--debug` | | Enable debug logging |
| `--json` | | Output in JSON format (where applicable) |
| `--quiet` | `-q` | Suppress non-essential output |
| `--help` | `-h` | Show help |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid usage / bad arguments |
| 3 | Skill not found |
| 4 | Source resolution failed |
| 5 | Filesystem error |

### GUI Mode Options

When running without CLI commands, Scribe launches in GUI mode:

```bash
./scribe [options]

Options:
  --no-gui       Run without system tray icon (headless mode)
  --debug        Enable debug logging
```
