#!/usr/bin/env bash
set -euo pipefail

cd /tmp/test-project
OUTPUT=$(git worktree list)
echo "$OUTPUT"

echo "$OUTPUT" | grep -q "feature-one" || (echo "FAIL: feature-one not in list" && exit 1)
echo "$OUTPUT" | grep -q "feature-two" || (echo "FAIL: feature-two not in list" && exit 1)

echo "PASS: worktree list"
