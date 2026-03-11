#!/usr/bin/env bash
# Test global --force: wtw --force rm removes dirty worktree without prompting.
source "$(dirname "$0")/helpers.sh"

cd /tmp/test-project
WORKTREE="/tmp/test-project--worktrees/test-project--feature-one"

# Ensure worktree exists and is dirty
if [ ! -d "$WORKTREE" ]; then
  run_wtw create feature-one
fi
echo "dirty" > "$WORKTREE/untracked.txt"

# --force should remove without prompt
OUTPUT=$(run_wtw --force rm feature-one)
assert_output_contains "$OUTPUT" "✓"
echo "$OUTPUT" | grep -q "Force remove" && _fail "should not prompt with --force"
assert_dir_not_exists "$WORKTREE"

pass "global --force rm (no prompt)"
