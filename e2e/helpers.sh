# helpers.sh — assertion helpers for e2e tests
# Source this at the top of every test: source "$(dirname "$0")/helpers.sh"

set -euo pipefail

WTW="${WTW:-./wtw}"

_fail() { echo "FAIL: $1"; exit 1; }

assert_dir_exists()     { test -d "$1"  || _fail "$1 does not exist"; }
assert_dir_not_exists() { test ! -d "$1" || _fail "$1 should not exist"; }
assert_file_exists()    { test -f "$1"  || _fail "$1 does not exist"; }
assert_file_not_exists() { test ! -f "$1" || _fail "$1 should not exist"; }

assert_file_contents() {
  local file="$1" expected="$2"
  local actual
  actual=$(cat "$file")
  test "$actual" = "$expected" || _fail "$file: expected '$expected', got '$actual'"
}

assert_file_contains() {
  local file="$1" pattern="$2"
  grep -q "$pattern" "$file" || _fail "$file does not contain '$pattern'"
}

assert_output_contains() {
  local output="$1" pattern="$2"
  echo "$output" | grep -q "$pattern" || _fail "output does not contain '$pattern': $output"
}

assert_branch() {
  local dir="$1" expected="$2"
  local actual
  actual=$(git -C "$dir" branch --show-current)
  test "$actual" = "$expected" || _fail "expected branch '$expected', got '$actual'"
}

assert_command_fails() {
  if eval "$@" >/dev/null 2>&1; then
    _fail "expected command to fail: $*"
  fi
}

assert_command_succeeds() {
  if ! eval "$@" >/dev/null 2>&1; then
    _fail "expected command to succeed: $*"
  fi
}

# Run wtw and capture combined output
run_wtw() { $WTW "$@" 2>&1; }

# Run wtw expecting failure, return output
run_wtw_fail() { $WTW "$@" 2>&1 || true; }

pass() { echo "PASS: $1"; }

# Write global config with auto_open_editor=false for CI
write_global_config() {
  mkdir -p ~/.config/worktree-workflow
  cat > ~/.config/worktree-workflow/config.json << 'EOF'
{
  "editor": "echo",
  "auto_open_editor": false
}
EOF
}
