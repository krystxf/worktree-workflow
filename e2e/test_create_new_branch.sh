#!/usr/bin/env bash
source "$(dirname "$0")/helpers.sh"

write_global_config
rm -rf /tmp/newbranch-project /tmp/newbranch-project--worktrees

mkdir -p /tmp/newbranch-project
cd /tmp/newbranch-project
git init
echo "hello" > file.txt
git add -A
git commit -m "init"

# Non-interactive mode should fail with clear error for non-existent branch
OUTPUT=$(run_wtw_fail create nonexistent-branch)
assert_output_contains "$OUTPUT" "does not exist"

pass "create non-existent branch shows error in non-interactive mode"
