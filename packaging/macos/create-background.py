#!/usr/bin/env python3
"""Generate a DMG background image with drag-to-install arrow."""

import os
import struct
import zlib

def create_png(width, height, pixels):
    """Create a PNG file from RGBA pixel data."""
    def png_chunk(chunk_type, data):
        chunk_len = struct.pack('>I', len(data))
        chunk_crc = struct.pack('>I', zlib.crc32(chunk_type + data) & 0xffffffff)
        return chunk_len + chunk_type + data + chunk_crc

    # PNG signature
    signature = b'\x89PNG\r\n\x1a\n'

    # IHDR chunk
    ihdr_data = struct.pack('>IIBBBBB', width, height, 8, 6, 0, 0, 0)
    ihdr = png_chunk(b'IHDR', ihdr_data)

    # IDAT chunk (image data)
    raw_data = b''
    for y in range(height):
        raw_data += b'\x00'  # filter byte
        for x in range(width):
            idx = (y * width + x) * 4
            raw_data += bytes(pixels[idx:idx+4])

    compressed = zlib.compress(raw_data, 9)
    idat = png_chunk(b'IDAT', compressed)

    # IEND chunk
    iend = png_chunk(b'IEND', b'')

    return signature + ihdr + idat + iend


def blend(bg, fg, alpha):
    """Blend foreground color onto background with alpha."""
    return int(bg * (1 - alpha) + fg * alpha)


def draw_rounded_rect(pixels, width, x1, y1, x2, y2, radius, r, g, b, a):
    """Draw a rounded rectangle."""
    for y in range(y1, y2):
        for x in range(x1, x2):
            # Check if inside rounded corners
            inside = True
            alpha = a / 255.0

            # Top-left corner
            if x < x1 + radius and y < y1 + radius:
                dx, dy = x - (x1 + radius), y - (y1 + radius)
                if dx*dx + dy*dy > radius*radius:
                    inside = False

            # Top-right corner
            if x >= x2 - radius and y < y1 + radius:
                dx, dy = x - (x2 - radius - 1), y - (y1 + radius)
                if dx*dx + dy*dy > radius*radius:
                    inside = False

            # Bottom-left corner
            if x < x1 + radius and y >= y2 - radius:
                dx, dy = x - (x1 + radius), y - (y2 - radius - 1)
                if dx*dx + dy*dy > radius*radius:
                    inside = False

            # Bottom-right corner
            if x >= x2 - radius and y >= y2 - radius:
                dx, dy = x - (x2 - radius - 1), y - (y2 - radius - 1)
                if dx*dx + dy*dy > radius*radius:
                    inside = False

            if inside and 0 <= x < width and 0 <= y < len(pixels) // (width * 4):
                idx = (y * width + x) * 4
                pixels[idx] = blend(pixels[idx], r, alpha)
                pixels[idx+1] = blend(pixels[idx+1], g, alpha)
                pixels[idx+2] = blend(pixels[idx+2], b, alpha)
                pixels[idx+3] = min(255, pixels[idx+3] + a)


def draw_arrow(pixels, width, cx, cy, size, r, g, b, a):
    """Draw a right-pointing arrow."""
    # Arrow dimensions
    shaft_width = size * 0.6
    shaft_height = size * 0.25
    head_width = size * 0.4
    head_height = size * 0.5

    alpha = a / 255.0

    for y in range(int(cy - head_height), int(cy + head_height)):
        for x in range(int(cx - shaft_width), int(cx + head_width)):
            inside = False

            # Shaft
            if x < cx and abs(y - cy) < shaft_height / 2:
                inside = True

            # Arrow head (triangle)
            if x >= cx - shaft_width * 0.1:
                # Triangle from cx to cx + head_width
                progress = (x - cx) / head_width if head_width > 0 else 0
                max_y_offset = head_height * (1 - progress)
                if abs(y - cy) < max_y_offset and x < cx + head_width:
                    inside = True

            if inside:
                idx = (y * width + x) * 4
                if 0 <= idx < len(pixels) - 3:
                    pixels[idx] = blend(pixels[idx], r, alpha)
                    pixels[idx+1] = blend(pixels[idx+1], g, alpha)
                    pixels[idx+2] = blend(pixels[idx+2], b, alpha)
                    pixels[idx+3] = min(255, pixels[idx+3] + a)


def main():
    # DMG window size (standard size that works well)
    width = 660
    height = 400

    # Create pixel buffer (RGBA)
    pixels = bytearray(width * height * 4)

    # Background gradient (light beige/cream like Claude's)
    for y in range(height):
        for x in range(width):
            # Gradient from top to bottom
            t = y / height
            # Light cream color similar to the Claude screenshot
            r = int(245 - t * 10)
            g = int(241 - t * 10)
            b = int(235 - t * 10)

            idx = (y * width + x) * 4
            pixels[idx] = r
            pixels[idx+1] = g
            pixels[idx+2] = b
            pixels[idx+3] = 255

    # Draw arrow in the center (dark gray like Claude's)
    arrow_cx = width // 2
    arrow_cy = height // 2 - 20  # Slightly above center
    draw_arrow(pixels, width, arrow_cx, arrow_cy, 80, 80, 80, 80, 230)

    # Save the image
    script_dir = os.path.dirname(os.path.abspath(__file__))
    output_path = os.path.join(script_dir, 'dmg-background.png')

    png_data = create_png(width, height, pixels)
    with open(output_path, 'wb') as f:
        f.write(png_data)

    print(f"Created: {output_path}")
    print(f"Size: {width}x{height}")


if __name__ == '__main__':
    main()
