#!/usr/bin/env bash
# Usage: test_rm_custom_naming.sh [rm|remove]
source "$(dirname "$0")/helpers.sh"

CMD="${1:-rm}"

# Distinct branch per subcommand so both rm and remove can run
if [ "$CMD" = "rm" ]; then BRANCH="rm-custom-naming-test"; else BRANCH="remove-custom-naming-test"; fi

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
run_wtw create "$BRANCH"

WORKTREE="/tmp/test-project-wt/test-project_$BRANCH"
assert_dir_exists "$WORKTREE"

run_wtw $CMD "$BRANCH"

assert_dir_not_exists "$WORKTREE"
OUTPUT=$(git worktree list)
echo "$OUTPUT" | grep -q "$BRANCH" && _fail "$BRANCH still in git worktree list"

pass "rm with custom naming ($CMD)"
