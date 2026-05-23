#!/usr/bin/env bash
set -euo pipefail

VERSION="${VERSION:-$(git describe --tags --abbrev=0 2>/dev/null || echo dev)}"
COMMIT_HASH="${COMMIT_HASH:-$(git rev-parse --short HEAD 2>/dev/null || echo unknown)}"
COUNT="${COUNT:-$(git rev-list "${VERSION}"..HEAD --count 2>/dev/null || echo 0)}"
ITERATION="${ITERATION:-${COUNT}.${COMMIT_HASH}}"
OUT_DIR="${OUT_DIR:-dist/rpm}"

mkdir -p build/rpm "${OUT_DIR}"

go build -o build/rpm/tl -ldflags "-s -w -X main.version=${VERSION}" .

fpm \
  --force \
  --version "${VERSION}" \
  --iteration "${ITERATION}" \
  --package "${OUT_DIR}" \
  build/rpm/tl=/usr/bin/tl
