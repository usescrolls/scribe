# Installation

This guide covers all installation methods for Scribe.

## Option 1: macOS DMG Installer (Recommended)

Download the DMG from [usescrolls.com/releases](https://usescrolls.com/releases), open it, and drag Scribe to your Applications folder.

## Option 2: Download Binary

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

## Option 3: Windows PowerShell Installer

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

## Option 4: Build from Source

```bash
# Clone the repo
git clone https://github.com/usescrolls/scribe.git
cd scribe

# Requires Go 1.21+
make deps
make build
./build/scribe

# For Windows cross-compilation from macOS/Linux:
make install-windows
```

## Option 5: Run with Go

```bash
go run ./cmd/scribe
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

Scribe provides a full CLI for managing plugins without needing the GUI or URL scheme.

### Commands

#### Install a plugin

```bash
scribe install <name> --github|--npm|--url|--zip <source> [flags]
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

#### Uninstall a plugin

```bash
scribe uninstall <name>
scribe uninstall --all
```

**Aliases:** `remove`, `rm`

**Examples:**
```bash
scribe uninstall prettier
scribe rm prettier
scribe uninstall --all
```

#### List installed plugins

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

#### Show plugin info

```bash
scribe info <name>
```

#### Show version

```bash
scribe version
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
| 3 | Plugin not found |
| 4 | Source resolution failed |
| 5 | Registry/filesystem error |

### GUI Mode Options

When running without CLI commands, Scribe launches in GUI mode:

```bash
./scribe [options]

Options:
  --no-gui       Run without system tray icon (headless mode)
  --debug        Enable debug logging
```
