#!/usr/bin/env bash
set -euo pipefail

WTW="${WTW:-./wtw}"

cd /tmp/test-project

cat > .worktree-workflow.json << 'EOF'
{
  "sync_ignored": true,
  "sync_excludes": ["node_modules"],
  "post_copy_hooks": [
    "touch hook-step-1.txt",
    "echo 'hello from hook' > hook-step-2.txt",
    "pwd > hook-step-3.txt"
  ]
}
EOF

$WTW create bugfix-123

WORKTREE_DIR="/tmp/test-project--worktrees/test-project--bugfix-123"

test -f "$WORKTREE_DIR/hook-step-1.txt" || (echo "FAIL: hook step 1 didn't run" && exit 1)
test "$(cat "$WORKTREE_DIR/hook-step-2.txt")" = "hello from hook" || (echo "FAIL: hook step 2 output wrong" && exit 1)
test "$(cat "$WORKTREE_DIR/hook-step-3.txt")" = "$WORKTREE_DIR" || (echo "FAIL: hook step 3 pwd wrong" && exit 1)

echo "PASS: post-copy hooks"
