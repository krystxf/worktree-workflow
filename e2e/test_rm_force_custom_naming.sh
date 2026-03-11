#!/usr/bin/env bash
# Usage: test_rm_force_custom_naming.sh <-f|--force>
source "$(dirname "$0")/helpers.sh"

FORCE_FLAG="${1:?Usage: test_rm_force_custom_naming.sh <-f|--force>}"

# Distinct branch per flag
if [ "$FORCE_FLAG" = "-f" ]; then BRANCH="force-rm-custom-f"; else BRANCH="force-rm-custom-double"; fi

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

# Make dirty
echo "dirty" > "$WORKTREE/untracked.txt"

# Normal remove should fail
OUTPUT=$(run_wtw_fail rm "$BRANCH")
assert_output_contains "$OUTPUT" "modified/untracked files"

# Force remove should succeed
run_wtw rm $FORCE_FLAG "$BRANCH"
assert_dir_not_exists "$WORKTREE"

pass "force rm with custom naming ($FORCE_FLAG)"
