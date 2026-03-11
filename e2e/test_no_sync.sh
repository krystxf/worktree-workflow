#!/usr/bin/env bash
source "$(dirname "$0")/helpers.sh"

write_global_config
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

run_wtw create no-sync-test

assert_file_not_exists "/tmp/nosync-project--worktrees/nosync-project--no-sync-test/.env"

pass "sync_ignored=false skips syncing"
