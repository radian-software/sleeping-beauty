#!/usr/bin/env bash

set -eou pipefail

groupadd -g "$(stat -c %g "$PWD")" -o -p '!' -r docker
useradd -u "$(stat -c %u "$PWD")" -g "$(stat -c %g "$PWD")" -o -p '!' -m -N -l -s "${SHELL}" -G sudo docker

chown docker:docker /go-cache
mkdir -p /home/docker/go/pkg
ln -sT /go-cache /home/docker/go/pkg/mod

tee /etc/sudoers.d/90-docker >/dev/null <<"EOF"
%sudo ALL=(ALL:ALL) NOPASSWD: ALL
EOF

runuser -u docker touch /home/docker/.sudo_as_admin_successful
exec runuser -u docker -- "$@"
