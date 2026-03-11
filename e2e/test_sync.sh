#!/usr/bin/env bash
source "$(dirname "$0")/helpers.sh"

WORKTREE="/tmp/test-project--worktrees/test-project--feature-one"

assert_file_exists "$WORKTREE/.env"
assert_file_contents "$WORKTREE/.env" "SECRET=abc"
assert_file_exists "$WORKTREE/.env.local"
assert_dir_not_exists "$WORKTREE/node_modules"
assert_file_exists "$WORKTREE/dist/index.js"

pass "sync gitignored files"
