#!/usr/bin/env bash
# Test global --force: wtw --force create creates branch without prompting.
source "$(dirname "$0")/helpers.sh"

write_global_config
rm -rf /tmp/force-create-project /tmp/force-create-project--worktrees

mkdir -p /tmp/force-create-project
cd /tmp/force-create-project
git init
echo "hello" > file.txt
git add -A
git commit -m "init"

# --force should create new branch without "Create it? [y/N]" prompt
OUTPUT=$(run_wtw --force create new-branch-from-force)
echo "$OUTPUT" | grep -q "Create it" && _fail "should not prompt with --force"
assert_output_contains "$OUTPUT" "Done"

WORKTREE="/tmp/force-create-project--worktrees/force-create-project--new-branch-from-force"
assert_dir_exists "$WORKTREE"
assert_command_succeeds "git branch | grep -q new-branch-from-force"

pass "global --force create (new branch, no prompt)"
