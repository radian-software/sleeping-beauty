#!/usr/bin/env bash

set -euo pipefail

repo_dir="$(cd "$(dirname "$0")/../.." && pwd)"
export PATH="${repo_dir}:$PATH"

test_dir="$(cd "$(dirname "$0")" && pwd)"
num_failed=0
while read case; do
    echo "TEST ${case} ..."
    if go run "${test_dir}/cases/${case}.go" 2>&1 | sed 's/^/    /'; then
        echo "TEST ${case} ... passed"
    else
        echo "TEST ${case} ... failed"
        num_failed+=1
    fi
done < <(ls "${test_dir}/cases" | grep '\.go$' | sed 's/\.go$//')

if (("${num_failed}" > 0)); then
    exit 1
fi
