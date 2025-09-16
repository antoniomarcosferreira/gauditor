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

# Extract notes for tag from CHANGELOG.md (tolerant a variações)
# Procura primeiro por "## [vX.Y.Z]", depois por "## vX.Y.Z".
start_line=$(grep -n "^## \[$TAG\]" CHANGELOG.md | head -1 | cut -d: -f1 || true)
if [[ -z "${start_line}" ]]; then
  start_line=$(grep -n "^## ${TAG}$" CHANGELOG.md | head -1 | cut -d: -f1 || true)
fi

# Fallback: for pre-releases (e.g., v0.0.1-alpha), try the base tag (v0.0.1)
if [[ -z "${start_line}" ]]; then
  base_tag=${TAG%%-*}
  if [[ -n "${base_tag}" && "${base_tag}" != "${TAG}" ]]; then
    start_line=$(grep -n "^## \[${base_tag}\]" CHANGELOG.md | head -1 | cut -d: -f1 || true)
    if [[ -z "${start_line}" ]]; then
      start_line=$(grep -n "^## ${base_tag}$" CHANGELOG.md | head -1 | cut -d: -f1 || true)
    fi
  fi
fi

# Fallback: use Unreleased section if present
if [[ -z "${start_line}" ]]; then
  start_line=$(grep -n "^## \[Unreleased\]" CHANGELOG.md | head -1 | cut -d: -f1 || true)
fi

if [[ -z "${start_line}" ]]; then
  echo "Could not find heading for ${TAG} in CHANGELOG.md (expected '## [${TAG}] - ...')" >&2
  exit 1
fi

end_line=$(awk -v s=${start_line} 'NR>s && /^## \[/ { print NR; exit }' CHANGELOG.md)
if [[ -z "${end_line}" ]]; then
  end_line=$(wc -l < CHANGELOG.md)
fi

sed -n "${start_line},${end_line}p" CHANGELOG.md | sed 's/\r$//' > NOTES.md

if [[ ! -s NOTES.md ]]; then
  echo "Could not extract notes for ${TAG} from CHANGELOG.md" >&2
  exit 1
fi

echo "Creating release $TAG with notes:"
echo "--------------------------------------------------"
sed -n '1,120p' NOTES.md
echo "--------------------------------------------------"

# Create or update the release (idempotent)
gh release view "$TAG" >/dev/null 2>&1 && EXIST=1 || EXIST=0
PRE=""
if [[ "$TAG" == *-* ]]; then PRE="--prerelease"; fi
if [[ "$EXIST" -eq 1 ]]; then
  gh release edit "$TAG" -F NOTES.md -t "$TAG" $PRE
else
  gh release create "$TAG" -F NOTES.md -t "$TAG" $PRE
fi

echo "Done."


