#!/usr/bin/env bash
set -euo pipefail

# Usage: test_help.sh [remove]
# No arg: full help checks (create, ls, rm, init, create --help, rm --help).
# With arg "remove": verify $WTW remove --help (alias of rm) shows force flag.

WTW="${WTW:-./wtw}"
ALIAS="${1:-}"

if [ -n "$ALIAS" ]; then
  # Test alias help (e.g. wtw remove --help same as wtw rm --help)
  RM_HELP=$($WTW "$ALIAS" --help 2>&1)
  echo "$RM_HELP" | grep -q "force" || (echo "FAIL: $ALIAS help missing force flag" && exit 1)
  echo "PASS: help output ($ALIAS)"
  exit 0
fi

# Capture output to avoid SIGPIPE with pipefail
HELP_OUTPUT=$($WTW --help 2>&1)
echo "$HELP_OUTPUT" | grep -q "create" || (echo "FAIL: help missing create" && exit 1)
echo "$HELP_OUTPUT" | grep -q "ls" || (echo "FAIL: help missing ls" && exit 1)
echo "$HELP_OUTPUT" | grep -q "rm" || (echo "FAIL: help missing rm" && exit 1)
echo "$HELP_OUTPUT" | grep -q "init" || (echo "FAIL: help missing init" && exit 1)

CREATE_HELP=$($WTW create --help 2>&1)
echo "$CREATE_HELP" | grep -q "branch" || (echo "FAIL: create help missing branch" && exit 1)

RM_HELP=$($WTW rm --help 2>&1)
echo "$RM_HELP" | grep -q "force" || (echo "FAIL: rm help missing force flag" && exit 1)

echo "PASS: help output"
