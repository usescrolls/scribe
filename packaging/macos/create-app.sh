#!/bin/bash
set -e

# Configuration
APP_NAME="Scribe"
BUNDLE_ID="dev.scribe"
VERSION="1.0.0"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
BUILD_DIR="$REPO_DIR/build"
APP_DIR="$BUILD_DIR/$APP_NAME.app"

# Detect architecture and find binary
ARCH=$(uname -m)
if [ "$ARCH" = "arm64" ]; then
    BINARY_SUFFIX="darwin-arm64"
else
    BINARY_SUFFIX="darwin-amd64"
fi

# Try architecture-specific binary first, then fall back to generic
if [ -f "$BUILD_DIR/bin/scribe" ]; then
    BINARY_PATH="$BUILD_DIR/bin/scribe"
elif [ -f "$BUILD_DIR/scribe-$BINARY_SUFFIX" ]; then
    BINARY_PATH="$BUILD_DIR/scribe-$BINARY_SUFFIX"
elif [ -f "$BUILD_DIR/scribe" ]; then
    BINARY_PATH="$BUILD_DIR/scribe"
else
    BINARY_PATH="$BUILD_DIR/bin/scribe"
fi

echo "Creating $APP_NAME.app bundle..."

# Check if binary exists
if [ ! -f "$BINARY_PATH" ]; then
    echo "Error: Binary not found at $BINARY_PATH"
    echo "Run 'make build-all' first to create the binaries."
    exit 1
fi

# Clean previous app bundle
rm -rf "$APP_DIR"

# Create app bundle structure
mkdir -p "$APP_DIR/Contents/MacOS"
mkdir -p "$APP_DIR/Contents/Resources"

# Copy binary
cp "$BINARY_PATH" "$APP_DIR/Contents/MacOS/scribe"
chmod +x "$APP_DIR/Contents/MacOS/scribe"

# Copy Info.plist
cp "$SCRIPT_DIR/Info.plist" "$APP_DIR/Contents/"

# Copy icon - prefer pre-generated icns from icons directory
if [ -f "$REPO_DIR/icons/AppIcon.icns" ]; then
    cp "$REPO_DIR/icons/AppIcon.icns" "$APP_DIR/Contents/Resources/"
    echo "Copied AppIcon.icns from icons directory"
elif [ -f "$SCRIPT_DIR/AppIcon.icns" ]; then
    cp "$SCRIPT_DIR/AppIcon.icns" "$APP_DIR/Contents/Resources/"
    echo "Copied AppIcon.icns from packaging directory"
elif [ -f "$REPO_DIR/icons/icon.png" ]; then
    # Fallback: Convert PNG to ICNS if pre-generated icns not available
    echo "Converting icon.png to AppIcon.icns..."
    ICONSET_DIR="$BUILD_DIR/AppIcon.iconset"
    mkdir -p "$ICONSET_DIR"

    PNG_SOURCE="$REPO_DIR/icons/icon.png"

    # Create iconset with different sizes (using sips for resizing)
    sips -z 16 16     "$PNG_SOURCE" --out "$ICONSET_DIR/icon_16x16.png" 2>/dev/null || cp "$PNG_SOURCE" "$ICONSET_DIR/icon_16x16.png"
    sips -z 32 32     "$PNG_SOURCE" --out "$ICONSET_DIR/icon_16x16@2x.png" 2>/dev/null || cp "$PNG_SOURCE" "$ICONSET_DIR/icon_16x16@2x.png"
    sips -z 32 32     "$PNG_SOURCE" --out "$ICONSET_DIR/icon_32x32.png" 2>/dev/null || cp "$PNG_SOURCE" "$ICONSET_DIR/icon_32x32.png"
    sips -z 64 64     "$PNG_SOURCE" --out "$ICONSET_DIR/icon_32x32@2x.png" 2>/dev/null || cp "$PNG_SOURCE" "$ICONSET_DIR/icon_32x32@2x.png"
    sips -z 128 128   "$PNG_SOURCE" --out "$ICONSET_DIR/icon_128x128.png" 2>/dev/null || cp "$PNG_SOURCE" "$ICONSET_DIR/icon_128x128.png"
    sips -z 256 256   "$PNG_SOURCE" --out "$ICONSET_DIR/icon_128x128@2x.png" 2>/dev/null || cp "$PNG_SOURCE" "$ICONSET_DIR/icon_128x128@2x.png"
    sips -z 256 256   "$PNG_SOURCE" --out "$ICONSET_DIR/icon_256x256.png" 2>/dev/null || cp "$PNG_SOURCE" "$ICONSET_DIR/icon_256x256.png"
    sips -z 512 512   "$PNG_SOURCE" --out "$ICONSET_DIR/icon_256x256@2x.png" 2>/dev/null || cp "$PNG_SOURCE" "$ICONSET_DIR/icon_256x256@2x.png"
    sips -z 512 512   "$PNG_SOURCE" --out "$ICONSET_DIR/icon_512x512.png" 2>/dev/null || cp "$PNG_SOURCE" "$ICONSET_DIR/icon_512x512.png"
    sips -z 1024 1024 "$PNG_SOURCE" --out "$ICONSET_DIR/icon_512x512@2x.png" 2>/dev/null || cp "$PNG_SOURCE" "$ICONSET_DIR/icon_512x512@2x.png"

    # Convert iconset to icns
    iconutil -c icns "$ICONSET_DIR" -o "$APP_DIR/Contents/Resources/AppIcon.icns" 2>/dev/null || {
        echo "Warning: Could not create .icns file. App will use default icon."
    }

    # Cleanup
    rm -rf "$ICONSET_DIR"
fi

# Create PkgInfo
echo -n "APPL????" > "$APP_DIR/Contents/PkgInfo"

echo "Created: $APP_DIR"
echo ""
echo "To test the app:"
echo "  open \"$APP_DIR\""
