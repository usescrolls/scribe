#Requires -Version 5.1
<#
.SYNOPSIS
    Scribe Windows Installer

.DESCRIPTION
    This script:
    1. Installs the binary to Program Files (or user's local AppData)
    2. Registers the agenthub:// URL scheme via Windows Registry

.PARAMETER UserInstall
    Install to user's AppData instead of Program Files (no admin required)

.EXAMPLE
    # Run as administrator for system-wide install
    .\install.ps1

    # Run without admin for user-only install
    .\install.ps1 -UserInstall
#>

param(
    [switch]$UserInstall
)

$ErrorActionPreference = "Stop"

$BinaryName = "scribe.exe"

# Determine install location
if ($UserInstall) {
    $InstallDir = "$env:LOCALAPPDATA\Scribe"
    $RegistryRoot = "HKCU:\Software\Classes"
} else {
    $InstallDir = "$env:ProgramFiles\Scribe"
    $RegistryRoot = "HKLM:\Software\Classes"

    # Check for admin privileges
    $currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
    if (-not $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
        Write-Host "Error: System-wide installation requires administrator privileges." -ForegroundColor Red
        Write-Host "Either run PowerShell as Administrator, or use -UserInstall for user-only installation." -ForegroundColor Yellow
        exit 1
    }
}

# Find script directory and build directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoDir = Split-Path -Parent (Split-Path -Parent $ScriptDir)
$BuildDir = Join-Path $RepoDir "build"

# Detect architecture and find binary
$Arch = if ([Environment]::Is64BitOperatingSystem) { "windows-amd64" } else { "windows-386" }
$BinaryPath = $null

$PossiblePaths = @(
    (Join-Path $BuildDir "scribe-$Arch.exe"),
    (Join-Path $BuildDir $BinaryName),
    (Join-Path $ScriptDir $BinaryName)
)

foreach ($path in $PossiblePaths) {
    if (Test-Path $path) {
        $BinaryPath = $path
        break
    }
}

if (-not $BinaryPath) {
    Write-Host "Error: Binary not found." -ForegroundColor Red
    Write-Host "Searched in:" -ForegroundColor Yellow
    $PossiblePaths | ForEach-Object { Write-Host "  $_" -ForegroundColor Yellow }
    Write-Host ""
    Write-Host "Please build first with: go build -o build\$BinaryName .\cmd\scribe" -ForegroundColor Cyan
    exit 1
}

Write-Host "Scribe Installer" -ForegroundColor Cyan
Write-Host "================" -ForegroundColor Cyan
Write-Host ""

# Create install directory
if (-not (Test-Path $InstallDir)) {
    Write-Host "Creating install directory: $InstallDir"
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Copy binary
$DestPath = Join-Path $InstallDir $BinaryName
Write-Host "Installing binary to: $DestPath"
Copy-Item -Path $BinaryPath -Destination $DestPath -Force

# Register URL scheme in registry
Write-Host ""
Write-Host "Registering URL scheme handler..."

# Create registry keys
$ProtocolKey = "$RegistryRoot\agenthub"

# Remove existing registration if present
if (Test-Path $ProtocolKey) {
    Remove-Item -Path $ProtocolKey -Recurse -Force
}

# Create protocol registration
New-Item -Path $ProtocolKey -Force | Out-Null
Set-ItemProperty -Path $ProtocolKey -Name "(Default)" -Value "URL:Scribe Protocol"
Set-ItemProperty -Path $ProtocolKey -Name "URL Protocol" -Value ""

# Create DefaultIcon key
$IconKey = "$ProtocolKey\DefaultIcon"
New-Item -Path $IconKey -Force | Out-Null
Set-ItemProperty -Path $IconKey -Name "(Default)" -Value "`"$DestPath`",0"

# Create shell\open\command key
$ShellKey = "$ProtocolKey\shell"
$OpenKey = "$ShellKey\open"
$CommandKey = "$OpenKey\command"

New-Item -Path $ShellKey -Force | Out-Null
New-Item -Path $OpenKey -Force | Out-Null
New-Item -Path $CommandKey -Force | Out-Null
Set-ItemProperty -Path $CommandKey -Name "(Default)" -Value "`"$DestPath`" `"%1`""

Write-Host "  Registry key: $ProtocolKey" -ForegroundColor Green

# Add to PATH (optional, for user install)
if ($UserInstall) {
    $UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($UserPath -notlike "*$InstallDir*") {
        Write-Host ""
        Write-Host "Adding to user PATH..."
        [Environment]::SetEnvironmentVariable("PATH", "$UserPath;$InstallDir", "User")
        Write-Host "  Added: $InstallDir" -ForegroundColor Green
    }
}

Write-Host ""
Write-Host "Installation complete!" -ForegroundColor Green
Write-Host ""
Write-Host "To start Scribe:" -ForegroundColor Cyan
Write-Host "  $DestPath"
Write-Host ""
Write-Host "To test the URL scheme:" -ForegroundColor Cyan
Write-Host "  Start-Process 'agenthub://install?name=test&source=github&repo=user/repo'"
Write-Host ""
Write-Host "To run at login, add a shortcut to:" -ForegroundColor Cyan
Write-Host "  $env:APPDATA\Microsoft\Windows\Start Menu\Programs\Startup"
Write-Host ""

# Offer to create startup shortcut
$CreateStartup = Read-Host "Create startup shortcut? (y/N)"
if ($CreateStartup -eq 'y' -or $CreateStartup -eq 'Y') {
    $StartupFolder = "$env:APPDATA\Microsoft\Windows\Start Menu\Programs\Startup"
    $ShortcutPath = Join-Path $StartupFolder "Scribe.lnk"

    $WshShell = New-Object -ComObject WScript.Shell
    $Shortcut = $WshShell.CreateShortcut($ShortcutPath)
    $Shortcut.TargetPath = $DestPath
    $Shortcut.WorkingDirectory = $InstallDir
    $Shortcut.Description = "Scribe - Plugin manager for Claude Code"
    $Shortcut.Save()

    Write-Host "Startup shortcut created: $ShortcutPath" -ForegroundColor Green
}
