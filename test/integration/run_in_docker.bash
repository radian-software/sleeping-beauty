#!/usr/bin/env bash

set -euo pipefail

docker() {
    if [[ "${OSTYPE:-}" != darwin* ]] && [[ "${EUID}" != 0 ]]; then
        command -- sudo -- docker "$@"
    else
        command -- docker "$@"
    fi
}

repo_dir="$(cd "$(dirname "$0")/../.." && pwd)"

if (("$#" == 0)); then
    set -- /src/test/integration/run.bash
fi

docker run -it --rm --init -v "${repo_dir}:/src:ro" \
    sleeping-beauty-integration-test:latest "$@"
