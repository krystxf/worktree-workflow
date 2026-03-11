#!/usr/bin/env bash
set -euo pipefail

# Sets up a test repo at /tmp/test-project with branches and gitignored files.
# Must be run before other e2e tests.

rm -rf /tmp/test-project /tmp/test-project--worktrees /tmp/test-project-wt
rm -rf /tmp/bare-project /tmp/bare-project--worktrees
rm -rf /tmp/nosync-project /tmp/nosync-project--worktrees
# Set global config that disables auto-open (no editor in CI)
mkdir -p ~/.config/worktree-workflow
cat > ~/.config/worktree-workflow/config.json << 'EOF'
{
  "editor": "echo",
  "auto_open_editor": false
}
EOF

git config --global user.email "test@test.com" 2>/dev/null || true
git config --global user.name "Test" 2>/dev/null || true
git config --global init.defaultBranch main

mkdir -p /tmp/test-project
cd /tmp/test-project
git init
echo "node_modules/" >> .gitignore
echo ".env" >> .gitignore
echo ".env.local" >> .gitignore
echo "dist/" >> .gitignore
echo '{"name": "test-project"}' > package.json
echo "hello" > index.js
git add -A
git commit -m "initial commit"

git branch feature-one
git branch feature-two
git branch bugfix-123

# Add gitignored files
echo "SECRET=abc" > .env
echo "OTHER=xyz" > .env.local
mkdir -p node_modules/fake-pkg
echo "{}" > node_modules/fake-pkg/package.json
mkdir -p dist
echo "built" > dist/index.js

# Make a commit on feature-one
git checkout feature-one
echo "feature one code" > feature.js
git add feature.js
git commit -m "add feature one"
git checkout main

echo "PASS: test repo created"
