#!/usr/bin/env bash
set -euo pipefail

WTW="${WTW:-./wtw}"

$WTW --help | grep -q "create" || (echo "FAIL: help missing create" && exit 1)
$WTW --help | grep -q "ps" || (echo "FAIL: help missing ps" && exit 1)
$WTW --help | grep -q "rm" || (echo "FAIL: help missing rm" && exit 1)
$WTW --help | grep -q "init" || (echo "FAIL: help missing init" && exit 1)
$WTW create --help | grep -q "branch" || (echo "FAIL: create help missing branch" && exit 1)
$WTW rm --help | grep -q "force" || (echo "FAIL: rm help missing force flag" && exit 1)

echo "PASS: help output"
