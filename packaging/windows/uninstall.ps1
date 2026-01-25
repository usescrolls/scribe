#Requires -Version 5.1
<#
.SYNOPSIS
    AgentHub Middleware Windows Uninstaller

.DESCRIPTION
    This script:
    1. Removes the binary from Program Files or AppData
    2. Removes the agenthub:// URL scheme from Windows Registry
    3. Removes startup shortcut if present

.PARAMETER UserInstall
    Uninstall from user's AppData instead of Program Files

.EXAMPLE
    # Run as administrator for system-wide uninstall
    .\uninstall.ps1

    # Run without admin for user-only uninstall
    .\uninstall.ps1 -UserInstall
#>

param(
    [switch]$UserInstall
)

$ErrorActionPreference = "Stop"

# Determine install location
if ($UserInstall) {
    $InstallDir = "$env:LOCALAPPDATA\AgentHub"
    $RegistryRoot = "HKCU:\Software\Classes"
} else {
    $InstallDir = "$env:ProgramFiles\AgentHub"
    $RegistryRoot = "HKLM:\Software\Classes"

    # Check for admin privileges
    $currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
    if (-not $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
        Write-Host "Error: System-wide uninstallation requires administrator privileges." -ForegroundColor Red
        Write-Host "Either run PowerShell as Administrator, or use -UserInstall for user-only uninstallation." -ForegroundColor Yellow
        exit 1
    }
}

Write-Host "AgentHub Middleware Uninstaller" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# Remove URL scheme registration
$ProtocolKey = "$RegistryRoot\agenthub"
if (Test-Path $ProtocolKey) {
    Write-Host "Removing URL scheme registration..."
    Remove-Item -Path $ProtocolKey -Recurse -Force
    Write-Host "  Removed: $ProtocolKey" -ForegroundColor Green
} else {
    Write-Host "URL scheme not registered (skipping)" -ForegroundColor Yellow
}

# Remove binary and install directory
if (Test-Path $InstallDir) {
    Write-Host "Removing installation directory..."
    Remove-Item -Path $InstallDir -Recurse -Force
    Write-Host "  Removed: $InstallDir" -ForegroundColor Green
} else {
    Write-Host "Installation directory not found (skipping)" -ForegroundColor Yellow
}

# Remove from PATH (user install only)
if ($UserInstall) {
    $UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($UserPath -like "*$InstallDir*") {
        Write-Host "Removing from user PATH..."
        $NewPath = ($UserPath -split ';' | Where-Object { $_ -ne $InstallDir }) -join ';'
        [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
        Write-Host "  Removed from PATH" -ForegroundColor Green
    }
}

# Remove startup shortcut
$StartupShortcut = "$env:APPDATA\Microsoft\Windows\Start Menu\Programs\Startup\AgentHub Middleware.lnk"
if (Test-Path $StartupShortcut) {
    Write-Host "Removing startup shortcut..."
    Remove-Item -Path $StartupShortcut -Force
    Write-Host "  Removed: $StartupShortcut" -ForegroundColor Green
}

# Optionally remove user data
$DataDir = "$env:USERPROFILE\.agenthub-middleware"
if (Test-Path $DataDir) {
    Write-Host ""
    $RemoveData = Read-Host "Remove user data at $DataDir? (y/N)"
    if ($RemoveData -eq 'y' -or $RemoveData -eq 'Y') {
        Remove-Item -Path $DataDir -Recurse -Force
        Write-Host "  Removed: $DataDir" -ForegroundColor Green
    } else {
        Write-Host "  Keeping user data" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "Uninstallation complete!" -ForegroundColor Green
