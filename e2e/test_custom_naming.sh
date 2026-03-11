#!/usr/bin/env bash
set -euo pipefail

WTW="${WTW:-./wtw}"

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

cat > .worktree-workflow.json << 'EOF'
{
  "sync_ignored": false,
  "sync_excludes": [],
  "post_copy_hooks": []
}
EOF

git branch custom-naming-test 2>/dev/null || true

$WTW create custom-naming-test

WORKTREE_DIR="/tmp/test-project-wt/test-project_custom-naming-test"
test -d "$WORKTREE_DIR" || (echo "FAIL: custom naming worktree not at expected path: $WORKTREE_DIR" && exit 1)

echo "PASS: custom naming config"
