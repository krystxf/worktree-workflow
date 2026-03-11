#!/usr/bin/env bash
source "$(dirname "$0")/helpers.sh"

# --local (and --local -f / --force) should fail outside a git repo
cd /tmp
assert_command_fails $WTW init --local
OUTPUT=$(run_wtw_fail init --local)
assert_output_contains "$OUTPUT" "not a git repository"

assert_command_fails $WTW init --local -f
OUTPUT=$(run_wtw_fail init --local -f)
assert_output_contains "$OUTPUT" "not a git repository"

assert_command_fails $WTW init --local --force
OUTPUT=$(run_wtw_fail init --local --force)
assert_output_contains "$OUTPUT" "not a git repository"

write_global_config

# --- init --local -f: write default project config in a repo ---
rm -rf /tmp/initlocal-force-project /tmp/initlocal-force-project--worktrees
mkdir -p /tmp/initlocal-force-project
cd /tmp/initlocal-force-project
git init
echo "hello" > file.txt
git add -A
git commit -m "init"
git branch init-local-test

assert_file_not_exists .worktree-workflow.json

run_wtw init --local -f >/dev/null

assert_file_exists .worktree-workflow.json
assert_file_contains .worktree-workflow.json '"sync_ignored"'
assert_file_contains .worktree-workflow.json '"sync_excludes"'
assert_file_contains .worktree-workflow.json '"post_copy_hooks"'

run_wtw create init-local-test >/dev/null
WORKTREE="/tmp/initlocal-force-project--worktrees/initlocal-force-project--init-local-test"
assert_dir_exists "$WORKTREE"

pass "init --local -f writes default project config"

# --- init --local config (simulated) + create: sync/hooks ---
rm -rf /tmp/initlocal-project /tmp/initlocal-project--worktrees
mkdir -p /tmp/initlocal-project
cd /tmp/initlocal-project
git init
echo "hello" > file.txt
echo "node_modules/" > .gitignore
echo ".env" >> .gitignore
git add -A
git commit -m "init"
git branch test-branch

# Create gitignored files
echo "SECRET=123" > .env
mkdir -p node_modules/pkg
echo "{}" > node_modules/pkg/package.json

# Simulate what `wtw init --local` produces
cat > .worktree-workflow.json << 'EOF'
{
  "sync_ignored": true,
  "sync_excludes": ["node_modules"],
  "post_copy_hooks": ["touch .initialized"]
}
EOF

run_wtw create test-branch

WORKTREE="/tmp/initlocal-project--worktrees/initlocal-project--test-branch"
assert_dir_exists "$WORKTREE"
assert_file_exists "$WORKTREE/.env"
assert_dir_not_exists "$WORKTREE/node_modules"
assert_file_exists "$WORKTREE/.initialized"

pass "init --local config (sync + hooks)"
