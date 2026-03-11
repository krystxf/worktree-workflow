#!/usr/bin/env bash
set -euo pipefail

WTW="${WTW:-./wtw}"

cd /tmp/test-project
$WTW rm feature-two

WORKTREE_DIR="/tmp/test-project--worktrees/test-project--feature-two"
test ! -d "$WORKTREE_DIR" || (echo "FAIL: worktree dir still exists" && exit 1)

OUTPUT=$(git worktree list)
echo "$OUTPUT" | grep -q "feature-two" && (echo "FAIL: feature-two still in git worktree list" && exit 1)

echo "PASS: remove worktree"
