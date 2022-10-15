#!/usr/bin/env bash

set -euo pipefail

# Don't Docker recursively. This lets you just use this as a wrapper
# script whether you're in Docker or not.
if [[ -n "${DOCKER:-}" ]]; then
    exec "$@"
fi

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
    -v "${repo_dir}/.cache/gopkg:/go-cache" -e DOCKER=1 \
    --entrypoint=/src/test/integration/pid2.bash \
    sleeping-beauty-integration-test:latest "$@"
