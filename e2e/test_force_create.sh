#!/usr/bin/env bash
set -euo pipefail

# Test global --force: wtw --force create <new-branch> creates branch and worktree without prompting.

WTW="${WTW:-./wtw}"

mkdir -p ~/.config/worktree-workflow
cat > ~/.config/worktree-workflow/config.json << 'EOF'
{
  "auto_open_editor": false
}
EOF
rm -rf /tmp/force-create-project /tmp/force-create-project--worktrees

mkdir -p /tmp/force-create-project
cd /tmp/force-create-project
git init
echo "hello" > file.txt
git add -A
git commit -m "init"

# Global --force: no "Create it? [y/N]" prompt, creates branch and worktree (non-interactive)
OUTPUT=$($WTW --force create new-branch-from-force 2>&1)
echo "$OUTPUT" | grep -q "Create it" && (echo "FAIL: should not prompt with --force, got: $OUTPUT" && exit 1)
echo "$OUTPUT" | grep -q "Done" || (echo "FAIL: expected Done, got: $OUTPUT" && exit 1)

WORKTREE_DIR="/tmp/force-create-project--worktrees/force-create-project--new-branch-from-force"
test -d "$WORKTREE_DIR" || (echo "FAIL: worktree dir not created" && exit 1)
cd /tmp/force-create-project
git branch | grep -q "new-branch-from-force" || (echo "FAIL: branch not created" && exit 1)

echo "PASS: global --force create (new branch, no prompt)"
