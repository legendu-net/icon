{pkgs, ...}: {
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
      "vscodevim.vim"
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
