#!/usr/bin/env bash
set -euo pipefail

WTW="${WTW:-./wtw}"

cd /tmp
if $WTW create some-branch 2>&1; then
  echo "FAIL: should fail outside git repo"
  exit 1
fi
if $WTW ps 2>&1; then
  echo "FAIL: ps should fail outside git repo"
  exit 1
fi

echo "PASS: fails outside git repo"
