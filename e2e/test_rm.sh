#!/usr/bin/env bash
set -euo pipefail

# Usage: test_rm.sh [rm|remove]
# Runs the same remove test with the given subcommand (default: rm).

WTW="${WTW:-./wtw}"
CMD="${1:-rm}"

cd /tmp/test-project

if [ "$CMD" = "rm" ]; then
  BRANCH="feature-two"
else
  # "remove" alias: need a worktree to remove (feature-two already removed by first run)
  BRANCH="bugfix-123"
  $WTW create "$BRANCH"
fi

$WTW $CMD "$BRANCH"

WORKTREE_DIR="/tmp/test-project--worktrees/test-project--$BRANCH"
test ! -d "$WORKTREE_DIR" || (echo "FAIL: worktree dir still exists" && exit 1)

OUTPUT=$(git worktree list)
echo "$OUTPUT" | grep -q "$BRANCH" && (echo "FAIL: $BRANCH still in git worktree list" && exit 1)

echo "PASS: remove worktree ($CMD)"
