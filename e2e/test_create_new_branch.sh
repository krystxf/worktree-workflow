#!/usr/bin/env bash
set -euo pipefail

WTW="${WTW:-./wtw}"

# Keep global config with auto_open_editor=false
mkdir -p ~/.config/worktree-workflow
cat > ~/.config/worktree-workflow/config.json << 'EOF'
{
  "auto_open_editor": false
}
EOF
rm -rf /tmp/newbranch-project /tmp/newbranch-project--worktrees

mkdir -p /tmp/newbranch-project
cd /tmp/newbranch-project
git init
echo "hello" > file.txt
git add -A
git commit -m "init"

# In non-interactive mode, creating a non-existent branch should fail with a clear message
OUTPUT=$($WTW create nonexistent-branch 2>&1 || true)
echo "$OUTPUT" | grep -q "does not exist" || (echo "FAIL: expected 'does not exist' error, got: $OUTPUT" && exit 1)

echo "PASS: create non-existent branch shows error in non-interactive mode"
