# bash completion for icon                                 -*- shell-script -*-

__icon_debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE:-} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}

# Homebrew on Macs have version 1.3 of bash-completion which doesn't include
# _init_completion. This is a very minimal version of that function.
__icon_init_completion()
{
    COMPREPLY=()
    _get_comp_words_by_ref "$@" cur prev words cword
}

__icon_index_of_word()
{
    local w word=$1
    shift
    index=0
    for w in "$@"; do
        [[ $w = "$word" ]] && return
        index=$((index+1))
    done
    index=-1
}

__icon_contains_word()
{
    local w word=$1; shift
    for w in "$@"; do
        [[ $w = "$word" ]] && return
    done
    return 1
}

__icon_handle_go_custom_completion()
{
    __icon_debug "${FUNCNAME[0]}: cur is ${cur}, words[*] is ${words[*]}, #words[@] is ${#words[@]}"

    local shellCompDirectiveError=1
    local shellCompDirectiveNoSpace=2
    local shellCompDirectiveNoFileComp=4
    local shellCompDirectiveFilterFileExt=8
    local shellCompDirectiveFilterDirs=16

    local out requestComp lastParam lastChar comp directive args

    # Prepare the command to request completions for the program.
    # Calling ${words[0]} instead of directly icon allows to handle aliases
    args=("${words[@]:1}")
    # Disable ActiveHelp which is not supported for bash completion v1
    requestComp="ICON_ACTIVE_HELP=0 ${words[0]} __completeNoDesc ${args[*]}"

    lastParam=${words[$((${#words[@]}-1))]}
    lastChar=${lastParam:$((${#lastParam}-1)):1}
    __icon_debug "${FUNCNAME[0]}: lastParam ${lastParam}, lastChar ${lastChar}"

    if [ -z "${cur}" ] && [ "${lastChar}" != "=" ]; then
        # If the last parameter is complete (there is a space following it)
        # We add an extra empty parameter so we can indicate this to the go method.
        __icon_debug "${FUNCNAME[0]}: Adding extra empty parameter"
        requestComp="${requestComp} \"\""
    fi

    __icon_debug "${FUNCNAME[0]}: calling ${requestComp}"
    # Use eval to handle any environment variables and such
    out=$(eval "${requestComp}" 2>/dev/null)

    # Extract the directive integer at the very end of the output following a colon (:)
    directive=${out##*:}
    # Remove the directive
    out=${out%:*}
    if [ "${directive}" = "${out}" ]; then
        # There is not directive specified
        directive=0
    fi
    __icon_debug "${FUNCNAME[0]}: the completion directive is: ${directive}"
    __icon_debug "${FUNCNAME[0]}: the completions are: ${out}"

    if [ $((directive & shellCompDirectiveError)) -ne 0 ]; then
        # Error code.  No completion.
        __icon_debug "${FUNCNAME[0]}: received error from custom completion go code"
        return
    else
        if [ $((directive & shellCompDirectiveNoSpace)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __icon_debug "${FUNCNAME[0]}: activating no space"
                compopt -o nospace
            fi
        fi
        if [ $((directive & shellCompDirectiveNoFileComp)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __icon_debug "${FUNCNAME[0]}: activating no file completion"
                compopt +o default
            fi
        fi
    fi

    if [ $((directive & shellCompDirectiveFilterFileExt)) -ne 0 ]; then
        # File extension filtering
        local fullFilter filter filteringCmd
        # Do not use quotes around the $out variable or else newline
        # characters will be kept.
        for filter in ${out}; do
            fullFilter+="$filter|"
        done

        filteringCmd="_filedir $fullFilter"
        __icon_debug "File filtering command: $filteringCmd"
        $filteringCmd
    elif [ $((directive & shellCompDirectiveFilterDirs)) -ne 0 ]; then
        # File completion for directories only
        local subdir
        # Use printf to strip any trailing newline
        subdir=$(printf "%s" "${out}")
        if [ -n "$subdir" ]; then
            __icon_debug "Listing directories in $subdir"
            __icon_handle_subdirs_in_dir_flag "$subdir"
        else
            __icon_debug "Listing directories in ."
            _filedir -d
        fi
    else
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${out}" -- "$cur")
    fi
}

__icon_handle_reply()
{
    __icon_debug "${FUNCNAME[0]}"
    local comp
    case $cur in
        -*)
            if [[ $(type -t compopt) = "builtin" ]]; then
                compopt -o nospace
            fi
            local allflags
            if [ ${#must_have_one_flag[@]} -ne 0 ]; then
                allflags=("${must_have_one_flag[@]}")
            else
                allflags=("${flags[*]} ${two_word_flags[*]}")
            fi
            while IFS='' read -r comp; do
                COMPREPLY+=("$comp")
            done < <(compgen -W "${allflags[*]}" -- "$cur")
            if [[ $(type -t compopt) = "builtin" ]]; then
                [[ "${COMPREPLY[0]}" == *= ]] || compopt +o nospace
            fi

            # complete after --flag=abc
            if [[ $cur == *=* ]]; then
                if [[ $(type -t compopt) = "builtin" ]]; then
                    compopt +o nospace
                fi

                local index flag
                flag="${cur%=*}"
                __icon_index_of_word "${flag}" "${flags_with_completion[@]}"
                COMPREPLY=()
                if [[ ${index} -ge 0 ]]; then
                    PREFIX=""
                    cur="${cur#*=}"
                    ${flags_completion[${index}]}
                    if [ -n "${ZSH_VERSION:-}" ]; then
                        # zsh completion needs --flag= prefix
                        eval "COMPREPLY=( \"\${COMPREPLY[@]/#/${flag}=}\" )"
                    fi
                fi
            fi

            if [[ -z "${flag_parsing_disabled}" ]]; then
                # If flag parsing is enabled, we have completed the flags and can return.
                # If flag parsing is disabled, we may not know all (or any) of the flags, so we fallthrough
                # to possibly call handle_go_custom_completion.
                return 0;
            fi
            ;;
    esac

    # check if we are handling a flag with special work handling
    local index
    __icon_index_of_word "${prev}" "${flags_with_completion[@]}"
    if [[ ${index} -ge 0 ]]; then
        ${flags_completion[${index}]}
        return
    fi

    # we are parsing a flag and don't have a special handler, no completion
    if [[ ${cur} != "${words[cword]}" ]]; then
        return
    fi

    local completions
    completions=("${commands[@]}")
    if [[ ${#must_have_one_noun[@]} -ne 0 ]]; then
        completions+=("${must_have_one_noun[@]}")
    elif [[ -n "${has_completion_function}" ]]; then
        # if a go completion function is provided, defer to that function
        __icon_handle_go_custom_completion
    fi
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions+=("${must_have_one_flag[@]}")
    fi
    while IFS='' read -r comp; do
        COMPREPLY+=("$comp")
    done < <(compgen -W "${completions[*]}" -- "$cur")

    if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${noun_aliases[*]}" -- "$cur")
    fi

    if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
        if declare -F __icon_custom_func >/dev/null; then
            # try command name qualified custom func
            __icon_custom_func
        else
            # otherwise fall back to unqualified for compatibility
            declare -F __custom_func >/dev/null && __custom_func
        fi
    fi

    # available in bash-completion >= 2, not always present on macOS
    if declare -F __ltrim_colon_completions >/dev/null; then
        __ltrim_colon_completions "$cur"
    fi

    # If there is only 1 completion and it is a flag with an = it will be completed
    # but we don't want a space after the =
    if [[ "${#COMPREPLY[@]}" -eq "1" ]] && [[ $(type -t compopt) = "builtin" ]] && [[ "${COMPREPLY[0]}" == --*= ]]; then
       compopt -o nospace
    fi
}

# The arguments should be in the form "ext1|ext2|extn"
__icon_handle_filename_extension_flag()
{
    local ext="$1"
    _filedir "@(${ext})"
}

__icon_handle_subdirs_in_dir_flag()
{
    local dir="$1"
    pushd "${dir}" >/dev/null 2>&1 && _filedir -d && popd >/dev/null 2>&1 || return
}

__icon_handle_flag()
{
    __icon_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    # if a command required a flag, and we found it, unset must_have_one_flag()
    local flagname=${words[c]}
    local flagvalue=""
    # if the word contained an =
    if [[ ${words[c]} == *"="* ]]; then
        flagvalue=${flagname#*=} # take in as flagvalue after the =
        flagname=${flagname%=*} # strip everything after the =
        flagname="${flagname}=" # but put the = back
    fi
    __icon_debug "${FUNCNAME[0]}: looking for ${flagname}"
    if __icon_contains_word "${flagname}" "${must_have_one_flag[@]}"; then
        must_have_one_flag=()
    fi

    # if you set a flag which only applies to this command, don't show subcommands
    if __icon_contains_word "${flagname}" "${local_nonpersistent_flags[@]}"; then
      commands=()
    fi

    # keep flag value with flagname as flaghash
    # flaghash variable is an associative array which is only supported in bash > 3.
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        if [ -n "${flagvalue}" ] ; then
            flaghash[${flagname}]=${flagvalue}
        elif [ -n "${words[ $((c+1)) ]}" ] ; then
            flaghash[${flagname}]=${words[ $((c+1)) ]}
        else
            flaghash[${flagname}]="true" # pad "true" for bool flag
        fi
    fi

    # skip the argument to a two word flag
    if [[ ${words[c]} != *"="* ]] && __icon_contains_word "${words[c]}" "${two_word_flags[@]}"; then
        __icon_debug "${FUNCNAME[0]}: found a flag ${words[c]}, skip the next argument"
        c=$((c+1))
        # if we are looking for a flags value, don't show commands
        if [[ $c -eq $cword ]]; then
            commands=()
        fi
    fi

    c=$((c+1))

}

__icon_handle_noun()
{
    __icon_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    if __icon_contains_word "${words[c]}" "${must_have_one_noun[@]}"; then
        must_have_one_noun=()
    elif __icon_contains_word "${words[c]}" "${noun_aliases[@]}"; then
        must_have_one_noun=()
    fi

    nouns+=("${words[c]}")
    c=$((c+1))
}

__icon_handle_command()
{
    __icon_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    local next_command
    if [[ -n ${last_command} ]]; then
        next_command="_${last_command}_${words[c]//:/__}"
    else
        if [[ $c -eq 0 ]]; then
            next_command="_icon_root_command"
        else
            next_command="_${words[c]//:/__}"
        fi
    fi
    c=$((c+1))
    __icon_debug "${FUNCNAME[0]}: looking for ${next_command}"
    declare -F "$next_command" >/dev/null && $next_command
}

__icon_handle_word()
{
    if [[ $c -ge $cword ]]; then
        __icon_handle_reply
        return
    fi
    __icon_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"
    if [[ "${words[c]}" == -* ]]; then
        __icon_handle_flag
    elif __icon_contains_word "${words[c]}" "${commands[@]}"; then
        __icon_handle_command
    elif [[ $c -eq 0 ]]; then
        __icon_handle_command
    elif __icon_contains_word "${words[c]}" "${command_aliases[@]}"; then
        # aliashash variable is an associative array which is only supported in bash > 3.
        if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
            words[c]=${aliashash[${words[c]}]}
            __icon_handle_command
        else
            __icon_handle_noun
        fi
    else
        __icon_handle_noun
    fi
    __icon_handle_word
}

_icon_alacritty()
{
    last_command="icon_alacritty"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--install")
    flags+=("-i")
    flags+=("--uninstall")
    flags+=("--yes")
    flags+=("-y")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_atuin()
{
    last_command="icon_atuin"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--install")
    flags+=("-i")
    flags+=("--uninstall")
    flags+=("--yes")
    flags+=("-y")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_bash_it()
{
    last_command="icon_bash_it"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--install")
    flags+=("-i")
    flags+=("--uninstall")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_bytehound()
{
    last_command="icon_bytehound"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--install")
    flags+=("-i")
    flags+=("--uninstall")
    flags+=("-u")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_completion()
{
    last_command="icon_completion"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")

    must_have_one_flag=()
    must_have_one_noun=()
    must_have_one_noun+=("bash")
    must_have_one_noun+=("fish")
    must_have_one_noun+=("zsh")
    noun_aliases=()
}

_icon_docker()
{
    last_command="icon_docker"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--install")
    flags+=("-i")
    flags+=("--uninstall")
    flags+=("-u")
    flags+=("--user-to-docker=")
    two_word_flags+=("--user-to-docker")
    flags+=("--yes")
    flags+=("-y")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_download_github_release()
{
    last_command="icon_download_github_release"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--KWD=")
    two_word_flags+=("--KWD")
    two_word_flags+=("-K")
    flags+=("--kwd=")
    two_word_flags+=("--kwd")
    two_word_flags+=("-k")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--repo=")
    two_word_flags+=("--repo")
    two_word_flags+=("-r")
    flags+=("--version=")
    two_word_flags+=("--version")
    two_word_flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--kwd=")
    must_have_one_flag+=("-k")
    must_have_one_flag+=("--output=")
    must_have_one_flag+=("-o")
    must_have_one_flag+=("--repo=")
    must_have_one_flag+=("-r")
    must_have_one_noun=()
    noun_aliases=()
}

_icon_firenvim()
{
    last_command="icon_firenvim"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--install")
    flags+=("-i")
    flags+=("--uninstall")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_ganymede()
{
    last_command="icon_ganymede"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--extra-pip-options=")
    two_word_flags+=("--extra-pip-options")
    flags+=("--install")
    flags+=("-i")
    flags+=("--profile-dir=")
    two_word_flags+=("--profile-dir")
    flags+=("--python=")
    two_word_flags+=("--python")
    flags+=("--sudo")
    flags+=("--uninstall")
    flags+=("--user")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_git()
{
    last_command="icon_git"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--append")
    flags+=("-a")
    flags+=("--config")
    flags+=("-c")
    flags+=("--dest-dir=")
    two_word_flags+=("--dest-dir")
    two_word_flags+=("-d")
    flags+=("--git=")
    two_word_flags+=("--git")
    flags+=("--gitui")
    flags+=("--install")
    flags+=("-i")
    flags+=("--lang=")
    two_word_flags+=("--lang")
    two_word_flags+=("-l")
    flags+=("--proxy=")
    two_word_flags+=("--proxy")
    flags+=("--uninstall")
    flags+=("--user-email=")
    two_word_flags+=("--user-email")
    two_word_flags+=("-e")
    flags+=("--user-name=")
    two_word_flags+=("--user-name")
    two_word_flags+=("-n")
    flags+=("--yes")
    flags+=("-y")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_golang()
{
    last_command="icon_golang"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--install")
    flags+=("-i")
    flags+=("--uninstall")
    flags+=("-u")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_helix()
{
    last_command="icon_helix"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--install")
    flags+=("-i")
    flags+=("--uninstall")
    flags+=("--yes")
    flags+=("-y")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_help()
{
    last_command="icon_help"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    has_completion_function=1
    noun_aliases=()
}

_icon_hyper()
{
    last_command="icon_hyper"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--install")
    flags+=("-i")
    flags+=("--uninstall")
    flags+=("--version=")
    two_word_flags+=("--version")
    two_word_flags+=("-v")
    flags+=("--yes")
    flags+=("-y")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_ipython()
{
    last_command="icon_ipython"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--extra-pip-options=")
    two_word_flags+=("--extra-pip-options")
    flags+=("--install")
    flags+=("-i")
    flags+=("--profile-dir=")
    two_word_flags+=("--profile-dir")
    flags+=("--python=")
    two_word_flags+=("--python")
    flags+=("--sudo")
    flags+=("--uninstall")
    flags+=("--user")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_jupyter_book()
{
    last_command="icon_jupyter_book"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--extra-pip-options=")
    two_word_flags+=("--extra-pip-options")
    flags+=("--install")
    flags+=("-i")
    flags+=("--python=")
    two_word_flags+=("--python")
    flags+=("--uninstall")
    flags+=("--user")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_jupyterlab_vim()
{
    last_command="icon_jupyterlab_vim"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--disable")
    flags+=("--enable")
    flags+=("--extra-pip-options=")
    two_word_flags+=("--extra-pip-options")
    flags+=("--install")
    flags+=("-i")
    flags+=("--python=")
    two_word_flags+=("--python")
    flags+=("--sudo")
    flags+=("--uninstall")
    flags+=("--user")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_keepassxc()
{
    last_command="icon_keepassxc"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--install")
    flags+=("-i")
    flags+=("--uninstall")
    flags+=("--version=")
    two_word_flags+=("--version")
    two_word_flags+=("-v")
    flags+=("--yes")
    flags+=("-y")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_ldc()
{
    last_command="icon_ldc"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--detach")
    flags+=("-d")
    flags+=("--docker-in-docker")
    flags+=("--dry-run")
    flags+=("--extra-port-mappings=")
    two_word_flags+=("--extra-port-mappings")
    flags+=("--mount-home")
    flags+=("-m")
    flags+=("--password=")
    two_word_flags+=("--password")
    two_word_flags+=("-P")
    flags+=("--port=")
    two_word_flags+=("--port")
    two_word_flags+=("-p")
    flags+=("--user=")
    two_word_flags+=("--user")
    two_word_flags+=("-u")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_neovim()
{
    last_command="icon_neovim"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--install")
    flags+=("-i")
    flags+=("--uninstall")
    flags+=("--yes")
    flags+=("-y")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_nushell()
{
    last_command="icon_nushell"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--install")
    flags+=("-i")
    flags+=("--uninstall")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_pytorch()
{
    last_command="icon_pytorch"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--cuda-version=")
    two_word_flags+=("--cuda-version")
    flags+=("--extra-pip-options=")
    two_word_flags+=("--extra-pip-options")
    flags+=("--install")
    flags+=("-i")
    flags+=("--python=")
    two_word_flags+=("--python")
    flags+=("--uninstall")
    flags+=("--user")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_pytype()
{
    last_command="icon_pytype"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--dest-dir=")
    two_word_flags+=("--dest-dir")
    two_word_flags+=("-d")
    flags+=("--extra-pip-options=")
    two_word_flags+=("--extra-pip-options")
    flags+=("--install")
    flags+=("-i")
    flags+=("--python=")
    two_word_flags+=("--python")
    flags+=("--uninstall")
    flags+=("--user")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_rip()
{
    last_command="icon_rip"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--install")
    flags+=("-i")
    flags+=("--uninstall")
    flags+=("--yes")
    flags+=("-y")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_rust()
{
    last_command="icon_rust"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--cargo-home=")
    two_word_flags+=("--cargo-home")
    flags+=("--config")
    flags+=("-c")
    flags+=("--install")
    flags+=("-i")
    flags+=("--link-to-dir=")
    two_word_flags+=("--link-to-dir")
    flags+=("--path")
    flags+=("-p")
    flags+=("--rustup-home=")
    two_word_flags+=("--rustup-home")
    flags+=("--toolchain=")
    two_word_flags+=("--toolchain")
    flags+=("--uninstall")
    flags+=("-u")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_spark()
{
    last_command="icon_spark"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--directory=")
    two_word_flags+=("--directory")
    two_word_flags+=("-d")
    flags+=("--hadoop-version=")
    two_word_flags+=("--hadoop-version")
    flags+=("--install")
    flags+=("-i")
    flags+=("--spark-version=")
    two_word_flags+=("--spark-version")
    flags+=("--uninstall")
    flags+=("-u")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_ssh_client()
{
    last_command="icon_ssh_client"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--install")
    flags+=("-i")
    flags+=("--uninstall")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_ssh_server()
{
    last_command="icon_ssh_server"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--install")
    flags+=("-i")
    flags+=("--uninstall")
    flags+=("--yes")
    flags+=("-y")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_update()
{
    last_command="icon_update"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_version()
{
    last_command="icon_version"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_visual_studio_code()
{
    last_command="icon_visual_studio_code"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config")
    flags+=("-c")
    flags+=("--install")
    flags+=("-i")
    flags+=("--uninstall")
    flags+=("--user-dir=")
    two_word_flags+=("--user-dir")
    two_word_flags+=("-d")
    flags+=("--yes")
    flags+=("-y")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_zellij()
{
    last_command="icon_zellij"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--bin-dir=")
    two_word_flags+=("--bin-dir")
    flags+=("--config")
    flags+=("-c")
    flags+=("--extra-pip-options=")
    two_word_flags+=("--extra-pip-options")
    flags+=("--install")
    flags+=("-i")
    flags+=("--python=")
    two_word_flags+=("--python")
    flags+=("--sudo")
    flags+=("--uninstall")
    flags+=("--user")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_icon_root_command()
{
    last_command="icon"

    command_aliases=()

    commands=()
    commands+=("alacritty")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("alac")
        aliashash["alac"]="alacritty"
    fi
    commands+=("atuin")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("atuin")
        aliashash["atuin"]="atuin"
    fi
    commands+=("bash_it")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("bashit")
        aliashash["bashit"]="bash_it"
        command_aliases+=("bit")
        aliashash["bit"]="bash_it"
    fi
    commands+=("bytehound")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("bh")
        aliashash["bh"]="bytehound"
        command_aliases+=("bhound")
        aliashash["bhound"]="bytehound"
        command_aliases+=("byteh")
        aliashash["byteh"]="bytehound"
    fi
    commands+=("completion")
    commands+=("docker")
    commands+=("download_github_release")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("download_github")
        aliashash["download_github"]="download_github_release"
        command_aliases+=("from_github")
        aliashash["from_github"]="download_github_release"
        command_aliases+=("github_release")
        aliashash["github_release"]="download_github_release"
    fi
    commands+=("firenvim")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("fvim")
        aliashash["fvim"]="firenvim"
    fi
    commands+=("ganymede")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("gmd")
        aliashash["gmd"]="ganymede"
    fi
    commands+=("git")
    commands+=("golang")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("go")
        aliashash["go"]="golang"
    fi
    commands+=("helix")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("nvim")
        aliashash["nvim"]="helix"
    fi
    commands+=("help")
    commands+=("hyper")
    commands+=("ipython")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("ipy")
        aliashash["ipy"]="ipython"
    fi
    commands+=("jupyter_book")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("jb")
        aliashash["jb"]="jupyter_book"
        command_aliases+=("jbook")
        aliashash["jbook"]="jupyter_book"
    fi
    commands+=("jupyterlab_vim")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("jlab_vim")
        aliashash["jlab_vim"]="jupyterlab_vim"
        command_aliases+=("jlabvim")
        aliashash["jlabvim"]="jupyterlab_vim"
        command_aliases+=("jvim")
        aliashash["jvim"]="jupyterlab_vim"
    fi
    commands+=("keepassxc")
    commands+=("ldc")
    commands+=("neovim")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("nvim")
        aliashash["nvim"]="neovim"
    fi
    commands+=("nushell")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("nu")
        aliashash["nu"]="nushell"
    fi
    commands+=("pytorch")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("torch")
        aliashash["torch"]="pytorch"
    fi
    commands+=("pytype")
    commands+=("rip")
    commands+=("rust")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("cargo")
        aliashash["cargo"]="rust"
        command_aliases+=("rustup")
        aliashash["rustup"]="rust"
    fi
    commands+=("spark")
    commands+=("ssh_client")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("sshc")
        aliashash["sshc"]="ssh_client"
    fi
    commands+=("ssh_server")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("sshs")
        aliashash["sshs"]="ssh_server"
    fi
    commands+=("update")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("upd")
        aliashash["upd"]="update"
    fi
    commands+=("version")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("v")
        aliashash["v"]="version"
    fi
    commands+=("visual_studio_code")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("code")
        aliashash["code"]="visual_studio_code"
        command_aliases+=("vscode")
        aliashash["vscode"]="visual_studio_code"
    fi
    commands+=("zellij")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("z")
        aliashash["z"]="zellij"
        command_aliases+=("zj")
        aliashash["zj"]="zellij"
    fi

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

__start_icon()
{
    local cur prev words cword split
    declare -A flaghash 2>/dev/null || :
    declare -A aliashash 2>/dev/null || :
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -s || return
    else
        __icon_init_completion -n "=" || return
    fi

    local c=0
    local flag_parsing_disabled=
    local flags=()
    local two_word_flags=()
    local local_nonpersistent_flags=()
    local flags_with_completion=()
    local flags_completion=()
    local commands=("icon")
    local command_aliases=()
    local must_have_one_flag=()
    local must_have_one_noun=()
    local has_completion_function=""
    local last_command=""
    local nouns=()
    local noun_aliases=()

    __icon_handle_word
}

if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_icon icon
else
    complete -o default -o nospace -F __start_icon icon
fi

# ex: ts=4 sw=4 et filetype=sh
