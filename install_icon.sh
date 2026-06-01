#!/usr/bin/env bash

function _install_icon_usage() {
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
    if [[ -z "$version" ]]; then
        return 1
    fi
    local tag_url="https://github.com/legendu-net/icon/releases/tag"
    curl -sfIL -o /dev/null "$tag_url/$version"
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
    local OPTIND=1
    while getopts "hd:v:" opt; do
        case $opt in
            h)
                _install_icon_usage
                return 0
                ;;
            d) install_dir="$OPTARG" ;;
            v) version="$OPTARG" ;;
            \?)
                _install_icon_usage
                return 1
                ;;
        esac
    done
    mkdir -p "$install_dir"
    local URL=https://github.com/legendu-net/icon/releases
    if [[ "$version" == "" ]]; then
        echo "Parsing the latest version ..."
        local latest_url
        if ! latest_url=$(curl -sfL -o /dev/null -w "%{url_effective}" "$URL/latest") ||
            [[ "$(basename "$latest_url")" == "latest" ]]; then
            echo "Failed to resolve the latest version from $URL/latest!"
            return 8
        fi
        version="$(basename "$latest_url")"
    elif ! _release_exists "$version"; then
        local version_alt
        version_alt="$(_get_alt_version "$version")"
        if _release_exists "$version_alt"; then
            echo "The release tag $version does not exist! Did you mean $version_alt?"
        else
            echo "The release tag $version does not exist!"
        fi
        return 3
    fi
    local os
    os="$(uname | tr '[:upper:]' '[:lower:]')"
    case "$os" in
        linux | darwin) ;;
        *)
            echo "The operating system $os is not supported!"
            return 7
            ;;
    esac
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
    url_download="$URL/download/$version/icon-$version-$os-${arch}.tar.gz"
    local output
    output="$(mktemp "${TMPDIR:-/tmp}/icon.XXXXXXXXXX")" || return 9
    echo "Downloading $url_download to $output ..."
    if ! curl -sSL "$url_download" -o "$output"; then
        echo "Failed to download $url_download to $output!"
        rm -f "$output"
        return 4
    fi
    echo "Installing icon into $install_dir ..."
    if ! tar -zxf "$output" -C "$install_dir"; then
        echo "Failed to extract $output into $install_dir!"
        rm -f "$output"
        return 5
    fi
    rm -f "$output"
    if ! chmod +x "$install_dir/icon"; then
        echo "Failed to make $install_dir/icon executable!"
        return 6
    fi
    echo "icon has been successfully installed into $install_dir."
}

if [[ "${BASH_SOURCE[0]}" == "" || "${BASH_SOURCE[0]}" == "$0" ]]; then
    install_icon "$@"
fi
