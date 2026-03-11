#!/usr/bin/env bash
source "$(dirname "$0")/helpers.sh"

cd /tmp/test-project
assert_command_fails "$WTW rm nonexistent-branch"

pass "non-existent worktree removal fails"
