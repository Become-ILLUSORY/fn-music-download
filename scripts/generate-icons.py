#!/usr/bin/env python3
"""Generate placeholder PNG icons for FNOS packaging."""

import struct, zlib, os

def create_png(width, height, r, g, b, filepath):
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

base = os.path.dirname(os.path.abspath(__file__))
pkg = os.path.join(base, '..', 'package')

create_png(64, 64, 124, 77, 255, os.path.join(pkg, 'ICON.PNG'))
create_png(256, 256, 124, 77, 255, os.path.join(pkg, 'ICON_256.PNG'))
create_png(64, 64, 124, 77, 255, os.path.join(pkg, 'app', 'ui', 'images', 'icon_64.png'))
create_png(256, 256, 124, 77, 255, os.path.join(pkg, 'app', 'ui', 'images', 'icon_256.png'))
print('Icons generated at', pkg)
