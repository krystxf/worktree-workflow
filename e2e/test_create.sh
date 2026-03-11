#!/usr/bin/env bash
source "$(dirname "$0")/helpers.sh"

cd /tmp/test-project

cat > .worktree-workflow.json << 'EOF'
{
  "sync_ignored": true,
  "sync_excludes": ["node_modules"],
  "post_copy_hooks": []
}
EOF

run_wtw create feature-one

WORKTREE="/tmp/test-project--worktrees/test-project--feature-one"
assert_dir_exists "$WORKTREE"
assert_branch "$WORKTREE" "feature-one"
assert_file_exists "$WORKTREE/feature.js"

pass "create worktree"
