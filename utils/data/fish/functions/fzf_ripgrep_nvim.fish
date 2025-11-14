function _fzf_ripgrep_nvim_usage
    echo "Leverage fzf as the UI to search for files by content using ripgrep,
preview it using bat, and open selections using NeoVim.
Syntax: fzf.ripgrep.nvim [-h] [dir]
Args:
    dir: The directory (default to .) under which to search for files.
"
end

function fzf_ripgrep_nvim
    argparse h/help -- $argv
    if set -q _flag_help
        _fzf_ripgrep_nvim_usage
        return 0
    end

    set -l search_path .
    if test (count $argv) -gt 0
      set search_path "$argv"
    end

    set -l RELOAD "reload:rg --column --color=always --smart-case {q} $search_path || :"

    # The OPENER needs careful handling to ensure fish variables are not expanded prematurely
    # and the shell logic within fzf's execute works correctly.
    # FZF's execute expects a shell command. Fish's string handling is different.
    # We'll build the command directly as a string that FZF will execute.
    set -l OPENER 'if [[ $FZF_SELECT_COUNT -eq 0 ]]; then
            nvim {1} +{2}     # No selection. Open the current line in Vim.
          else
            nvim +cw -q {+f}  # Build quickfix list for the selected items.
          fi'
          #set -l OPENER 'if test $FZF_SELECT_COUNT -eq 0
          #              nvim {1} +{2}
          #         else
          #              nvim +cw -q {+f}
          #         end'

    #fzf -m --disabled --ansi --multi \
      #    --bind "home:$RELOAD" --bind "change:$RELOAD" \
      #  --bind "enter:execute[ $OPENER ]" \
      #  --bind "ctrl-o:execute[ $OPENER ]" \
      #  --bind 'alt-a:select-all,alt-d:deselect-all,ctrl-/:toggle-preview' \
      #  --delimiter : \
      #  --preview 'bat --style=full --color=always --highlight-line {2} {1}' \
      #  --preview-window 'right:40%'
    #fzf -m --disabled --ansi --multi \
        #    --bind "home:$RELOAD" --bind "change:$RELOAD" \
        #--bind "enter:execute:$OPENER" \
        #--bind "ctrl-o:execute:$OPENER" \
        #--bind 'alt-a:select-all,alt-d:deselect-all,ctrl-/:toggle-preview' \
        #--delimiter : \
        #--preview 'bat --style=full --color=always --highlight-line {2} {1}' \
        #--preview-window 'right:40%'
    fzf --disabled --ansi --multi \
      --bind "start:$RELOAD" --bind "change:$RELOAD" \
      --bind "enter:become:$OPENER" \
      --bind "ctrl-o:execute:$OPENER" \
      --bind 'alt-a:select-all,alt-d:deselect-all,ctrl-/:toggle-preview' \
      --delimiter : \
      --preview 'bat --style=full --color=always --highlight-line {2} {1}' \
      --preview-window '~4,+{2}+4/3,<80(up)' \
      --query "$argv"
end

