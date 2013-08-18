#!/bin/bash

SCRIPT_SRC="${PWD}/scripts"

BLOCKYARD_SCRIPT="$0"
for init_script in "${SCRIPT_SRC}"/*.sh; do
    if [[ $(basename "$init_script") != "blockyard.sh" ]]; then
        source "$init_script"
    fi
done

if [ 0 -eq $# ]; then 
    echo 'missing argument [command]'
    exit 1
fi

ARGS=("$@")
COMMAND="$1"
if [ "${COMMAND}" == "help" ]; then
    echo "you won't find any of that here... >_>"
    exit 1
fi

TARGET="${SCRIPT_SRC}/${COMMAND}.sh"
if [[ -x ${TARGET} ]]; then
    TARGET_ARGS=${ARGS[@]:1}

    "${TARGET}" ${TARGET_ARGS}
else
    echo "not a command: $COMMAND"
    exit 1
fi
