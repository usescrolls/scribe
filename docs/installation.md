# Installation

This guide covers all installation methods for Scribe.

## Option 1: Homebrew (macOS — Recommended)

```bash
brew install usescrolls/tap/scribe
```

## Option 2: macOS DMG Installer

Download the DMG from the [latest GitHub release](https://github.com/usescrolls/scribe/releases/latest), open it, and drag Scribe to your Applications folder.

## Option 3: Shell Installer (macOS, Linux & WSL)

```bash
curl -fsSL https://raw.githubusercontent.com/usescrolls/scribe/main/scripts/install.sh | bash
```

This detects your OS and architecture, downloads the latest binary to `/usr/local/bin`, sets up a background service (launchd on macOS, systemd on Linux), and on Linux also registers the `agenthub://` URL scheme.

> **Note (macOS):** The standalone binary does not support URL scheme handling (`agenthub://`). For full functionality, use the DMG installer or Homebrew which provide a proper `.app` bundle.

> **Note (Linux):** The binary requires the following runtime dependencies: `libgtk-3`, `libwebkit2gtk-4.1`, and `libayatana-appindicator3`. Install them via your package manager before running Scribe:
> ```bash
> # Debian/Ubuntu
> sudo apt install libgtk-3-0 libwebkit2gtk-4.1-0 libayatana-appindicator3-1
>
> # Fedora
> sudo dnf install gtk3 webkit2gtk4.1 libayatana-appindicator-gtk3
>
> # Arch Linux
> sudo pacman -S gtk3 webkit2gtk-4.1 libayatana-appindicator
> ```

## Option 4: Windows PowerShell Installer

For Windows, use the PowerShell installer script for full setup including URL scheme registration. The install script is available in the repository at `packaging/windows/install.ps1`:

```powershell
# Download the binary
Invoke-WebRequest -Uri https://github.com/usescrolls/scribe/releases/latest/download/scribe-windows-amd64.exe -OutFile scribe.exe

# Clone the repo to get the installer script (or download install.ps1 from the repo)
git clone https://github.com/usescrolls/scribe.git
copy scribe.exe scribe\packaging\windows\

# System-wide install (requires admin)
.\scribe\packaging\windows\install.ps1

# Or user-only install (no admin required)
.\scribe\packaging\windows\install.ps1 -UserInstall
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

The shell installer (Option 3) automatically sets up the background service on macOS (launchd) and Linux (systemd).

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

For CLI usage, see [CLI Specification](cli-spec.md).
