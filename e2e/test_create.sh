#!/usr/bin/env bash
set -euo pipefail

WTW="${WTW:-./wtw}"

cd /tmp/test-project
$WTW create feature-one

WORKTREE_DIR="/tmp/test-project--worktrees/test-project--feature-one"

test -d "$WORKTREE_DIR" || (echo "FAIL: worktree dir not created" && exit 1)

BRANCH=$(git -C "$WORKTREE_DIR" branch --show-current)
test "$BRANCH" = "feature-one" || (echo "FAIL: expected branch feature-one, got $BRANCH" && exit 1)

test -f "$WORKTREE_DIR/feature.js" || (echo "FAIL: feature.js missing" && exit 1)

echo "PASS: create worktree"
