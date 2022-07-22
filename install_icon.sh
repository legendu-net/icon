#!/usr/bin/env bash

function install_icon.usage() {
    cat << EOF
NAME
    /scripts/sys/install_icon.sh - Download and install icon to /usr/local/bin/.
SYNTAX 
    /scripts/sys/install_icon.sh [-h]
EOF
}

function install_icon() {
    if [[ $1 == "-h" ]]; then
        install_icon.usage
        return 0
    fi
    if [[ $# == 1 && "$1" != "-h" || $# > 1 ]]; then
        install_icon.usage
        return 1
    fi
    local URL=https://github.com/legendu-net/icon/releases
    local VERSION=$(basename $(curl -sL -o /dev/null -w %{url_effective} $URL/latest))
    case $(uname -m) in
        x84_64 )
            local ARCH=amd64
            ;;
        arm64 )
            local ARCH=arm64
            ;;
        *)
            echo "The architecture is not supported!"
            return 2
            ;;
    esac
    curl -sSL $URL/download/$VERSION/icon-$VERSION-$(uname)-${ARCH}.tar.gz -o /tmp/icon.tar.gz
    tar zxf /tmp/icon.tar.gz -C /usr/local/bin/
    chmod +x /usr/local/bin/icon
}

if [[ "$0" == ${BASH_SOURCE[0]} ]]; then
    install_icon $@
fi

