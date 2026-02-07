#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-https://mjolnirarmory.com}"

printf "Checking %s...\n" "$BASE_URL"

check() {
  local name="$1"
  local url="$2"
  local expected="$3"

  local code
  code=$(curl -sS -o /dev/null -w "%{http_code}" "$url")

  if [[ "$code" != "$expected" ]]; then
    printf "FAIL: %s -> %s (expected %s)\n" "$name" "$code" "$expected" >&2
    return 1
  fi

  printf "OK: %s -> %s\n" "$name" "$code"
}

check "frontend" "$BASE_URL/" "200"
check "backend health" "$BASE_URL/health" "200"

printf "All checks passed.\n"
