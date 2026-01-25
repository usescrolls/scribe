#!/bin/bash
set -e

# Configuration
APP_NAME="Scribe"
DMG_NAME="Scribe-Installer"
VERSION="1.0.0"
VOLUME_NAME="Scribe"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
BUILD_DIR="$REPO_DIR/build"
APP_PATH="$BUILD_DIR/$APP_NAME.app"
DMG_PATH="$BUILD_DIR/$DMG_NAME.dmg"
BACKGROUND_PATH="$SCRIPT_DIR/dmg-background.png"

echo "Creating DMG installer..."

# Check if app bundle exists
if [ ! -d "$APP_PATH" ]; then
    echo "Error: App bundle not found at $APP_PATH"
    echo "Run './create-app.sh' first."
    exit 1
fi

# Check if create-dmg is installed
if ! command -v create-dmg &> /dev/null; then
    echo "Error: create-dmg not found. Install with: brew install create-dmg"
    exit 1
fi

# Generate background if it doesn't exist
if [ ! -f "$BACKGROUND_PATH" ]; then
    echo "Generating background image..."
    python3 "$SCRIPT_DIR/create-background.py"
fi

# Clean previous DMG
rm -f "$DMG_PATH"

echo "Building DMG with create-dmg..."

# Create the DMG using create-dmg tool
create-dmg \
    --volname "$VOLUME_NAME" \
    --volicon "$APP_PATH/Contents/Resources/AppIcon.icns" \
    --background "$BACKGROUND_PATH" \
    --window-pos 200 120 \
    --window-size 660 400 \
    --icon-size 100 \
    --icon "$APP_NAME.app" 180 170 \
    --icon "Applications" 480 170 \
    --hide-extension "$APP_NAME.app" \
    --app-drop-link 480 170 \
    "$DMG_PATH" \
    "$APP_PATH"

echo ""
echo "DMG created successfully: $DMG_PATH"
echo ""
echo "Size: $(du -h "$DMG_PATH" | cut -f1)"
