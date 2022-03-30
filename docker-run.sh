#!/bin/sh

IMAGE="${2:-ghcr.io/kubeclarity/kubeclarity}"
DIR="${3:-$HOME}"

docker run --privileged --device /dev/fuse -v /var/run/docker.sock:/var/run/docker.sock -v ${DIR}:/tmp --rm ${IMAGE} $1
