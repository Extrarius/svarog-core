#!/usr/bin/env bash
# Point this repo at versioned git hooks (strips Cursor Co-authored-by trailers).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
chmod +x "$ROOT/.githooks/commit-msg"
git -C "$ROOT" config core.hooksPath .githooks
echo "Installed .githooks (core.hooksPath=.githooks)"
