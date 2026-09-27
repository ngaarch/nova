#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-1.0.0}"
GIT_COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
BUILD_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
DIST_DIR="dist"

echo "=========================================="
echo " Building Nova CLI Release $VERSION"
echo " Commit: $GIT_COMMIT"
echo " Date:   $BUILD_DATE"
echo "=========================================="

mkdir -p "$DIST_DIR"
rm -rf "$DIST_DIR"/*

PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

LDFLAGS="-s -w -X nova/internal/cli.Version=${VERSION} -X nova/internal/cli.GitCommit=${GIT_COMMIT} -X nova/internal/cli.BuildDate=${BUILD_DATE}"

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS="${PLATFORM%/*}"
    GOARCH="${PLATFORM#*/}"
    OUTPUT_NAME="nova_${VERSION}_${GOOS}_${GOARCH}"
    BIN_NAME="nova"
    if [ "$GOOS" = "windows" ]; then
        BIN_NAME="nova.exe"
    fi

    TARGET_DIR="$DIST_DIR/$OUTPUT_NAME"
    mkdir -p "$TARGET_DIR"

    echo "-> Building $GOOS/$GOARCH..."
    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -trimpath -ldflags "$LDFLAGS" -o "$TARGET_DIR/$BIN_NAME" ./cmd/nova

    cp LICENSE "$TARGET_DIR/"
    cp README.md "$TARGET_DIR/"

    if [ "$GOOS" = "windows" ]; then
        (cd "$DIST_DIR" && zip -q -r "${OUTPUT_NAME}.zip" "$OUTPUT_NAME")
    else
        tar -czf "$DIST_DIR/${OUTPUT_NAME}.tar.gz" -C "$DIST_DIR" "$OUTPUT_NAME"
    fi

    rm -rf "$TARGET_DIR"
done

echo "-> Generating SHA256 checksums..."
(cd "$DIST_DIR" && sha256sum * > SHA256SUMS.txt)

echo "=========================================="
echo " Release artifacts generated in $DIST_DIR:"
ls -lh "$DIST_DIR"
echo "=========================================="
