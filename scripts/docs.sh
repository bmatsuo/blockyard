#!/bin/bash

function docs() {
    export GOPATH=${PWD}
    godoc $@
}

if [[ -z "${BLOCKYARD_SCRIPT+x}" ]]; then
    docs $@
fi
