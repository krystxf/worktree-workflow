#!/usr/bin/env bash
source "$(dirname "$0")/helpers.sh"

cd /tmp/test-project
run_wtw rm -f bugfix-123 || true
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

run_wtw create hook-fail-test || true

WORKTREE="/tmp/test-project--worktrees/test-project--hook-fail-test"
assert_file_exists "$WORKTREE/before-fail.txt"
assert_file_not_exists "$WORKTREE/after-fail.txt"

pass "failing hook stops execution"
