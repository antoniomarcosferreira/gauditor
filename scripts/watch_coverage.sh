#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
cd "$ROOT_DIR"

if ! command -v fswatch >/dev/null 2>&1; then
  echo "fswatch is required for watch mode. Install via: brew install fswatch"
  exit 1
fi

echo "[coverage] watching for changes... (Ctrl+C to stop)"
fswatch -or . | while read -r _; do
  bash scripts/update_coverage.sh || true
done


