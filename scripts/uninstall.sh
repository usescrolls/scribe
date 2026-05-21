#!/bin/bash
#
# Scribe Uninstaller (macOS, Linux & WSL)
#
# Usage:
#   curl -fsSL https://gitlab.com/usescrolls/scribe/-/raw/main/scripts/uninstall.sh | bash
#

set -e

BINARY_NAME="scribe"
DESKTOP_BINARY_NAME="scribe-desktop"
INSTALL_DIR="$HOME/.local/bin"
BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"
DESKTOP_BINARY_PATH="$HOME/.local/lib/scribe/$DESKTOP_BINARY_NAME"
OS=$(uname -s)

echo "Scribe Uninstaller"
echo "=================="
echo ""

# Remove skills and symlinks first (while binary still exists)
if [ -d "$HOME/.scribe" ]; then
    if command -v scribe &> /dev/null; then
        scribe uninstall --all --yes 2>/dev/null || true
    fi

    printf "Would you like to remove the Scribe config folder (~/.scribe)? [y/N] "
    read -r answer < /dev/tty
    case "$answer" in
        [yY]|[yY][eE][sS])
            rm -rf "$HOME/.scribe"
            echo "  Removed: ~/.scribe"
            ;;
        *)
            echo "  Kept: ~/.scribe"
            ;;
    esac
fi

# Stop and remove background service
if [ "$OS" = "Darwin" ]; then
    PLIST="$HOME/Library/LaunchAgents/dev.scribe.plist"
    launchctl bootout "gui/$(id -u)/dev.scribe" 2>/dev/null || true
    rm -f "$PLIST"
    echo "  Removed: launchd service"

    # Remove app bundle and unregister URL scheme
    APP_BUNDLE="$HOME/Applications/Scribe.app"
    if [ -d "$APP_BUNDLE" ]; then
        /System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister -u "$APP_BUNDLE" 2>/dev/null || true
        rm -rf "$APP_BUNDLE"
        echo "  Removed: $APP_BUNDLE"
    fi
fi

if [ "$OS" = "Linux" ]; then
    # Clean up legacy systemd service if present
    if command -v systemctl &> /dev/null; then
        systemctl --user stop scribe 2>/dev/null || true
        systemctl --user disable scribe 2>/dev/null || true
        rm -f "$HOME/.config/systemd/user/scribe.service"
        systemctl --user daemon-reload 2>/dev/null || true
    fi

    rm -f "$HOME/.config/autostart/scribe.desktop"
    echo "  Removed: autostart entry"

    rm -f "$HOME/.local/share/applications/scribe.desktop"
    if command -v update-desktop-database &> /dev/null; then
        update-desktop-database "$HOME/.local/share/applications" 2>/dev/null || true
    fi
    echo "  Removed: desktop entry"

    rm -f "$HOME/.local/share/icons/hicolor/256x256/apps/scribe.png"
fi

# Remove CLI binary / symlink
rm -f "$BINARY_PATH"
echo "  Removed: $BINARY_PATH"

rm -f "$DESKTOP_BINARY_PATH"
rmdir "$HOME/.local/lib/scribe" 2>/dev/null || true
rm -f "$HOME/.scribe/install.json"
echo "  Removed: desktop binary and install manifest"

echo ""
echo "Scribe has been uninstalled."
echo "Note: the PATH entry in your shell rc file was left in place (harmless)."
