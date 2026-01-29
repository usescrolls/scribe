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

# Clean previous DMG and any temp files
rm -f "$DMG_PATH"
rm -f "$BUILD_DIR"/rw.*.dmg

# Function to create simple DMG without fancy styling
create_simple_dmg() {
    echo "Creating simple DMG with hdiutil..."
    TEMP_DIR="$BUILD_DIR/dmg_temp"
    rm -rf "$TEMP_DIR"
    mkdir -p "$TEMP_DIR"

    # Copy app and create Applications symlink
    cp -R "$APP_PATH" "$TEMP_DIR/"
    ln -s /Applications "$TEMP_DIR/Applications"

    # Create DMG
    hdiutil create -volname "$VOLUME_NAME" -srcfolder "$TEMP_DIR" -ov -format UDZO "$DMG_PATH"

    # Cleanup
    rm -rf "$TEMP_DIR"
}

# Check if create-dmg is installed
if ! command -v create-dmg &> /dev/null; then
    echo "create-dmg not found, using simple DMG creation..."
    create_simple_dmg
else
    # Generate background if it doesn't exist
    if [ ! -f "$BACKGROUND_PATH" ]; then
        echo "Generating background image..."
        python3 "$SCRIPT_DIR/create-background.py"
    fi

    echo "Building DMG with create-dmg..."

    # Try styled DMG first, fall back to simple if AppleScript fails
    if create-dmg \
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
        "$APP_PATH" 2>&1; then
        echo "Styled DMG created successfully"
    else
        echo ""
        echo "Styled DMG failed (likely Finder automation permission issue)"
        echo "Falling back to simple DMG..."
        rm -f "$BUILD_DIR"/rw.*.dmg
        create_simple_dmg
    fi
fi

echo ""
echo "DMG created successfully: $DMG_PATH"
echo ""
echo "Size: $(du -h "$DMG_PATH" | cut -f1)"
