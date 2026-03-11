#!/usr/bin/env bash
set -euo pipefail

# Runs all e2e tests in order.
# Usage: WTW=./wtw ./e2e/run.sh

DIR="$(cd "$(dirname "$0")" && pwd)"
export WTW="$(cd "$(dirname "${WTW:-./wtw}")" && pwd)/$(basename "${WTW:-./wtw}")"

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# Colorize PASS: (green) and FAIL: (red) in test output
color_output() {
  sed -e "s/^PASS:/${GREEN}PASS:${NC}/g" -e "s/^FAIL:/${RED}FAIL:${NC}/g"
}

echo "=== E2E Tests ==="
echo "Binary: $WTW"
echo ""

PASS=0
FAIL=0

# Setup must run first
bash "$DIR/setup.sh"

# Tests that depend on setup state (order matters).
# Entries with a space are "script arg(s)" to run both flag/alias variants without duplicating test code.
ORDERED_TESTS=(
  test_create.sh
  test_sync.sh
  test_create_second.sh
  test_create_duplicate.sh
  "test_list.sh"
  "test_list.sh ls"
  "test_list.sh list"
  "test_rm.sh rm"
  "test_rm.sh remove"
  "test_rm_force.sh -f"
  "test_rm_force.sh --force"
  test_force_rm.sh
  test_rm_nonexistent.sh
  test_hooks.sh
  test_hooks_fail.sh
)

# Independent tests (same script+args pattern for alias/flag variants)
INDEPENDENT_TESTS=(
  test_custom_naming.sh
  "test_rm_custom_naming.sh rm"
  "test_rm_custom_naming.sh remove"
  "test_rm_force_custom_naming.sh -f"
  "test_rm_force_custom_naming.sh --force"
  test_defaults.sh
  test_no_sync.sh
  test_not_git_repo.sh
  "test_help.sh"
  "test_help.sh remove"
  test_create_new_branch.sh
  test_force_create.sh
  test_init_naming.sh
  test_init_local.sh
)

run_one() {
  local script="$1"
  shift
  local output
  output=$(bash "$DIR/$script" "$@" 2>&1)
  local ret=$?
  echo "$output" | color_output
  return $ret
}

for test in "${ORDERED_TESTS[@]}" "${INDEPENDENT_TESTS[@]}"; do
  if [[ "$test" == *" "* ]]; then
    script="${test%% *}"
    args="${test#* }"
    if run_one "$script" $args; then
      PASS=$((PASS + 1))
    else
      echo -e "${RED}FAIL: $test${NC}"
      FAIL=$((FAIL + 1))
    fi
  else
    if run_one "$test"; then
      PASS=$((PASS + 1))
    else
      echo -e "${RED}FAIL: $test${NC}"
      FAIL=$((FAIL + 1))
    fi
  fi
done

echo ""
if [ "$FAIL" -gt 0 ]; then
  echo -e "=== Results: ${GREEN}$PASS passed${NC}, ${RED}$FAIL failed${NC} ==="
else
  echo -e "=== Results: ${GREEN}$PASS passed${NC}, $FAIL failed ==="
fi

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
