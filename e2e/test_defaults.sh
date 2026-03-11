#!/usr/bin/env bash
set -euo pipefail

WTW="${WTW:-./wtw}"

# Keep global config with auto_open_editor=false but remove naming overrides
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

$WTW create test-defaults

WORKTREE_DIR="/tmp/bare-project--worktrees/bare-project--test-defaults"
test -d "$WORKTREE_DIR" || (echo "FAIL: default naming worktree not at $WORKTREE_DIR" && exit 1)

echo "PASS: defaults without config"
