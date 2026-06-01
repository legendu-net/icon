#!/usr/bin/env bash

function install_icon.usage() {
    cat <<EOF
Usage: $0 [options]

Options:
  -h        Display this help message
  -d <dir>  Specify the installation directory (default: /usr/local/bin/)
  -v <version>  The version (latest by default) to install.
    Notice that a valid version starts with "v".
EOF
}

function _release_exists() {
    local version="$1"
    local tag_url="https://github.com/legendu-net/icon/releases/tag"
    curl -sfI -o /dev/null "$tag_url/$version"
}

function _get_alt_version() {
    local version="$1"
    if [[ "$version" =~ ^v ]]; then
        echo "${version#v}"
    else
        echo "v${version}"
    fi
}

function install_icon() {
    local install_dir="/usr/local/bin/"
    local version=""
    while getopts "hd:v:" opt; do
        case $opt in
            h)
                install_icon.usage
                return 0
                ;;
            d) install_dir="$OPTARG" ;;
            v) version="$OPTARG" ;;
            \?)
                install_icon.usage
                return 1
                ;;
        esac
    done
    mkdir -p "$install_dir"
    local URL=https://github.com/legendu-net/icon/releases
    if [[ "$version" == "" ]]; then
        echo "Parsing the latest version ..."
        version=$(basename "$(curl -sL -o /dev/null -w "%{url_effective}" "$URL/latest")")
    elif ! _release_exists "$version"; then
        local version_alt
        version_alt=$(_get_alt_version "$version")
        if _release_exists "$version_alt"; then
            echo "The release tag $version does not exist! Did you mean $version_alt?"
        else
            echo "The release tag $version does not exist!"
        fi
        return 3
    fi
    local arch
    arch="$(uname -m)"
    case "$arch" in
        x86_64 | amd64)
            arch=amd64
            ;;
        aarch64 | arm64)
            arch=arm64
            ;;
        *)
            echo "The architecture $arch is not supported!"
            return 2
            ;;
    esac
    local url_download
    url_download=$URL/download/$version/icon-$version-$(uname)-${arch}.tar.gz
    local output
    output=/tmp/icon_$(date +%Y%m%d%H%M%S).tar.gz
    echo "Downloading $url_download to $output ..."
    if ! curl -sSL "$url_download" -o "$output"; then
        echo "Failed to download $url_download to $output!"
        return 4
    fi
    echo "Installing icon into $install_dir ..."
    if ! tar -zxf "$output" -C "$install_dir"; then
        echo "Failed to extract $output into $install_dir!"
        return 5
    fi
    if ! chmod +x "$install_dir/icon"; then
        echo "Failed to make $install_dir/icon executable!"
        return 6
    fi
    echo "icon has been successfully installed into $install_dir."
    add_script_ldc "$install_dir"
}

function add_script_ldc() {
    local install_dir=$1
    echo "Creating script $install_dir/ldc ..."
    cat <<EOF >"$install_dir/ldc"
#/usr/bin/env bash
icon ldc \$@
EOF
    chmod +x "$install_dir/ldc"
}

if [[ "${BASH_SOURCE[0]}" == "" || "${BASH_SOURCE[0]}" == "$0" ]]; then
    install_icon "$@"
fi
