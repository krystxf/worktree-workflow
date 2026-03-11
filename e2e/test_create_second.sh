#!/usr/bin/env bash
source "$(dirname "$0")/helpers.sh"

cd /tmp/test-project
run_wtw create feature-two

WORKTREE="/tmp/test-project--worktrees/test-project--feature-two"
assert_dir_exists "$WORKTREE"
assert_branch "$WORKTREE" "feature-two"

pass "create second worktree"
