#!/bin/bash
#
# Scribe Linux Installer
#
# This script:
# 1. Installs the binary to /usr/local/bin (or ~/.local/bin)
# 2. Registers the agenthub:// URL scheme via XDG desktop entry
#
# Usage: ./install.sh
#

set -e

INSTALL_DIR="/usr/local/bin"
DESKTOP_DIR="$HOME/.local/share/applications"
BINARY_NAME="scribe"

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        ARCH_SUFFIX="linux-amd64"
        ;;
    aarch64|arm64)
        ARCH_SUFFIX="linux-arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Find script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
BUILD_DIR="$REPO_DIR/build"

# Find binary
if [ -f "$BUILD_DIR/$BINARY_NAME-$ARCH_SUFFIX" ]; then
    BINARY_PATH="$BUILD_DIR/$BINARY_NAME-$ARCH_SUFFIX"
elif [ -f "$BUILD_DIR/$BINARY_NAME" ]; then
    BINARY_PATH="$BUILD_DIR/$BINARY_NAME"
else
    echo "Error: Binary not found at $BUILD_DIR"
    echo "Please build first with: make build"
    exit 1
fi

echo "Scribe Installer"
echo "================"
echo ""

# Install binary
if [ -w "$INSTALL_DIR" ]; then
    echo "Installing binary to $INSTALL_DIR/$BINARY_NAME..."
    cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    echo "Installing to $INSTALL_DIR requires sudo..."
    sudo cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi
echo "  Binary installed: $INSTALL_DIR/$BINARY_NAME"

# Install desktop entry
echo ""
echo "Installing URL scheme handler..."
mkdir -p "$DESKTOP_DIR"
cp "$SCRIPT_DIR/scribe.desktop" "$DESKTOP_DIR/"
echo "  Desktop entry: $DESKTOP_DIR/scribe.desktop"

# Register URL scheme
xdg-mime default scribe.desktop x-scheme-handler/agenthub
echo "  Registered: x-scheme-handler/agenthub"

# Update desktop database
if command -v update-desktop-database &> /dev/null; then
    update-desktop-database "$DESKTOP_DIR" 2>/dev/null || true
fi

echo ""
echo "Installation complete!"
echo ""
echo "To start Scribe:"
echo "  $BINARY_NAME"
echo ""
echo "To test the URL scheme:"
echo "  xdg-open 'agenthub://install?name=test&source=github&repo=user/repo'"
echo ""
echo "To run at login, add to your startup applications or create a systemd user service."
