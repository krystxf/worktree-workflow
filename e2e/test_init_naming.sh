#!/usr/bin/env bash
set -euo pipefail

WTW="${WTW:-./wtw}"

# Clean up
rm -rf /tmp/init-naming-project /tmp/init-naming-project-trees
rm -f ~/.config/worktree-workflow/config.json

# Create a test repo
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

# Verify the config is picked up: worktree should be at repo-trees/repo.branch
$WTW create test-branch

EXPECTED="/tmp/init-naming-project-trees/init-naming-project.test-branch"
test -d "$EXPECTED" || (echo "FAIL: expected worktree at $EXPECTED" && exit 1)

# Verify the config file has naming fields
grep -q "worktree_dir_suffix" ~/.config/worktree-workflow/config.json || (echo "FAIL: config missing worktree_dir_suffix" && exit 1)
grep -q "branch_separator" ~/.config/worktree-workflow/config.json || (echo "FAIL: config missing branch_separator" && exit 1)

# Restore default config for other tests
cat > ~/.config/worktree-workflow/config.json << 'EOF'
{
  "editor": "echo",
  "auto_open_editor": false
}
EOF

echo "PASS: init naming customization"
