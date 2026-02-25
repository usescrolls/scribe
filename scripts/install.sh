#!/bin/bash
#
# Scribe Installer (macOS, Linux & WSL)
#
# Downloads the latest release from GitHub and installs it.
# Sets up the background service, PATH, and URL scheme handler.
#
# Install:
#   curl -fsSL https://raw.githubusercontent.com/usescrolls/scribe/main/scripts/install.sh | bash
#
# Uninstall:
#   curl -fsSL https://raw.githubusercontent.com/usescrolls/scribe/main/scripts/install.sh | bash -s -- --uninstall
#

set -e

BINARY_NAME="scribe"
INSTALL_DIR="$HOME/.local/bin"
BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"
OS=$(uname -s)

# --- Uninstall ---

if [ "${1:-}" = "--uninstall" ]; then
    echo "Scribe Uninstaller"
    echo "=================="
    echo ""

    # Stop and remove background service
    if [ "$OS" = "Darwin" ]; then
        PLIST="$HOME/Library/LaunchAgents/dev.scribe.plist"
        launchctl bootout "gui/$(id -u)/dev.scribe" 2>/dev/null || true
        rm -f "$PLIST"
        echo "  Removed: launchd service"
    fi

    if [ "$OS" = "Linux" ]; then
        if command -v systemctl &> /dev/null; then
            systemctl --user stop scribe 2>/dev/null || true
            systemctl --user disable scribe 2>/dev/null || true
            rm -f "$HOME/.config/systemd/user/scribe.service"
            systemctl --user daemon-reload 2>/dev/null || true
            echo "  Removed: systemd service"
        fi

        rm -f "$HOME/.local/share/applications/scribe.desktop"
        if command -v update-desktop-database &> /dev/null; then
            update-desktop-database "$HOME/.local/share/applications" 2>/dev/null || true
        fi
        echo "  Removed: desktop entry"
    fi

    # Remove binary
    rm -f "$BINARY_PATH"
    echo "  Removed: $BINARY_PATH"

    # Remove skills and symlinks
    if [ -d "$HOME/.scribe" ]; then
        # Remove symlinks from agent directories before deleting data
        if [ -f "$BINARY_PATH" ] || command -v scribe &> /dev/null; then
            scribe uninstall --all --yes 2>/dev/null || true
        fi
        rm -rf "$HOME/.scribe"
        echo "  Removed: ~/.scribe"
    fi

    echo ""
    echo "Scribe has been uninstalled."
    echo "Note: the PATH entry in your shell rc file was left in place (harmless)."
    exit 0
fi

# --- Install ---

REPO="usescrolls/scribe"
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
mkdir -p "$INSTALL_DIR"
echo "Installing to $BINARY_PATH..."
mv "$TMP_FILE" "$BINARY_PATH"
echo "  Binary installed: $BINARY_PATH"

# --- Add to PATH if needed ---

add_to_path() {
    local rc_file="$1"
    local line='export PATH="$HOME/.local/bin:$PATH"'

    if [ -f "$rc_file" ] && grep -qF '.local/bin' "$rc_file"; then
        return
    fi

    echo "" >> "$rc_file"
    echo "# Added by Scribe installer" >> "$rc_file"
    echo "$line" >> "$rc_file"
    echo "  Updated: $rc_file"
}

if ! echo "$PATH" | tr ':' '\n' | grep -qx "$HOME/.local/bin"; then
    echo ""
    echo "Adding ~/.local/bin to PATH..."

    case "$(basename "$SHELL")" in
        zsh)  add_to_path "$HOME/.zshrc" ;;
        bash) add_to_path "$HOME/.bashrc" ;;
        *)    add_to_path "$HOME/.profile" ;;
    esac

    export PATH="$HOME/.local/bin:$PATH"
    echo "  Note: restart your shell or run 'source ~/.zshrc' (or ~/.bashrc) for PATH to take effect"
fi

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
