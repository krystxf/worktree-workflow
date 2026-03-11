#!/usr/bin/env bash
set -euo pipefail

BRANCH=$1
if [ -z "${BRANCH:-}" ]; then
  echo "Usage: cleanup.sh <branch>"
  exit 1
fi

ROOT=$(git rev-parse --show-toplevel)
REPO=$(basename "$ROOT")
WORKTREE_DIR="$(dirname "$ROOT")/${REPO}--worktrees/${REPO}--${BRANCH}"

if ! git worktree remove "$WORKTREE_DIR"; then
  exit 1
fi

echo "Removed worktree for branch '$BRANCH'"
