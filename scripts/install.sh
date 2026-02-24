#!/bin/bash
#
# Scribe Installer (macOS, Linux & WSL)
#
# Downloads the latest release from GitHub and installs it.
# Sets up the background service and URL scheme handler.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/usescrolls/scribe/main/scripts/install.sh | bash
#

set -e

REPO="usescrolls/scribe"
BINARY_NAME="scribe"
INSTALL_DIR="/usr/local/bin"
BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"

# Detect OS and architecture
OS=$(uname -s)
ARCH=$(uname -m)

case "$OS" in
    Linux)
        case "$ARCH" in
            x86_64)  ASSET_NAME="scribe-linux-amd64" ;;
            aarch64) ASSET_NAME="scribe-linux-arm64" ;;
            *)       echo "Error: unsupported architecture: $ARCH"; exit 1 ;;
        esac
        ;;
    Darwin)
        case "$ARCH" in
            arm64)  ASSET_NAME="scribe-darwin-arm64" ;;
            x86_64) ASSET_NAME="scribe-darwin-amd64" ;;
            *)      echo "Error: unsupported architecture: $ARCH"; exit 1 ;;
        esac
        ;;
    *)
        echo "Error: unsupported OS: $OS (use this script on macOS or Linux)"
        exit 1
        ;;
esac

DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/$ASSET_NAME"

echo "Scribe Installer"
echo "================"
echo ""
echo "Downloading $ASSET_NAME..."
TMP_FILE=$(mktemp)
trap 'rm -f "$TMP_FILE"' EXIT

if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE"; then
    echo "Error: failed to download from $DOWNLOAD_URL"
    exit 1
fi
chmod +x "$TMP_FILE"

# Install binary
if [ -w "$INSTALL_DIR" ]; then
    echo "Installing to $BINARY_PATH..."
    mv "$TMP_FILE" "$BINARY_PATH"
else
    echo "Installing to $INSTALL_DIR requires sudo..."
    sudo mv "$TMP_FILE" "$BINARY_PATH"
    sudo chmod +x "$BINARY_PATH"
fi
echo "  Binary installed: $BINARY_PATH"

# --- Background service ---

if [ "$OS" = "Darwin" ]; then
    PLIST="$HOME/Library/LaunchAgents/dev.scribe.plist"
    echo ""
    echo "Setting up background service (launchd)..."

    # Unload existing service if present
    launchctl bootout "gui/$(id -u)/dev.scribe" 2>/dev/null || true

    cat > "$PLIST" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>dev.scribe</string>
    <key>ProgramArguments</key>
    <array>
        <string>$BINARY_PATH</string>
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

    launchctl bootstrap "gui/$(id -u)" "$PLIST"
    echo "  Service loaded: dev.scribe"
fi

if [ "$OS" = "Linux" ]; then
    # --- URL scheme handler ---
    DESKTOP_DIR="$HOME/.local/share/applications"
    echo ""
    echo "Registering URL scheme handler..."
    mkdir -p "$DESKTOP_DIR"
    cat > "$DESKTOP_DIR/scribe.desktop" << DESKTOP
[Desktop Entry]
Version=1.0
Name=Scribe
GenericName=Skill Manager
Comment=Skill manager for coding agents
Exec=$BINARY_PATH %u
Icon=scribe
Terminal=false
Type=Application
Categories=Development;Utility;
MimeType=x-scheme-handler/agenthub;
StartupNotify=false
Keywords=claude;skill;agent;scribe;
DESKTOP

    if command -v xdg-mime &> /dev/null; then
        xdg-mime default scribe.desktop x-scheme-handler/agenthub
        echo "  Registered: x-scheme-handler/agenthub"
    fi

    if command -v update-desktop-database &> /dev/null; then
        update-desktop-database "$DESKTOP_DIR" 2>/dev/null || true
    fi

    # --- Systemd user service ---
    if command -v systemctl &> /dev/null; then
        SERVICE_DIR="$HOME/.config/systemd/user"
        echo ""
        echo "Setting up background service (systemd)..."
        mkdir -p "$SERVICE_DIR"
        cat > "$SERVICE_DIR/scribe.service" << EOF
[Unit]
Description=Scribe
After=network.target

[Service]
ExecStart=$BINARY_PATH
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
EOF

        systemctl --user daemon-reload
        systemctl --user enable scribe
        systemctl --user start scribe
        echo "  Service enabled: scribe.service"
    fi
fi

echo ""
echo "Installation complete!"
echo ""
echo "To start Scribe:"
echo "  scribe"
