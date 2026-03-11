#!/usr/bin/env bash
set -euo pipefail

WTW="${WTW:-./wtw}"

cd /tmp/test-project
WORKTREE_DIR="/tmp/test-project--worktrees/test-project--feature-one"

# Create untracked file in worktree
echo "dirty" > "$WORKTREE_DIR/untracked.txt"

# Normal remove should fail
if $WTW rm feature-one 2>&1 | grep -q "modified or untracked"; then
  echo "OK: normal remove detected dirty worktree"
fi

# Force remove
$WTW rm -f feature-one

test ! -d "$WORKTREE_DIR" || (echo "FAIL: worktree still exists after force remove" && exit 1)

echo "PASS: force remove"
