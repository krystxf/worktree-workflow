#!/usr/bin/env bash
set -euo pipefail

WTW="${WTW:-./wtw}"

rm -rf ~/.config/worktree-workflow
rm -rf /tmp/nosync-project /tmp/nosync-project--worktrees

mkdir -p /tmp/nosync-project
cd /tmp/nosync-project
git init
echo ".env" >> .gitignore
echo "hello" > file.txt
echo "SECRET=123" > .env
git add -A
git commit -m "init"
git branch no-sync-test

cat > .worktree-workflow.json << 'EOF'
{
  "sync_ignored": false,
  "sync_excludes": [],
  "post_copy_hooks": []
}
EOF

$WTW create no-sync-test

WORKTREE_DIR="/tmp/nosync-project--worktrees/nosync-project--no-sync-test"
test ! -f "$WORKTREE_DIR/.env" || (echo "FAIL: .env should not be synced when sync_ignored=false" && exit 1)

echo "PASS: sync_ignored=false skips syncing"
