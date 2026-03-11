#!/usr/bin/env bash
set -euo pipefail

WTW="${WTW:-./wtw}"

cd /tmp/test-project
$WTW create feature-two

WORKTREE_DIR="/tmp/test-project--worktrees/test-project--feature-two"
test -d "$WORKTREE_DIR" || (echo "FAIL: second worktree not created" && exit 1)

BRANCH=$(git -C "$WORKTREE_DIR" branch --show-current)
test "$BRANCH" = "feature-two" || (echo "FAIL: expected feature-two, got $BRANCH" && exit 1)

echo "PASS: create second worktree"
