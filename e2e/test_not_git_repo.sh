#!/usr/bin/env bash
source "$(dirname "$0")/helpers.sh"

cd /tmp

OUTPUT=$(run_wtw_fail create some-branch)
assert_output_contains "$OUTPUT" "not a git repository"

OUTPUT=$(run_wtw_fail ls)
assert_output_contains "$OUTPUT" "not a git repository"

pass "fails outside git repo"
