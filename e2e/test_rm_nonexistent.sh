#!/usr/bin/env bash
set -euo pipefail

WTW="${WTW:-./wtw}"

cd /tmp/test-project
if $WTW rm nonexistent-branch 2>&1; then
  echo "FAIL: should have failed"
  exit 1
fi

echo "PASS: non-existent worktree removal fails"
