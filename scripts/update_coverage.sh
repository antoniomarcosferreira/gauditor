#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
cd "$ROOT_DIR"

OUT_PROFILE="coverage.out"
OUT_HTML="docs/coverage.html"
TEMPLATE="docs/coverage_template.html"

echo "[coverage] running go tests with coverage..."
go test ./... -covermode=atomic -coverprofile="$OUT_PROFILE" >/dev/null

mkdir -p docs
echo "[coverage] generating HTML report at $OUT_HTML"
# Build pretty dashboard by embedding go tool cover -func output into template
FUNC_OUTPUT=$(go tool cover -func="$OUT_PROFILE")
ESCAPED=$(printf '%s' "$FUNC_OUTPUT" | python3 -c 'import sys,json; print(json.dumps(sys.stdin.read()))')
CONTENT=$(cat "$TEMPLATE")
CONTENT=${CONTENT/__COVERAGE_FUNC__/$ESCAPED}
printf "%s" "$CONTENT" > "$OUT_HTML"

rm -f "$OUT_PROFILE"
echo "[coverage] done. open $OUT_HTML"


