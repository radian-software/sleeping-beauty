#!/usr/bin/env bash

set -euo pipefail

# For load tests (bash resets it by default)
ulimit -n "$(ulimit -n -H)"

repo_dir="$(cd "$(dirname "$0")/../.." && pwd)"
export PATH="${repo_dir}:$PATH"

cd "$(dirname "$0")"
echo >&2 "Running integration tests..."
go test ./cases "$@"
