#!/bin/bash

function build() {
    TARGET_OS_ARCH=$(go_os_arch $1 $2)
    TARGET_OS_ARCH_ERR=$?
    if [ $TARGET_OS_ARCH_ERR -ne 0 ]; then
        echo "$TARGET_OS_ARCH"
        return 1
    fi

    export GOOS=`echo "$TARGET_OS_ARCH" | awk '{print $1}'`
    export GOARCH=`echo "$TARGET_OS_ARCH" | awk '{print $2}'`

    go get -v ./src/cmd/blockd
    go build -v -o bin/blockd ./src/cmd/blockd
}

if [[ -z "${BLOCKYARD_SCRIPT+x}" ]]; then
    build $@
fi
