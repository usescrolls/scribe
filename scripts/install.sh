#!/bin/bash
#
# Scribe Installer (macOS, Linux & WSL)
#
# Downloads the latest release from a public release host and installs it.
# Override the default host with PUBLIC_DOWNLOAD_BASE or SCRIBE_DOWNLOAD_BASE.
# Sets up the background service, PATH, and URL scheme handler.
#
# Install:
#   curl -fsSL https://gitlab.com/usescrolls/scribe/-/raw/main/scripts/install.sh | bash
#
# Uninstall:
#   curl -fsSL https://gitlab.com/usescrolls/scribe/-/raw/main/scripts/uninstall.sh | bash
#

set -e

BINARY_NAME="scribe"
DESKTOP_BINARY_NAME="scribe-desktop"
INSTALL_DIR="$HOME/.local/bin"
BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"
DESKTOP_INSTALL_DIR="$HOME/.local/lib/scribe"
DESKTOP_BINARY_PATH="$DESKTOP_INSTALL_DIR/$DESKTOP_BINARY_NAME"
APP_BUNDLE="$HOME/Applications/Scribe.app"
APP_BINARY="$APP_BUNDLE/Contents/MacOS/$DESKTOP_BINARY_NAME"
OS=$(uname -s)

fetch_release_manifest() {
    curl -fsSL "$DOWNLOAD_BASE/releases/latest" 2>/dev/null || true
}

compact_release_manifest() {
    printf '%s' "$1" | tr -d '[:space:]'
}

extract_latest_version() {
    local manifest="$1"
    local tail="${manifest#*\"tag_name\":\"v}"

    if [ "$tail" = "$manifest" ]; then
        return 0
    fi

    printf '%s\n' "${tail%%\"*}"
}

extract_asset_download_url() {
    local manifest="$1"
    local asset_name="$2"
    local marker="\"name\":\"${asset_name}\",\"browser_download_url\":\""
    local tail="${manifest#*${marker}}"

    if [ "$tail" = "$manifest" ]; then
        return 0
    fi

    printf '%s\n' "${tail%%\"*}"
}

# --- Install ---

DEFAULT_DOWNLOAD_BASE="https://cdn.usescrolls.com/scribe"
DOWNLOAD_BASE="${SCRIBE_DOWNLOAD_BASE:-${PUBLIC_DOWNLOAD_BASE:-$DEFAULT_DOWNLOAD_BASE}}"
DOWNLOAD_BASE="${DOWNLOAD_BASE%/}"
MACOS_MIN_VERSION="${MACOS_MIN_VERSION:-11.0}"
ARCH=$(uname -m)

case "$OS" in
    Linux)
        case "$ARCH" in
            x86_64)
                CLI_ASSET_NAME="scribe-cli-linux-amd64"
                DESKTOP_ASSET_NAME="scribe-desktop-linux-amd64"
                ;;
            aarch64)
                CLI_ASSET_NAME="scribe-cli-linux-arm64"
                DESKTOP_ASSET_NAME="scribe-desktop-linux-arm64"
                ;;
            *)       echo "Error: unsupported architecture: $ARCH"; exit 1 ;;
        esac
        ;;
    Darwin)
        case "$ARCH" in
            arm64)
                CLI_ASSET_NAME="scribe-cli-darwin-arm64"
                DESKTOP_ASSET_NAME="scribe-desktop-darwin-arm64"
                ;;
            x86_64)
                # Detect Rosetta 2: use arm64 binary on Apple Silicon
                if [ "$(sysctl -n sysctl.proc_translated 2>/dev/null)" = "1" ]; then
                    CLI_ASSET_NAME="scribe-cli-darwin-arm64"
                    DESKTOP_ASSET_NAME="scribe-desktop-darwin-arm64"
                else
                    CLI_ASSET_NAME="scribe-cli-darwin-amd64"
                    DESKTOP_ASSET_NAME="scribe-desktop-darwin-amd64"
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

if [ "$OS" = "Darwin" ]; then
    DESKTOP_BINARY_PATH="$APP_BINARY"
else
    APP_BUNDLE=""
fi

CLI_DOWNLOAD_URL="$DOWNLOAD_BASE/$CLI_ASSET_NAME"
DESKTOP_DOWNLOAD_URL="$DOWNLOAD_BASE/$DESKTOP_ASSET_NAME"
RELEASE_MANIFEST="$(fetch_release_manifest)"
COMPACT_RELEASE_MANIFEST="$(compact_release_manifest "$RELEASE_MANIFEST")"

# Check for existing installation
INSTALLED_VERSION=""
if [ -x "$BINARY_PATH" ]; then
    INSTALLED_VERSION=$("$BINARY_PATH" version 2>/dev/null | sed 's/scribe version //' || true)
fi

# Get latest version and the exact asset URLs from release metadata.
LATEST_VERSION="$(extract_latest_version "$COMPACT_RELEASE_MANIFEST")"
CLI_MANIFEST_DOWNLOAD_URL="$(extract_asset_download_url "$COMPACT_RELEASE_MANIFEST" "$CLI_ASSET_NAME")"
DESKTOP_MANIFEST_DOWNLOAD_URL="$(extract_asset_download_url "$COMPACT_RELEASE_MANIFEST" "$DESKTOP_ASSET_NAME")"

if [ -n "$CLI_MANIFEST_DOWNLOAD_URL" ]; then
    CLI_DOWNLOAD_URL="$CLI_MANIFEST_DOWNLOAD_URL"
fi
if [ -n "$DESKTOP_MANIFEST_DOWNLOAD_URL" ]; then
    DESKTOP_DOWNLOAD_URL="$DESKTOP_MANIFEST_DOWNLOAD_URL"
fi

if [ -n "$INSTALLED_VERSION" ] && [ "$INSTALLED_VERSION" = "$LATEST_VERSION" ] && [ -x "$DESKTOP_BINARY_PATH" ]; then
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
    echo "Downloading $CLI_ASSET_NAME and $DESKTOP_ASSET_NAME..."
fi

TMP_DIR=$(mktemp -d)
CLI_TMP_FILE="$TMP_DIR/scribe"
DESKTOP_TMP_FILE="$TMP_DIR/scribe-desktop"
trap 'rm -rf "$TMP_DIR"' EXIT

if ! curl -fsSL "$CLI_DOWNLOAD_URL" -o "$CLI_TMP_FILE"; then
    echo "Error: failed to download from $CLI_DOWNLOAD_URL"
    exit 1
fi
if ! curl -fsSL "$DESKTOP_DOWNLOAD_URL" -o "$DESKTOP_TMP_FILE"; then
    echo "Error: failed to download from $DESKTOP_DOWNLOAD_URL"
    exit 1
fi
chmod +x "$CLI_TMP_FILE" "$DESKTOP_TMP_FILE"

# Install CLI and desktop binaries
mkdir -p "$INSTALL_DIR"
if [ "$OS" = "Darwin" ]; then
    # Create a minimal .app bundle so macOS registers the agenthub:// URL scheme
    mkdir -p "$APP_BUNDLE/Contents/MacOS"

    echo "Installing to $APP_BUNDLE..."
    mv "$DESKTOP_TMP_FILE" "$APP_BINARY"

    echo "Installing CLI to $BINARY_PATH..."
    rm -f "$BINARY_PATH"
    mv "$CLI_TMP_FILE" "$BINARY_PATH"
    echo "  App bundle: $APP_BUNDLE"
    echo "  CLI binary: $BINARY_PATH"
else
    mkdir -p "$DESKTOP_INSTALL_DIR"
    echo "Installing CLI to $BINARY_PATH..."
    mv "$CLI_TMP_FILE" "$BINARY_PATH"
    echo "Installing desktop app to $DESKTOP_BINARY_PATH..."
    mv "$DESKTOP_TMP_FILE" "$DESKTOP_BINARY_PATH"
    echo "  CLI binary: $BINARY_PATH"
    echo "  Desktop binary: $DESKTOP_BINARY_PATH"
fi

# Remove the legacy macOS app executable name from pre-split installs.
if [ "$OS" = "Darwin" ]; then
    rm -f "$APP_BUNDLE/Contents/MacOS/scribe"
fi

mkdir -p "$HOME/.scribe"
INSTALL_VERSION="${LATEST_VERSION:-unknown}"
cat > "$HOME/.scribe/install.json" << EOF
{
  "version": "$INSTALL_VERSION",
  "channel": "stable",
  "cliPath": "$BINARY_PATH",
  "desktopPath": "$DESKTOP_BINARY_PATH",
  "appBundlePath": "$APP_BUNDLE",
  "installedComponents": ["cli", "desktop"],
  "publicDownloadBase": "$DOWNLOAD_BASE"
}
EOF
echo "  Install manifest: $HOME/.scribe/install.json"

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
    # --- App bundle Info.plist (registers agenthub:// URL scheme) ---
    echo ""
    echo "Configuring app bundle..."
    mkdir -p "$APP_BUNDLE/Contents/MacOS"
    cat > "$APP_BUNDLE/Contents/Info.plist" << PLISTEOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDevelopmentRegion</key>
    <string>en</string>
    <key>CFBundleExecutable</key>
    <string>scribe-desktop</string>
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
    <string>${MACOS_MIN_VERSION}</string>
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
    ICON_URL="$DOWNLOAD_BASE/icons/AppIcon.icns"
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
    ICON_URL="$DOWNLOAD_BASE/icons/icon.png"
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
Exec=$DESKTOP_BINARY_PATH %u
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
Exec=$DESKTOP_BINARY_PATH
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
    if [ -n "${DISPLAY:-}${WAYLAND_DISPLAY:-}" ]; then
        # Kill any existing instance before launching
        pkill -x "$BINARY_NAME" 2>/dev/null || true
        pkill -x "$DESKTOP_BINARY_NAME" 2>/dev/null || true
        nohup "$DESKTOP_BINARY_PATH" > /dev/null 2>&1 &
    else
        echo "Desktop launch skipped: no graphical session detected."
    fi
fi
