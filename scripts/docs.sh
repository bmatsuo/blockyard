#!/bin/bash

function docs() {
    godoc $@
}

if [[ -z "${BLOCKYARD_SCRIPT+x}" ]]; then
    docs $@
fi
