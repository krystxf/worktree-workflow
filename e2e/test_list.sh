#!/usr/bin/env bash
set -euo pipefail

# Usage: test_list.sh [ls|list]
# No arg: verify worktrees via git worktree list.
# With arg (ls or list): verify $WTW <cmd> --help works for the alias.

WTW="${WTW:-./wtw}"
CMD="${1:-}"

cd /tmp/test-project

if [ -z "$CMD" ]; then
  OUTPUT=$(git worktree list)
  echo "$OUTPUT"
  echo "$OUTPUT" | grep -q "feature-one" || (echo "FAIL: feature-one not in list" && exit 1)
  echo "$OUTPUT" | grep -q "feature-two" || (echo "FAIL: feature-two not in list" && exit 1)
  echo "PASS: worktree list"
else
  # Test that the subcommand alias works (e.g. wtw list --help)
  $WTW "$CMD" --help >/dev/null 2>&1 || (echo "FAIL: $WTW $CMD --help failed" && exit 1)
  echo "PASS: worktree list ($CMD)"
fi
