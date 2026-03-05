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
#   curl -fsSL https://raw.githubusercontent.com/usescrolls/scribe/main/scripts/uninstall.sh | bash
#

set -e

BINARY_NAME="scribe"
INSTALL_DIR="$HOME/.local/bin"
BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"
OS=$(uname -s)

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
            x86_64)
                # Detect Rosetta 2: use arm64 binary on Apple Silicon
                if [ "$(sysctl -n sysctl.proc_translated 2>/dev/null)" = "1" ]; then
                    ASSET_NAME="scribe-darwin-arm64"
                else
                    ASSET_NAME="scribe-darwin-amd64"
                fi
                ;;
            *)      echo "Error: unsupported architecture: $ARCH"; exit 1 ;;
        esac
        ;;
    *)
        echo "Error: unsupported OS: $OS (use this script on macOS or Linux)"
        exit 1
        ;;
esac

DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/$ASSET_NAME"

# Check for existing installation
INSTALLED_VERSION=""
if [ -x "$BINARY_PATH" ]; then
    INSTALLED_VERSION=$("$BINARY_PATH" version 2>/dev/null | sed 's/scribe version //' || true)
fi

# Get latest version tag from GitHub
LATEST_VERSION=$(curl -fsSL -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest" | sed 's|.*/v||')

if [ -n "$INSTALLED_VERSION" ] && [ "$INSTALLED_VERSION" = "$LATEST_VERSION" ]; then
    echo "Scribe v$INSTALLED_VERSION is already installed and up to date."
    exit 0
elif [ -n "$INSTALLED_VERSION" ]; then
    echo "Scribe Installer"
    echo "================"
    echo ""
    echo "Updating Scribe v$INSTALLED_VERSION → v$LATEST_VERSION..."
else
    echo "Scribe Installer"
    echo "================"
    echo ""
    echo "Downloading $ASSET_NAME..."
fi

TMP_FILE=$(mktemp)
trap 'rm -f "$TMP_FILE"' EXIT

if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE"; then
    echo "Error: failed to download from $DOWNLOAD_URL"
    exit 1
fi
chmod +x "$TMP_FILE"

# Install binary
if [ "$OS" = "Darwin" ]; then
    # Create a minimal .app bundle so macOS registers the agenthub:// URL scheme
    APP_BUNDLE="$HOME/Applications/Scribe.app"
    APP_BINARY="$APP_BUNDLE/Contents/MacOS/scribe"
    mkdir -p "$APP_BUNDLE/Contents/MacOS"

    echo "Installing to $APP_BUNDLE..."
    mv "$TMP_FILE" "$APP_BINARY"

    # Symlink CLI binary into PATH
    mkdir -p "$INSTALL_DIR"
    ln -sf "$APP_BINARY" "$BINARY_PATH"
    echo "  App bundle: $APP_BUNDLE"
    echo "  CLI symlink: $BINARY_PATH"
else
    mkdir -p "$INSTALL_DIR"
    echo "Installing to $BINARY_PATH..."
    mv "$TMP_FILE" "$BINARY_PATH"
    echo "  Binary installed: $BINARY_PATH"
fi

# Migrate: if macOS binary is not in .app bundle yet, move it there
if [ "$OS" = "Darwin" ]; then
    APP_BUNDLE="$HOME/Applications/Scribe.app"
    APP_BINARY="$APP_BUNDLE/Contents/MacOS/scribe"
    if [ -x "$BINARY_PATH" ] && [ ! -L "$BINARY_PATH" ] && [ ! -x "$APP_BINARY" ]; then
        echo "  Migrating binary into app bundle..."
        mkdir -p "$APP_BUNDLE/Contents/MacOS"
        mv "$BINARY_PATH" "$APP_BINARY"
        ln -sf "$APP_BINARY" "$BINARY_PATH"
    fi
fi

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
    APP_BUNDLE="$HOME/Applications/Scribe.app"
    APP_BINARY="$APP_BUNDLE/Contents/MacOS/scribe"

    # --- App bundle Info.plist (registers agenthub:// URL scheme) ---
    echo ""
    echo "Configuring app bundle..."
    mkdir -p "$APP_BUNDLE/Contents/MacOS"
    cat > "$APP_BUNDLE/Contents/Info.plist" << 'PLISTEOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDevelopmentRegion</key>
    <string>en</string>
    <key>CFBundleExecutable</key>
    <string>scribe</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>CFBundleIdentifier</key>
    <string>dev.scribe</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundleName</key>
    <string>Scribe</string>
    <key>CFBundleDisplayName</key>
    <string>Scribe</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleVersion</key>
    <string>1</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.13</string>
    <key>LSUIElement</key>
    <true/>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSSupportsAutomaticGraphicsSwitching</key>
    <true/>
    <key>LSApplicationCategoryType</key>
    <string>public.app-category.developer-tools</string>
    <key>CFBundleURLTypes</key>
    <array>
        <dict>
            <key>CFBundleURLName</key>
            <string>Scribe Plugin Install</string>
            <key>CFBundleURLSchemes</key>
            <array>
                <string>agenthub</string>
            </array>
        </dict>
    </array>
</dict>
</plist>
PLISTEOF

    # Download app icon
    ICON_URL="https://raw.githubusercontent.com/$REPO/main/icons/AppIcon.icns"
    mkdir -p "$APP_BUNDLE/Contents/Resources"
    curl -fsSL "$ICON_URL" -o "$APP_BUNDLE/Contents/Resources/AppIcon.icns" 2>/dev/null || true

    # Register with Launch Services so macOS knows about agenthub://
    /System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister -f "$APP_BUNDLE" 2>/dev/null || true
    echo "  Registered: agenthub:// URL scheme"

    # --- Background service (launchd) ---
    PLIST="$HOME/Library/LaunchAgents/dev.scribe.plist"
    echo ""
    echo "Setting up background service (launchd)..."

    # Unload existing service if present
    launchctl bootout "gui/$(id -u)/dev.scribe" 2>/dev/null || true

    mkdir -p "$HOME/Library/LaunchAgents"
    cat > "$PLIST" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>dev.scribe</string>
    <key>ProgramArguments</key>
    <array>
        <string>$APP_BINARY</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <dict>
        <key>SuccessfulExit</key>
        <false/>
    </dict>
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
    # --- App icon ---
    ICON_DIR="$HOME/.local/share/icons/hicolor/256x256/apps"
    ICON_URL="https://raw.githubusercontent.com/$REPO/main/icons/icon.png"
    mkdir -p "$ICON_DIR"
    curl -fsSL "$ICON_URL" -o "$ICON_DIR/scribe.png" 2>/dev/null || true

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

    # --- Autostart on login (XDG autostart) ---
    AUTOSTART_DIR="$HOME/.config/autostart"
    echo ""
    echo "Setting up autostart..."
    mkdir -p "$AUTOSTART_DIR"
    cat > "$AUTOSTART_DIR/scribe.desktop" << AUTOSTART
[Desktop Entry]
Version=1.0
Name=Scribe
Comment=Skill manager for coding agents
Exec=$BINARY_PATH
Icon=scribe
Terminal=false
Type=Application
X-GNOME-Autostart-enabled=true
AUTOSTART

    echo "  Autostart entry: $AUTOSTART_DIR/scribe.desktop"
fi

echo ""
echo "Installation complete!"

# Open the Scribe window
if [ "$OS" = "Darwin" ]; then
    # launchd already started the service; trigger the URL scheme to show the window
    sleep 1
    open "agenthub://show" 2>/dev/null || true
elif [ "$OS" = "Linux" ]; then
    # Kill any existing instance before launching
    pkill -x "$BINARY_NAME" 2>/dev/null || true
    nohup "$BINARY_PATH" > /dev/null 2>&1 &
fi
