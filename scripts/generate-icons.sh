#!/bin/bash
# Generate placeholder icon PNGs for FNOS packaging
# Requires: python3 with PIL (Pillow), or uses a fallback

set -e
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
PACKAGE_DIR="$PROJECT_DIR/package"

generate_icon_python() {
    python3 << 'PYEOF'
import struct, zlib, os, sys

def create_png(width, height, r, g, b, filepath):
    """Create a simple solid-color PNG."""
    def chunk(chunk_type, data):
        c = chunk_type + data
        crc = struct.pack('>I', zlib.crc32(c) & 0xFFFFFFFF)
        return struct.pack('>I', len(data)) + c + crc

    sig = b'\x89PNG\r\n\x1a\n'
    ihdr = chunk(b'IHDR', struct.pack('>IIBBBBB', width, height, 8, 2, 0, 0, 0))
    raw = b''
    for _ in range(height):
        raw += b'\x00' + bytes([r, g, b]) * width
    idat = chunk(b'IDAT', zlib.compress(raw))
    iend = chunk(b'IEND', b'')

    os.makedirs(os.path.dirname(filepath) or '.', exist_ok=True)
    with open(filepath, 'wb') as f:
        f.write(sig + ihdr + idat + iend)

base = sys.argv[1]
# Root icons
create_png(64, 64, 124, 77, 255, os.path.join(base, 'ICON.PNG'))
create_png(256, 256, 124, 77, 255, os.path.join(base, 'ICON_256.PNG'))
# UI icons
create_png(64, 64, 124, 77, 255, os.path.join(base, 'app', 'ui', 'images', 'icon_64.png'))
create_png(256, 256, 124, 77, 255, os.path.join(base, 'app', 'ui', 'images', 'icon_256.png'))
print("Icons generated at:", base)
PYEOF
}

if command -v python3 &>/dev/null; then
    generate_icon_python "$PACKAGE_DIR"
else
    echo "WARNING: python3 not found. Create icon files manually:"
    echo "  $PACKAGE_DIR/ICON.PNG (64x64)"
    echo "  $PACKAGE_DIR/ICON_256.PNG (256x256)"
    echo "  $PACKAGE_DIR/app/ui/images/icon_64.png (64x64)"
    echo "  $PACKAGE_DIR/app/ui/images/icon_256.png (256x256)"
fi
