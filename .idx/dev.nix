{pkgs, ...}: {
  channel = "stable-24.11";
  packages = with pkgs; [
    bash-completion
    gitui
    neovim
    ripgrep
    rm-improved
    bat
    fzf
    util-linux
    go
  ];
  env = {};
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
