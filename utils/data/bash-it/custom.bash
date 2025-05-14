function _trace_dir_upwards.usage {
    cat << EOF
Trace a directory upwards until the name is found.
Syntax: trace_dir_upwards <dir> <name>
EOF
}

function trace_dir_upwards {
    if [[ "$1" == "-h" || "$1" == "--help" ]]; then
        _trace_dir_upwards.usage
        return 0
    fi
    local dir="$1"
    local name="$2"
    local stem
    stem=$(basename "$dir")
    while [[ "$stem" != "$name" ]]; do 
        if [[ "$stem" == "/" || "$stem" == "" ]]; then
            echo -e "$name is not found in $dir!\n"
            return 1
        fi
        dir=$(dirname "$dir")
        stem=$(basename "$dir")
    done
    echo "$dir"
}

