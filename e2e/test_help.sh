#!/usr/bin/env bash
# Usage: test_help.sh [remove]
# No arg: full help checks. With "remove": test alias help.
source "$(dirname "$0")/helpers.sh"

ALIAS="${1:-}"

if [ -n "$ALIAS" ]; then
  RM_HELP=$(run_wtw "$ALIAS" --help)
  assert_output_contains "$RM_HELP" "force"
  pass "help output ($ALIAS)"
  exit 0
fi

HELP=$(run_wtw --help)
assert_output_contains "$HELP" "create"
assert_output_contains "$HELP" "ls"
assert_output_contains "$HELP" "rm"
assert_output_contains "$HELP" "init"

CREATE_HELP=$(run_wtw create --help)
assert_output_contains "$CREATE_HELP" "branch"

RM_HELP=$(run_wtw rm --help)
assert_output_contains "$RM_HELP" "force"

INIT_HELP=$(run_wtw init --help)
assert_output_contains "$INIT_HELP" "local"

pass "help output"
