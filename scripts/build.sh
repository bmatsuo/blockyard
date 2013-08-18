#!/bin/bash

function build() {
    go get -v ./src/cmd/blockd
    go build -v -o bin/blockd ./src/cmd/blockd
}

if [[ -z "${BLOCKYARD_SCRIPT+x}" ]]; then
    build $@
fi
