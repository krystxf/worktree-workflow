#!/usr/bin/env bash
set -euo pipefail

# Runs all e2e tests in order.
# Usage: WTW=./wtw ./e2e/run.sh

DIR="$(cd "$(dirname "$0")" && pwd)"
export WTW="${WTW:-./wtw}"

echo "=== E2E Tests ==="
echo "Binary: $WTW"
echo ""

PASS=0
FAIL=0

# Setup must run first
bash "$DIR/setup.sh"

# Tests that depend on setup state (order matters)
ORDERED_TESTS=(
  test_create.sh
  test_sync.sh
  test_create_second.sh
  test_create_duplicate.sh
  test_list.sh
  test_rm.sh
  test_rm_force.sh
  test_rm_nonexistent.sh
  test_hooks.sh
  test_hooks_fail.sh
)

# Independent tests
INDEPENDENT_TESTS=(
  test_custom_naming.sh
  test_defaults.sh
  test_no_sync.sh
  test_not_git_repo.sh
  test_help.sh
)

for test in "${ORDERED_TESTS[@]}" "${INDEPENDENT_TESTS[@]}"; do
  if bash "$DIR/$test"; then
    PASS=$((PASS + 1))
  else
    echo "FAIL: $test"
    FAIL=$((FAIL + 1))
  fi
done

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
