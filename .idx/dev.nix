{pkgs, ...}: {
  channel = "stable-24.11";
  packages = [
    pkgs.neovim
    pkgs.rm-improved
    pkgs.go
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
