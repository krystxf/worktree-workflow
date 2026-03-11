#!/usr/bin/env bash
# Usage: test_rm.sh [rm|remove]
source "$(dirname "$0")/helpers.sh"

CMD="${1:-rm}"
cd /tmp/test-project

if [ "$CMD" = "rm" ]; then
  BRANCH="feature-two"
else
  # "remove" alias: feature-two already removed by first run
  BRANCH="bugfix-123"
  run_wtw create "$BRANCH"
fi

run_wtw $CMD "$BRANCH"

WORKTREE="/tmp/test-project--worktrees/test-project--$BRANCH"
assert_dir_not_exists "$WORKTREE"

OUTPUT=$(git worktree list)
echo "$OUTPUT" | grep -q "$BRANCH" && _fail "$BRANCH still in git worktree list"

pass "remove worktree ($CMD)"
