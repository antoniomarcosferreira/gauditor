#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
cd "$ROOT_DIR"

OUT_PROFILE="coverage.out"
OUT_HTML="docs/coverage.html"
TEMPLATE="docs/coverage_template.html"

echo "[coverage] running go tests with coverage (excluding examples)..."
PKGS=$(go list ./... | grep -v "/examples/")
if [ -z "$PKGS" ]; then
  echo "no packages found"; exit 1
fi
go test $PKGS -covermode=atomic -coverprofile="$OUT_PROFILE" >/dev/null

mkdir -p docs
echo "[coverage] generating HTML report at $OUT_HTML"
RAW=$(go tool cover -func="$OUT_PROFILE")
AGG=$(printf '%s\n' "$RAW" | awk '
  BEGIN{FS="[[:space:]]+"}
  /total:/ {total=$NF; next}
  {
    pct=$NF; $NF=""; gsub(/[[:space:]]+$/,"",$0);
    key=$0; gsub(/"/,"",key);
    data[key]=pct; order[++i]=key
  }
  END{
    for (k=1;k<=i;k++){ key=order[k]; if (seen[key]++) continue; print key" "data[key] }
    print "total: (statements) "total
  }')
# Escape backticks and backslashes for JS template literal safety
ESCAPED=${AGG//\\/\\\\}
ESCAPED=${ESCAPED//\`/\`}
CONTENT=$(cat "$TEMPLATE")
CONTENT=${CONTENT/__COVERAGE_FUNC__/$ESCAPED}
printf "%s" "$CONTENT" > "$OUT_HTML"

rm -f "$OUT_PROFILE"
echo "[coverage] done. open $OUT_HTML"


