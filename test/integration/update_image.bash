#!/usr/bin/env bash

set -euo pipefail

: "${GITHUB_TOKEN}"

cd "$(dirname "$0")"

ts="$(date +%s)"

if [[ "$#" -eq 0 || "$1" != "-n" ]]; then
    docker build . -t sleeping-beauty-integration-test:latest
    ../../docker/run_in_docker.bash sleeping-beauty-integration-test:latest ./test/integration/run.bash
fi

image="ghcr.io/radian-software/sleeping-beauty-integration-test-ci:${ts}"

echo "${GITHUB_TOKEN}" | docker login ghcr.io -u radian-software --password-stdin

docker tag sleeping-beauty-integration-test:latest "${image}"
docker push "${image}"

echo "${image}" >./ci_image
