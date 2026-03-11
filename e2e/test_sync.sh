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

# Verify hard links (same inode)
if [[ "$(uname)" == "Darwin" ]]; then
  INODE_ORIG=$(stat -f %i /tmp/test-project/.env)
  INODE_WT=$(stat -f %i "$WORKTREE_DIR/.env")
else
  INODE_ORIG=$(stat -c %i /tmp/test-project/.env)
  INODE_WT=$(stat -c %i "$WORKTREE_DIR/.env")
fi
test "$INODE_ORIG" = "$INODE_WT" || (echo "FAIL: .env not hard-linked (inodes: $INODE_ORIG vs $INODE_WT)" && exit 1)

echo "PASS: sync gitignored files"
