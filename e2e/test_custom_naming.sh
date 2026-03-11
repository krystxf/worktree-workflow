#!/usr/bin/env bash
source "$(dirname "$0")/helpers.sh"

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
run_wtw create custom-naming-test

assert_dir_exists "/tmp/test-project-wt/test-project_custom-naming-test"

pass "custom naming config"
