#!/usr/bin/env bash
# Usage: test_list.sh [ls|list]
# No arg: verify worktrees via git worktree list.
# With arg: verify alias --help works.
source "$(dirname "$0")/helpers.sh"

CMD="${1:-}"
cd /tmp/test-project

if [ -z "$CMD" ]; then
  OUTPUT=$(git worktree list)
  echo "$OUTPUT"
  assert_output_contains "$OUTPUT" "feature-one"
  assert_output_contains "$OUTPUT" "feature-two"
  pass "worktree list"
else
  assert_command_succeeds "$WTW $CMD --help"
  pass "worktree list ($CMD)"
fi
