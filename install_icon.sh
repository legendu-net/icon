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
    echo "Parsing the latest version ..."
    local URL=https://github.com/legendu-net/icon/releases
    local VERSION=$(basename $(curl -sL -o /dev/null -w %{url_effective} $URL/latest))
    local ARCH="$(uname -m)"
    case "$ARCH" in
        x86_64 )
            ARCH=amd64
            ;;
        arm64 )
            ARCH=arm64
            ;;
        *)
            echo "The architecture $ARCH is not supported!"
            return 2
            ;;
    esac
    echo "Downloading icon ..."
    curl -sSL $URL/download/$VERSION/icon-$VERSION-$(uname)-${ARCH}.tar.gz -o /tmp/icon.tar.gz
    echo "Installing icon ..."
    tar zxf /tmp/icon.tar.gz -C /usr/local/bin/
    chmod +x /usr/local/bin/icon
    add_script_ldc
}

function add_script_ldc() {
    echo "Creating script /usr/local/bin/ldc ..."
    cat << EOF > /usr/local/bin/ldc
#/usr/bin/env bash
icon ldc \$@
EOF
    chmod +x /usr/local/bin/ldc
}

if [[ "${BASH_SOURCE[0]}" == "" || "${BASH_SOURCE[0]}" == "$0" ]]; then
    install_icon $@
fi