{pkgs, ...}: {
  channel = "stable-24.11";
  packages = with pkgs; [
    ncurses
    fd
    moreutils
    bash-completion
    gitui
    delta
    neovim
    ripgrep
    rm-improved
    bat
    fzf
    util-linux
    go
    uv
    dos2unix
  ];
  env = {};
  services.docker.enable = true;
  idx = {
    # check extensions on https://open-vsx.org/
    extensions = [
      "golang.go"
      #"vscodevim.vim"
      "asvetliakov.vscode-neovim"
    ];
    workspace = {
      #onCreate = {
      #}
      onStart = {
      };
    };
    # Enable previews and customize configuration
    previews = {};
  };
}
