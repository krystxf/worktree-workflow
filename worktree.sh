#!/usr/bin/env bash
set -euo pipefail

# --- Configuration ---
SYNC_IGNORED=true
SYNC_EXCLUDES=('node_modules')

# --- Main ---
BRANCH=$1
if [ -z "${BRANCH:-}" ]; then
  echo "Usage: worktree.sh <branch>"
  exit 1
fi

ROOT=$(git rev-parse --show-toplevel)
REPO=$(basename "$ROOT")
WORKTREE_DIR="$(dirname "$ROOT")/${REPO}--worktrees/${REPO}--${BRANCH}"

# Create worktree
echo "🌳 Creating worktree for branch '$BRANCH' at $WORKTREE_DIR"
if ! git worktree add "$WORKTREE_DIR" "$BRANCH"; then
  echo "💥 Failed to create worktree"
  exit 1
fi

# Post create

## Sync gitignored files (e.g. .env) via hard links
if $SYNC_IGNORED; then
  echo "🔗 Syncing gitignored files via hard links..."
  grep_args=()
  for exc in "${SYNC_EXCLUDES[@]}"; do
    grep_args+=(-e "$exc")
  done
  git -C "$ROOT" ls-files --others --ignored --exclude-standard -z | grep -zvF "${grep_args[@]}" | rsync -av --hard-links --from0 --files-from=- "$ROOT/" "$WORKTREE_DIR/"
fi

## Other post create actions
## echo "📦 Installing dependencies..."
## yarn --cwd "$WORKTREE_DIR" install
## etc.

## Open worktree in IDE
echo "🚀 Opening worktree in IDE..."
cursor "$WORKTREE_DIR/"

echo "✅ Done!"
