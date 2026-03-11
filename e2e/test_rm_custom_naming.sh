#!/usr/bin/env bash
set -euo pipefail

# Usage: test_rm_custom_naming.sh [rm|remove]
# Runs the same custom-naming remove test with the given subcommand (default: rm).

WTW="${WTW:-./wtw}"
CMD="${1:-rm}"

# Distinct branch per subcommand so we can run with both rm and remove
if [ "$CMD" = "rm" ]; then
  BRANCH="rm-custom-naming-test"
else
  BRANCH="remove-custom-naming-test"
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

$WTW $CMD "$BRANCH"

test ! -d "$WORKTREE_DIR" || (echo "FAIL: worktree dir still exists" && exit 1)
OUTPUT=$(git worktree list)
echo "$OUTPUT" | grep -q "$BRANCH" && (echo "FAIL: branch still in git worktree list" && exit 1)

echo "PASS: rm with custom naming ($CMD)"
