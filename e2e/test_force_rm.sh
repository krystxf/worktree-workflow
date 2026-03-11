#!/usr/bin/env bash
set -euo pipefail

# Test global --force: wtw --force rm removes dirty worktree without prompting.

WTW="${WTW:-./wtw}"

cd /tmp/test-project
WORKTREE_DIR="/tmp/test-project--worktrees/test-project--feature-one"

# Ensure worktree exists and has untracked file
if [ ! -d "$WORKTREE_DIR" ]; then
  $WTW create feature-one
fi
echo "dirty" > "$WORKTREE_DIR/untracked.txt"

# Global --force: no "Force remove? [y/N]" prompt, removes immediately
OUTPUT=$($WTW --force rm feature-one 2>&1)
echo "$OUTPUT" | grep -q "✓" || (echo "FAIL: expected success message, got: $OUTPUT" && exit 1)
echo "$OUTPUT" | grep -q "Force remove" && (echo "FAIL: should not prompt with --force, got: $OUTPUT" && exit 1)

test ! -d "$WORKTREE_DIR" || (echo "FAIL: worktree still exists after force remove" && exit 1)

echo "PASS: global --force rm (no prompt)"
