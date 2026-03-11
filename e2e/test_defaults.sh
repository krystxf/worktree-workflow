#!/usr/bin/env bash
source "$(dirname "$0")/helpers.sh"

# Reset to minimal global config (no naming overrides)
mkdir -p ~/.config/worktree-workflow
cat > ~/.config/worktree-workflow/config.json << 'EOF'
{
  "auto_open_editor": false
}
EOF

rm -rf /tmp/bare-project /tmp/bare-project--worktrees
mkdir -p /tmp/bare-project
cd /tmp/bare-project
git init
echo "hello" > file.txt
git add -A
git commit -m "init"
git branch test-defaults

run_wtw create test-defaults

assert_dir_exists "/tmp/bare-project--worktrees/bare-project--test-defaults"

pass "defaults without config"
