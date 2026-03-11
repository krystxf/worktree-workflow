#!/usr/bin/env bash
set -euo pipefail

# Usage: test_rm_force_custom_naming.sh <-f|--force>
# Runs the same force-remove test with custom naming and the given flag.

WTW="${WTW:-./wtw}"
FORCE_FLAG="${1:?Usage: test_rm_force_custom_naming.sh <-f|--force>}"

# Use a distinct branch per flag so we can run this test twice (-f and --force)
if [ "$FORCE_FLAG" = "-f" ]; then
  BRANCH="force-rm-custom-f"
else
  BRANCH="force-rm-custom-double"
fi

mkdir -p ~/.config/worktree-workflow
cat > ~/.config/worktree-workflow/config.json << 'EOF'
{
  "editor": "echo",
  "auto_open_editor": false,
  "naming": {
    "worktree_dir_suffix": "-wt",
    "branch_separator": "_"
  }
}
EOF

cd /tmp/test-project

git branch "$BRANCH" 2>/dev/null || true
$WTW create "$BRANCH"

WORKTREE_DIR="/tmp/test-project-wt/test-project_$BRANCH"
test -d "$WORKTREE_DIR" || (echo "FAIL: worktree not at expected path: $WORKTREE_DIR" && exit 1)

echo "dirty" > "$WORKTREE_DIR/untracked.txt"

if $WTW rm "$BRANCH" 2>&1 | grep -q "modified or untracked"; then
  echo "OK: normal remove detected dirty worktree"
fi

$WTW rm $FORCE_FLAG "$BRANCH"

test ! -d "$WORKTREE_DIR" || (echo "FAIL: worktree still exists after force remove" && exit 1)

echo "PASS: force rm with custom naming ($FORCE_FLAG)"
