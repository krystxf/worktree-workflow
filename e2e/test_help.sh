#!/usr/bin/env bash
set -euo pipefail

WTW="${WTW:-./wtw}"

# Capture output to avoid SIGPIPE with pipefail
HELP_OUTPUT=$($WTW --help 2>&1)
echo "$HELP_OUTPUT" | grep -q "create" || (echo "FAIL: help missing create" && exit 1)
echo "$HELP_OUTPUT" | grep -q "ps" || (echo "FAIL: help missing ps" && exit 1)
echo "$HELP_OUTPUT" | grep -q "rm" || (echo "FAIL: help missing rm" && exit 1)
echo "$HELP_OUTPUT" | grep -q "init" || (echo "FAIL: help missing init" && exit 1)

CREATE_HELP=$($WTW create --help 2>&1)
echo "$CREATE_HELP" | grep -q "branch" || (echo "FAIL: create help missing branch" && exit 1)

RM_HELP=$($WTW rm --help 2>&1)
echo "$RM_HELP" | grep -q "force" || (echo "FAIL: rm help missing force flag" && exit 1)

echo "PASS: help output"
