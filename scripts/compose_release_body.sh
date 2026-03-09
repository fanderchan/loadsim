#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "usage: $0 <tag> <output-file>" >&2
  exit 1
fi

TAG="$1"
OUTPUT_FILE="$2"
REPO_SLUG="${GITHUB_REPOSITORY:-}"
NOTES_FILE=".github/release-notes/${TAG}.md"

if [ ! -f "$NOTES_FILE" ]; then
  echo "missing release notes file: $NOTES_FILE" >&2
  exit 1
fi

mkdir -p "$(dirname "$OUTPUT_FILE")"
cat "$NOTES_FILE" >"$OUTPUT_FILE"

PREVIOUS_TAG="$(
  git for-each-ref --sort=creatordate --format='%(refname:short)' refs/tags |
    awk -v current="$TAG" '
      BEGIN { previous = "" }
      {
        if ($0 == current) {
          print previous
          exit
        }
        previous = $0
      }
    '
)"

printf "\n\n## Changelog\n\n" >>"$OUTPUT_FILE"

if [ -n "$PREVIOUS_TAG" ] && [ -n "$REPO_SLUG" ]; then
  printf "**Full Changelog**: https://github.com/%s/compare/%s...%s\n" \
    "$REPO_SLUG" "$PREVIOUS_TAG" "$TAG" >>"$OUTPUT_FILE"
else
  printf "Initial release. No previous tag is available for comparison.\n" >>"$OUTPUT_FILE"
fi
