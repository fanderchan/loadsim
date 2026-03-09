#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "用法: $0 <tag> <output-file>" >&2
  exit 1
fi

TAG="$1"
OUTPUT_FILE="$2"
REPO_SLUG="${GITHUB_REPOSITORY:-}"
NOTES_FILE=".github/release-notes/${TAG}.md"
IS_BETA="false"

if [[ "$TAG" == *"-beta."* ]]; then
  IS_BETA="true"
fi

if [ ! -f "$NOTES_FILE" ]; then
  echo "缺少发布说明文件: $NOTES_FILE" >&2
  exit 1
fi

mkdir -p "$(dirname "$OUTPUT_FILE")"
cat "$NOTES_FILE" >"$OUTPUT_FILE"

PREVIOUS_TAG="$(
  git for-each-ref --sort=creatordate --format='%(refname:short)' refs/tags |
    awk -v current="$TAG" -v is_beta="$IS_BETA" '
      BEGIN { previous = "" }
      {
        if (is_beta != "true" && $0 ~ /-beta\./) {
          next
        }
        if ($0 == current) {
          print previous
          exit
        }
        previous = $0
      }
    '
)"

printf "\n\n## 变更对比\n\n" >>"$OUTPUT_FILE"

if [ -n "$PREVIOUS_TAG" ] && [ -n "$REPO_SLUG" ]; then
  printf "**完整变更**: https://github.com/%s/compare/%s...%s\n" \
    "$REPO_SLUG" "$PREVIOUS_TAG" "$TAG" >>"$OUTPUT_FILE"
elif [ "$IS_BETA" = "true" ]; then
  printf "当前 beta 版本之前没有可用于对比的已发布标签。\n" >>"$OUTPUT_FILE"
else
  printf "当前版本是首个正式版，变更范围为当前仓库全部代码。\n" >>"$OUTPUT_FILE"
  if [ -n "$REPO_SLUG" ]; then
    printf "**当前代码**: https://github.com/%s/tree/%s\n" \
      "$REPO_SLUG" "$TAG" >>"$OUTPUT_FILE"
  fi
fi
