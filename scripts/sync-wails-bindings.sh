#!/bin/sh

set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT_DIR"

STAGE_FILES=0
if [ "${1:-}" = "--stage" ]; then
    STAGE_FILES=1
fi

export GOCACHE=${GOCACHE:-$ROOT_DIR/.cache/go/build}
export GOMODCACHE=${GOMODCACHE:-$ROOT_DIR/.cache/go/pkg/mod}
mkdir -p "$GOCACHE" "$GOMODCACHE"

TRACKED_BINDINGS_DIR=frontend/bindings/gitlab.com/usescrolls/scribe
EXPECTED_WAILS_VERSION=${WAILS_VERSION:-$(sed -n 's/^WAILS_VERSION ?= //p' Makefile | head -n 1)}
EXPECTED_BINDINGS="
frontend/bindings/gitlab.com/usescrolls/scribe/index.js
frontend/bindings/gitlab.com/usescrolls/scribe/appservice.js
frontend/bindings/gitlab.com/usescrolls/scribe/internal/index.js
frontend/bindings/gitlab.com/usescrolls/scribe/internal/models.js
"

if ! command -v wails3 >/dev/null 2>&1; then
    echo "Error: wails3 is not installed."
    echo "Install it with: make deps"
    exit 1
fi

INSTALLED_WAILS_VERSION=$(wails3 version 2>&1 | tail -n 1)
if [ "$INSTALLED_WAILS_VERSION" != "$EXPECTED_WAILS_VERSION" ]; then
    echo "Error: wails3 version mismatch."
    echo "Expected: $EXPECTED_WAILS_VERSION"
    echo "Found:    $INSTALLED_WAILS_VERSION"
    echo "Run: go install github.com/wailsapp/wails/v3/cmd/wails3@$EXPECTED_WAILS_VERSION"
    exit 1
fi

echo "Refreshing Wails bindings..."
wails3 generate bindings

for file in $EXPECTED_BINDINGS; do
    if [ ! -f "$file" ]; then
        echo "Error: expected generated binding missing: $file"
        exit 1
    fi
done

if [ "$STAGE_FILES" -eq 1 ]; then
    git add "$TRACKED_BINDINGS_DIR"
fi
