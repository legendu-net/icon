if status is-interactive
    # Commands to run in interactive sessions can go here
    set -g EDITOR nvim
    set -g VISUAL nvim
    abbr --add mvi mv -i
    abbr --add cpi cp -ir
    abbr --add blog ./blog.py
    abbr --add fcs fzf_cs
    abbr --add fcd fzf_cs
    abbr --add fbat fzf_bat
    abbr --add frgvim fzf_ripgrep_nvim
    abbr --add fhist fzf_history
    abbr --add fh fzf_history
end
