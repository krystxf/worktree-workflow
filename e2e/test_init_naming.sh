#!/usr/bin/env bash
source "$(dirname "$0")/helpers.sh"

rm -rf /tmp/init-naming-project /tmp/init-naming-project-trees

mkdir -p /tmp/init-naming-project
cd /tmp/init-naming-project
git init
echo "hello" > file.txt
git add -A
git commit -m "init"
git branch test-branch

# Simulate what `wtw init` produces with custom naming
mkdir -p ~/.config/worktree-workflow
cat > ~/.config/worktree-workflow/config.json << 'EOF'
{
  "editor": "echo",
  "auto_open_editor": false,
  "naming": {
    "worktree_dir_suffix": "-trees",
    "branch_separator": "."
  }
}
EOF

run_wtw create test-branch

assert_dir_exists "/tmp/init-naming-project-trees/init-naming-project.test-branch"
assert_file_contains ~/.config/worktree-workflow/config.json "worktree_dir_suffix"
assert_file_contains ~/.config/worktree-workflow/config.json "branch_separator"

# Restore default config
write_global_config

pass "init naming customization"
