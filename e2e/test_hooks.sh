#!/usr/bin/env bash
source "$(dirname "$0")/helpers.sh"

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

run_wtw create bugfix-123

WORKTREE="/tmp/test-project--worktrees/test-project--bugfix-123"
assert_file_exists "$WORKTREE/hook-step-1.txt"
assert_file_contents "$WORKTREE/hook-step-2.txt" "hello from hook"

# Resolve symlinks (macOS /tmp -> /private/tmp)
EXPECTED_PWD=$(cd "$WORKTREE" && pwd -P)
assert_file_contents "$WORKTREE/hook-step-3.txt" "$EXPECTED_PWD"

pass "post-copy hooks"
