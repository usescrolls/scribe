#!/bin/bash
#
# Generate platform-specific icons from the source PNG
#
# Requires: sips, iconutil (macOS), optionally png2ico for Windows
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
SOURCE_PNG="$SCRIPT_DIR/icon.png"
ICONS_DIR="$SCRIPT_DIR"

if [ ! -f "$SOURCE_PNG" ]; then
    echo "Error: Source icon not found at $SOURCE_PNG"
    exit 1
fi

echo "Generating icons from $SOURCE_PNG..."

# ============================================================================
# macOS .icns generation
# ============================================================================

echo ""
echo "Creating macOS .icns..."

ICONSET_DIR="$ICONS_DIR/AppIcon.iconset"
mkdir -p "$ICONSET_DIR"

# Create iconset with all required sizes
# macOS requires specific sizes for proper display at different resolutions
sips -z 16 16     "$SOURCE_PNG" --out "$ICONSET_DIR/icon_16x16.png"      2>/dev/null
sips -z 32 32     "$SOURCE_PNG" --out "$ICONSET_DIR/icon_16x16@2x.png"   2>/dev/null
sips -z 32 32     "$SOURCE_PNG" --out "$ICONSET_DIR/icon_32x32.png"      2>/dev/null
sips -z 64 64     "$SOURCE_PNG" --out "$ICONSET_DIR/icon_32x32@2x.png"   2>/dev/null
sips -z 128 128   "$SOURCE_PNG" --out "$ICONSET_DIR/icon_128x128.png"    2>/dev/null
sips -z 256 256   "$SOURCE_PNG" --out "$ICONSET_DIR/icon_128x128@2x.png" 2>/dev/null
sips -z 256 256   "$SOURCE_PNG" --out "$ICONSET_DIR/icon_256x256.png"    2>/dev/null
sips -z 512 512   "$SOURCE_PNG" --out "$ICONSET_DIR/icon_256x256@2x.png" 2>/dev/null
sips -z 512 512   "$SOURCE_PNG" --out "$ICONSET_DIR/icon_512x512.png"    2>/dev/null
# For 512@2x, we need 1024x1024 but our source is only 256x256
# We'll use the best we have (256 upscaled isn't great but works)
sips -z 1024 1024 "$SOURCE_PNG" --out "$ICONSET_DIR/icon_512x512@2x.png" 2>/dev/null

# Convert iconset to icns
iconutil -c icns "$ICONSET_DIR" -o "$ICONS_DIR/AppIcon.icns"

# Cleanup iconset directory
rm -rf "$ICONSET_DIR"

echo "  Created: $ICONS_DIR/AppIcon.icns"

# ============================================================================
# Windows .ico generation
# ============================================================================

echo ""
echo "Creating Windows .ico..."

# Check for png2ico or ImageMagick
if command -v png2ico &> /dev/null; then
    # Create temporary resized PNGs for .ico
    ICO_TEMP="$ICONS_DIR/ico_temp"
    mkdir -p "$ICO_TEMP"
    sips -z 16 16   "$SOURCE_PNG" --out "$ICO_TEMP/16.png"  2>/dev/null
    sips -z 32 32   "$SOURCE_PNG" --out "$ICO_TEMP/32.png"  2>/dev/null
    sips -z 48 48   "$SOURCE_PNG" --out "$ICO_TEMP/48.png"  2>/dev/null
    sips -z 256 256 "$SOURCE_PNG" --out "$ICO_TEMP/256.png" 2>/dev/null

    png2ico "$ICONS_DIR/scribe.ico" "$ICO_TEMP/16.png" "$ICO_TEMP/32.png" "$ICO_TEMP/48.png" "$ICO_TEMP/256.png"
    rm -rf "$ICO_TEMP"
    echo "  Created: $ICONS_DIR/scribe.ico"
elif command -v magick &> /dev/null || command -v convert &> /dev/null; then
    MAGICK_CMD=$(command -v magick || command -v convert)
    "$MAGICK_CMD" "$SOURCE_PNG" -define icon:auto-resize=256,128,64,48,32,16 "$ICONS_DIR/scribe.ico"
    echo "  Created: $ICONS_DIR/scribe.ico"
else
    echo "  Warning: Neither png2ico nor ImageMagick found"
    echo "  Creating .ico using Python fallback..."

    # Python fallback for creating .ico
    ICONS_DIR="$ICONS_DIR" SOURCE_PNG="$SOURCE_PNG" python3 - <<'PYTHON_SCRIPT'
import sys
import struct
import zlib
from pathlib import Path

def create_ico(png_path, ico_path, sizes=[16, 32, 48, 256]):
    """Create a basic .ico file from a PNG."""
    import subprocess
    import tempfile
    import os

    icons_dir = Path(ico_path).parent
    temp_pngs = []

    # Create resized PNGs using sips
    for size in sizes:
        temp_png = icons_dir / f"_temp_{size}.png"
        subprocess.run(['sips', '-z', str(size), str(size), str(png_path), '--out', str(temp_png)],
                      capture_output=True)
        temp_pngs.append((size, temp_png))

    # Read PNG data
    images = []
    for size, png_file in temp_pngs:
        with open(png_file, 'rb') as f:
            images.append((size, f.read()))
        png_file.unlink()  # Delete temp file

    # Build .ico file
    # ICO Header: 2 bytes reserved, 2 bytes type (1=icon), 2 bytes count
    num_images = len(images)
    header = struct.pack('<HHH', 0, 1, num_images)

    # Calculate offsets
    entries = []
    offset = 6 + num_images * 16  # Header + directory entries

    for size, data in images:
        w = size if size < 256 else 0  # 0 means 256
        h = size if size < 256 else 0
        entry = struct.pack('<BBBBHHII',
            w, h, 0, 0,  # width, height, palette, reserved
            1, 32,       # planes, bpp
            len(data),   # size of image data
            offset)      # offset
        entries.append(entry)
        offset += len(data)

    # Write .ico file
    with open(ico_path, 'wb') as f:
        f.write(header)
        for entry in entries:
            f.write(entry)
        for size, data in images:
            f.write(data)

    print(f"  Created: {ico_path}")

import os
script_dir = Path(os.environ.get('ICONS_DIR', '/Users/littlebrat/projects/usescrolls/scribe/icons'))
source_png = Path(os.environ.get('SOURCE_PNG', script_dir / 'icon.png'))
create_ico(source_png, script_dir / 'scribe.ico')
PYTHON_SCRIPT
fi

# ============================================================================
# Linux PNG icons (various sizes)
# ============================================================================

echo ""
echo "Creating Linux PNG icons..."

LINUX_DIR="$ICONS_DIR/linux"
mkdir -p "$LINUX_DIR"

# Create common icon sizes for Linux
for SIZE in 16 24 32 48 64 128 256; do
    sips -z $SIZE $SIZE "$SOURCE_PNG" --out "$LINUX_DIR/scribe-${SIZE}x${SIZE}.png" 2>/dev/null
done

# Also copy a scalable version (original)
cp "$SOURCE_PNG" "$LINUX_DIR/scribe.png"

echo "  Created: $LINUX_DIR/scribe-*.png"

# ============================================================================
# Summary
# ============================================================================

echo ""
echo "Icon generation complete!"
echo ""
echo "Files created:"
ls -la "$ICONS_DIR"/*.icns "$ICONS_DIR"/*.ico "$LINUX_DIR"/*.png 2>/dev/null || true
