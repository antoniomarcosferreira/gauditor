#!/usr/bin/env bash
set -euo pipefail

# Create a GitHub release locally using gh CLI, extracting notes from CHANGELOG.
# Prereqs: gh CLI authenticated (gh auth login) and git remote origin set to GitHub.
# Usage: bash scripts/create_release_local.sh [vX.Y.Z]

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"

TAG="${1:-}"
if [[ -z "$TAG" ]]; then
  if [[ -f VERSION ]]; then TAG=$(tr -d '\n' < VERSION); fi
fi
if [[ -z "$TAG" ]]; then echo "Provide tag or ensure VERSION exists" >&2; exit 1; fi

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI not found. Install: https://cli.github.com/" >&2
  exit 1
fi

# Extract notes for tag from CHANGELOG.md
awk -v tag="$TAG" '
  BEGIN { start=0 }
  { sub(/\r$/, "") }
  /^## \[/ && start { exit }
  $0 ~ "^## \[" tag "\]" { start=1 }
  start { print }
' CHANGELOG.md > NOTES.md

if [[ ! -s NOTES.md ]]; then
  echo "Could not extract notes for $TAG from CHANGELOG.md" >&2
  exit 1
fi

echo "Creating release $TAG with notes:"
echo "--------------------------------------------------"
sed -n '1,120p' NOTES.md
echo "--------------------------------------------------"

# Create or update the release (idempotent)
gh release view "$TAG" >/dev/null 2>&1 && EXIST=1 || EXIST=0
if [[ "$EXIST" -eq 1 ]]; then
  gh release edit "$TAG" -F NOTES.md -t "$TAG"
else
  gh release create "$TAG" -F NOTES.md -t "$TAG"
fi

echo "Done."


