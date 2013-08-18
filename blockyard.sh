#!/bin/bash

SCRIPT_SRC="${PWD}"

if [[ "$#" < 1 ]]; then
   echo 1>&2  'missing argument [command]'
   exit 1
fi

for init_script in "${SCRIPT_SRC}"/*.sh; do
    if [[ $(basename "$init_script") != "blockyard.sh" ]]; then
        source "$init_script"
    fi
done
