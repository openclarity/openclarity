#!/bin/sh

set -euo pipefail

alias goswagger="docker run --rm -it --user $(id -u):$(id -g) -e GOPATH=$GOPATH:/go -v $HOME:$HOME -w $(pwd) quay.io/goswagger/swagger:v0.27.0"

cp server/restapi/configure_kube_clarity_a_p_is.go /tmp/configure_kube_clarity_a_p_is.go
rm -rf server/*
rm -rf client/*
goswagger generate server -f swagger.yaml -t ./server
goswagger generate client -f swagger.yaml -t ./client
cp /tmp/configure_kube_clarity_a_p_is.go server/restapi/configure_kube_clarity_a_p_is.go
