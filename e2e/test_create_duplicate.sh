#!/usr/bin/env bash
source "$(dirname "$0")/helpers.sh"

cd /tmp/test-project
assert_command_fails "$WTW create feature-one"

pass "duplicate worktree fails gracefully"
