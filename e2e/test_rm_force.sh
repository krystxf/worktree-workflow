#!/usr/bin/env bash
# Usage: test_rm_force.sh <-f|--force>
source "$(dirname "$0")/helpers.sh"

FORCE_FLAG="${1:?Usage: test_rm_force.sh <-f|--force>}"
cd /tmp/test-project

# Use distinct branch per flag so both -f and --force can run
if [ "$FORCE_FLAG" = "-f" ]; then BRANCH="force-rm-f"; else BRANCH="force-rm-double"; fi

git branch "$BRANCH" 2>/dev/null || true
run_wtw create "$BRANCH"

WORKTREE="/tmp/test-project--worktrees/test-project--$BRANCH"
assert_dir_exists "$WORKTREE"

# Make worktree dirty
echo "dirty" > "$WORKTREE/untracked.txt"

# Normal remove should detect dirty state
OUTPUT=$(run_wtw_fail rm "$BRANCH")
assert_output_contains "$OUTPUT" "modified or untracked"

# Force remove should succeed
run_wtw rm $FORCE_FLAG "$BRANCH"
assert_dir_not_exists "$WORKTREE"

pass "force remove ($FORCE_FLAG)"
