#!/usr/bin/env bash

set -euo pipefail

repo_dir="$(cd "$(dirname "$0")/../.." && pwd)"
export PATH="${repo_dir}:$PATH"

cd "$(dirname "$0")"
go test ./cases "$@"
