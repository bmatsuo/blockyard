#!/bin/bash

export GOPATH="${PWD}"

function error() {
    echo "$@" 1>&2
}

# prints space separated os and arch for build targets. one or both may be
# empty strings, in which case `go build` will figure it out.
function go_os_arch() {
    TARGET_OS=$1
    TARGET_ARCH=$2

    if [[ -z "$TARGET_OS" ]]; then
        TARGET_OS="$GOOS"
    fi

    # target os must be unixy.
    case "$TARGET_OS" in
        darwin|freebsd|netbsd|openbsd|linux)
            # ok
            ;;
        "")
            # FIXME attempt to uname it out
            ;;
        plan9)
            echo 1>&2 'plan9 is not supported (syslog) *sadface*'
            return 1
            ;;
        windows)
            echo 1>&2 'windows is not supported'
            return 1
            ;;
        *)
            echo 1>&2 "target os not supported: $TARGET_OS"
            return 1
            ;;
    esac

    if [[ -z "$TARGET_ARCH" ]]; then
        TARGET_ARCH="$GOARCH"
    fi

    case "$TARGET_ARCH" in
        arm|386|amd64)
            # ok
            ;;
        "")
            # FIXME attempt to uname it out
            ;;
        *)
            echo 1>&2 "target arch not supported: $TARGET_ARCH"
            return 1
    esac

    if [[ -n "$TARGET_OS" ]] && [[ -n "$TARGET_ARCH" ]]; then
        case "$TARGET_OS $TARGET_ARCH" in
            "darwin 386")
                ;;
            "darwin amd64")
                ;;
            "freebsd 386")
                ;;
            "freebsd amd64")
                ;;
            "linux 386")
                ;;
            "linux amd64")
                ;;
            "linux arm")
                ;;
            "netbsd 386")
                ;;
            "netbsd amd64")
                ;;
            "openbsd 386")
                ;;
            "openbsd amd64")
                ;;
            "plan9 386")
                ;;
            "windows 386")
                ;;
            "windows amd64")
                ;;
            *)
                echo 1>&2 "invalid os-arch combination: $TARGET_OS $TARGET_ARCH"
                return 1
        esac
    fi

    echo "$TARGET_OS $TARGET_ARCH"
    return 0
}
