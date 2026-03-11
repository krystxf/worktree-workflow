#!/usr/bin/env bash
set -euo pipefail

WTW="${WTW:-./wtw}"

cd /tmp/test-project

$WTW rm -f bugfix-123 || true

git branch hook-fail-test 2>/dev/null || true

cat > .worktree-workflow.json << 'EOF'
{
  "sync_ignored": false,
  "sync_excludes": [],
  "post_copy_hooks": [
    "touch before-fail.txt",
    "exit 1",
    "touch after-fail.txt"
  ]
}
EOF

$WTW create hook-fail-test || true

WORKTREE_DIR="/tmp/test-project--worktrees/test-project--hook-fail-test"

test -f "$WORKTREE_DIR/before-fail.txt" || (echo "FAIL: hook before failure didn't run" && exit 1)
test ! -f "$WORKTREE_DIR/after-fail.txt" || (echo "FAIL: hook after failure should not have run" && exit 1)

echo "PASS: failing hook stops execution"
