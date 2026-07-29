#!/bin/bash
# Build script for fn-music-dl FNOS package
# Builds Go backend, React frontend, and assembles the .fpk package

set -e
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

BACKEND_DIR="$PROJECT_DIR/backend"
FRONTEND_DIR="$PROJECT_DIR/frontend"
PACKAGE_DIR="$PROJECT_DIR/package"
PACKAGE_SERVER_DIR="$PACKAGE_DIR/app/server"

echo "=== Building fn-music-dl ==="

# 1. Build frontend
echo ""
echo "--- Building frontend ---"
cd "$FRONTEND_DIR"
if [ ! -d "node_modules" ]; then
    npm install
fi
npm run build

# Copy frontend dist to backend static dir (for embedding)
echo "Copying frontend dist to backend/api/static/"
rm -rf "$BACKEND_DIR/api/static"
cp -r "$FRONTEND_DIR/dist" "$BACKEND_DIR/api/static"

# 2. Build Go backend
echo ""
echo "--- Building Go backend ---"
cd "$BACKEND_DIR"

# Download dependencies
go mod download
go mod tidy

# Build for target architecture
GOOS=${GOOS:-linux}
GOARCH=${GOARCH:-arm64}
CGO_ENABLED=0
echo "Building for $GOOS/$GOARCH..."

go build -ldflags="-s -w" -o "$PACKAGE_SERVER_DIR/music-dl" .

echo "Binary size: $(du -h "$PACKAGE_SERVER_DIR/music-dl" | cut -f1)"

# 3. Copy frontend dist to package server dir
echo ""
echo "--- Assembling package ---"
mkdir -p "$PACKAGE_SERVER_DIR/public"
cp -r "$FRONTEND_DIR/dist/"* "$PACKAGE_SERVER_DIR/public/"

# 4. Set permissions
chmod 755 "$PACKAGE_SERVER_DIR/music-dl"
chmod 755 "$PACKAGE_DIR/cmd/main"
chmod 755 "$PACKAGE_DIR/cmd/install_callback"
chmod 755 "$PACKAGE_DIR/cmd/config_callback"

# 5. Build FPK package
echo ""
echo "--- Building FPK package ---"
cd "$PACKAGE_DIR"
if command -v fnpack &>/dev/null; then
    fnpack build --directory "$PACKAGE_DIR"
    echo ""
    echo "FPK package created: $PACKAGE_DIR/../music-dl.fpk"
    ls -lh "$PACKAGE_DIR/../music-dl.fpk"
else
    echo "WARNING: fnpack not found."
    echo "Install fnpack first or manually run: fnpack build --directory $PACKAGE_DIR"
    echo "Package directory is ready at: $PACKAGE_DIR"
fi

echo ""
echo "=== Build complete ==="
