function _cs.usage() {
    cat << EOF
Enter a directory and display its content.
Syntax: cs dir
EOF
}

function cs() {
    if [ "$1" == "-h" ]; then
        _cs.usage
        return 0
    fi
    local dir="$@"
    if [[ -f "$dir" ]]; then
      dir="$(dirname "$dir")"
    fi
    if [[ "$dir" == "" ]]; then
      dir="$HOME"
    fi
    cd "$dir" 
    if [[ "$?" != 0 ]]; then
      echo "Failed to cd into $dir!"
      return $?
    fi
    ls --color=auto
}

function _fzf.cs.usage {
    cat << EOF
Search for a directory using fzf, cd into it, and run ls.
Syntax: fzf.cs [-h] [dir]
Args:
    dir: The directory (default to .) under which to search for sub directories.
EOF
}

function _check_fdfind {
  if [[ "$(which fdfind)" == "" ]]; then
    echo "fd-find executable is not found! Please install it first!"
    return 1
  fi
}

function fzf.cs {
  if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    _fzf.cs.usage
    return 0
  fi
  _check_fdfind
  local dir=.
  if [[ $# > 0 ]]; then
    dir="$@"
  fi
  cd "$(fdfind --type d --print0 . $dir | fzf --read0)"
  ls
}

alias fcs=fzf.cs
alias fcd=fzf.cs


function _fzf.bat.usage {
    cat << EOF
Search for files using fzf and preview it using bat.
Syntax: fzf.bat [-h] [dir]
Args:
    dir: The directory (default to .) under which to search for files.
EOF
}

function fzf.bat {
  if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    _fzf.bat.usage
    return 0
  fi
  _check_fdfind
  local dir=.
  if [[ $# > 0 ]]; then
    dir="$@"
  fi
  o="$(fdfind --type f --print0 . $dir | fzf --read0 --preview 'bat --color=always {}')"
  echo $o
  nvim $o
}

alias fbat=fzf.bat


function _fzf.ripgrep.nvim.usage {
    cat << EOF
Leverage fzf as the UI to search for files by content using ripgrep, \
preview it using bat, and open selections using NeoVim.
Syntax: fzf.ripgrep.nvim [-h] [dir]
Args:
    dir: The directory (default to .) under which to search for files.
EOF
}

function fzf.ripgrep.nvim (
  if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    _fzf.ripgrep.nvim.usage
    return 0
  fi
  local dir=.
  if [[ $# > 0 ]]; then
    dir="$@"
  fi
  RELOAD="reload:rg --column --color=always --smart-case {q} $dir || :"
  OPENER='if [[ $FZF_SELECT_COUNT -eq 0 ]]; then
            nvim {1} +{2}     # No selection. Open the current line in Vim.
          else
            nvim +cw -q {+f}  # Build quickfix list for the selected items.
          fi'
  fzf --disabled --ansi --multi \
      --bind "start:$RELOAD" --bind "change:$RELOAD" \
      --bind "enter:become:$OPENER" \
      --bind "ctrl-o:execute:$OPENER" \
      --bind 'alt-a:select-all,alt-d:deselect-all,ctrl-/:toggle-preview' \
      --delimiter : \
      --preview 'bat --style=full --color=always --highlight-line {2} {1}' \
      --preview-window '~4,+{2}+4/3,<80(up)'
)
  
alias frgvim=fzf.ripgrep.nvim


function _fzf.history.usage {
    cat << EOF
Search for files using fzf and preview it using bat.
Syntax: fzf.bat [-h] [dir]
Args:
    dir: The directory (default to .) under which to search for files.
EOF
}

function fzf.history {
  if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    _fzf.history.usage
    return 0
  fi
  local command=$(fc -ln 1 | sed 's/^[[:space:]]*//;s/[[:space:]]*$//' | fzf | vipe)
  echo $command
  history -s "$command"
  eval "$command"
}

alias fhist=fzf.history
alias fh=fzf.history

