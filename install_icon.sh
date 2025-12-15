#!/usr/bin/env bash

function install_icon.usage() {
    cat << EOF
Usage: $0 [options]

Options:
  -h        Display this help message
  -d <dir>  Specify the installation directory (default: /usr/local/bin/)
  -v <version>  The version (latest by default) to install.
EOF
}

function install_icon() {
    local install_dir="/usr/local/bin/"
    local version=""
    while getopts "hd:v:" opt; do
        case $opt in
            h) install_icon.usage; return 0 ;;
            d) install_dir="$OPTARG" ;;
            v) version="$OPTARG" ;;
            \?) install_icon.usage; return 1 ;;
        esac
    done
    mkdir -p "$install_dir"
    add_script_ldc "$install_dir"
    echo "Parsing the latest version ..."
    local URL=https://github.com/legendu-net/icon/releases
    if [[ "$version" == "" ]]; then
        version=$(basename $(curl -sL -o /dev/null -w %{url_effective} $URL/latest))
    fi
    local arch="$(uname -m)"
    case "$arch" in
        x86_64 )
            arch=amd64
            ;;
        aarch64 )
            arch=arm64
            ;;
        *)
            echo "The architecture $arch is not supported!"
            return 2
            ;;
    esac
    local url_download=$URL/download/$version/icon-$version-$(uname)-${arch}.tar.gz
    local output=/tmp/icon_$(date +%Y%m%d%H%M%S).tar.gz
    echo "Downloading $url_download to $output ..."
    curl -sSL $url_download -o $output
    echo "Installing icon ..."
    tar -zxvf $output -C "$install_dir"
    chmod +x "$install_dir/icon"
    echo "icon has been installed successfully."
}

function add_script_ldc() {
    local install_dir=$1
    echo "Creating script $install_dir/ldc ..."
    cat << EOF > "$install_dir/ldc"
#/usr/bin/env bash
icon ldc \$@
EOF
    chmod +x "$install_dir/ldc"
}

if [[ "${BASH_SOURCE[0]}" == "" || "${BASH_SOURCE[0]}" == "$0" ]]; then
    install_icon $@
fi
