# Installation

This guide covers all installation methods for Scribe.

## Option 1: Shell Installer (macOS, Linux & WSL)

```bash
curl -fsSL https://gitlab.com/usescrolls/scribe/-/raw/main/scripts/install.sh | bash
```

This detects your OS and architecture, downloads the latest binary, adds it to your PATH if needed, creates a minimal `.app` bundle on macOS, sets up a background service (launchd on macOS, XDG autostart on Linux), and on Linux also registers the `agenthub://` URL scheme.

If you publish releases somewhere other than the default CDN, override the base URL when running the installer:

```bash
curl -fsSL https://gitlab.com/usescrolls/scribe/-/raw/main/scripts/install.sh | PUBLIC_DOWNLOAD_BASE="https://downloads.example.com/scribe" bash
```

To uninstall:

```bash
curl -fsSL https://gitlab.com/usescrolls/scribe/-/raw/main/scripts/uninstall.sh | bash
```

This stops the background service, removes the binary, cleans up `~/.scribe`, and removes platform-specific registrations.

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

## Option 2: Windows

Download the binary directly:

```powershell
Invoke-WebRequest -Uri <PUBLIC_DOWNLOAD_BASE>/scribe-windows-amd64.exe -OutFile "$env:LOCALAPPDATA\Scribe\scribe.exe"
```

Replace `<PUBLIC_DOWNLOAD_BASE>` with your public release host. The default is `https://cdn.usescrolls.com/scribe`.

Then add `%LOCALAPPDATA%\Scribe` to your PATH.

## Option 3: Build from Source

Requires Go 1.26+, Node.js 20+, pnpm, and Wails v3. See [Development](development.md) for full setup instructions.

```bash
git clone https://gitlab.com/usescrolls/scribe.git
cd scribe
make deps
make build
./build/bin/scribe
```

---

## Running as a Background Service

The shell installer (Option 1) automatically sets up the background service on macOS (launchd) and Linux (XDG autostart).

### Windows (Startup Folder)

Windows installation is currently manual. To start Scribe automatically on login, create a Startup shortcut:

```powershell
# Create a shortcut in the Startup folder
$WshShell = New-Object -ComObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut("$env:APPDATA\Microsoft\Windows\Start Menu\Programs\Startup\Scribe.lnk")
$Shortcut.TargetPath = "$env:LOCALAPPDATA\Scribe\scribe.exe"
$Shortcut.Save()
```

Or use Task Scheduler for more control:

```powershell
# Create a scheduled task to run at login
$Action = New-ScheduledTaskAction -Execute "$env:LOCALAPPDATA\Scribe\scribe.exe"
$Trigger = New-ScheduledTaskTrigger -AtLogon
$Settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries
Register-ScheduledTask -TaskName "Scribe" -Action $Action -Trigger $Trigger -Settings $Settings
```

---

For CLI usage, see [CLI Specification](cli-spec.md).
