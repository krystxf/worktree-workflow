#!/usr/bin/env bash
set -euo pipefail

WORKTREE_DIR="/tmp/test-project--worktrees/test-project--feature-one"

# .env should be synced
test -f "$WORKTREE_DIR/.env" || (echo "FAIL: .env not synced" && exit 1)
test "$(cat "$WORKTREE_DIR/.env")" = "SECRET=abc" || (echo "FAIL: .env content wrong" && exit 1)

# .env.local should be synced
test -f "$WORKTREE_DIR/.env.local" || (echo "FAIL: .env.local not synced" && exit 1)

# node_modules should NOT be synced (excluded by default)
test ! -d "$WORKTREE_DIR/node_modules" || (echo "FAIL: node_modules should be excluded" && exit 1)

# dist should be synced (not excluded)
test -f "$WORKTREE_DIR/dist/index.js" || (echo "FAIL: dist/index.js not synced" && exit 1)

echo "PASS: sync gitignored files"
