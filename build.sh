#!/bin/bash

set -euo pipefail

TARGET="${1:-linux}"

case "$TARGET" in
    linux)
        OS="linux"
        ARCH="amd64"
        ;;
    mac)
        OS="darwin"
        ARCH="arm64"
        ;;
    win)
        OS="windows"
        ARCH="amd64"
        ;;
    *)
        echo "Usage: $0 [linux|mac|win]" >&2
        exit 1
        ;;
esac

export GOOS=${OS}
export GOARCH=${ARCH}

echo "Building for ${OS} (${ARCH})..."
go build -ldflags "-X main.version=$(git describe --abbrev=0 --always) -X main.commit=$(git rev-parse HEAD)"
