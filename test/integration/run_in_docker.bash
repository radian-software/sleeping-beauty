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

mkdir -p "${repo_dir}/.cache/gopkg"
docker run -it --rm --init -v "${repo_dir}:/src:ro" -w /src \
    -v "${repo_dir}/.cache/gopkg:/go-cache" \
    --entrypoint=/src/test/integration/pid2.bash \
    sleeping-beauty-integration-test:latest "$@"
