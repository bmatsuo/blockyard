#!/bin/bash

function build() {
    export GOPATH=${PWD}:${GOPATH}
    go get -v ./src/cmd/blockd
    go build -v -o bin/blockd ./src/cmd/blockd
}

if [[ -z "${BLOCKYARD_SCRIPT+x}" ]]; then
    build $@
fi
