#!/usr/bin/env bash
set -euo pipefail

WTW="${WTW:-./wtw}"

cd /tmp/test-project
if $WTW create feature-one 2>&1; then
  echo "FAIL: should have failed for existing worktree"
  exit 1
fi

echo "PASS: duplicate worktree fails gracefully"
