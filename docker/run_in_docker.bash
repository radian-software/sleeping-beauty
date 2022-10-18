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

repo_dir="$(cd "$(dirname "$0")/.." && pwd)"

if (("$#" == 0)); then
    echo >&2 "usage: run_in_docker.bash IMAGE[:TAG] ARG..."
    exit 1
fi

if (("$#" == 1)); then
    set -- "$1" bash
fi

it=()
if [[ -t 1 ]]; then
    it+=(-it)
fi

mkdir -p "${repo_dir}/.cache/gopkg"
docker run "${it[@]}" --rm --init -v "${repo_dir}:/src:ro" -w /src \
    -v "${repo_dir}/.cache/gopkg:/go-cache" -e DOCKER=1 \
    --entrypoint=/src/docker/pid2.bash "$@"
