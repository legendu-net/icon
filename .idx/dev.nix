{pkgs, ...}: {
  channel = "stable-25.05";
  packages = with pkgs; [
    fish
    less
    ncurses
    fd
    gitui
    delta
    neovim
    ripgrep
    rm-improved
    bat
    fzf
    util-linux
    go
    golangci-lint
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
