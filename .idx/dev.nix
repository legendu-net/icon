{pkgs, ...}: {
  packages = [
    pkgs.neovim
    pkgs.python311
    pkgs.python311Packages.pip
    pkgs.poetry
  ];
  env = {};
  idx = {
    # check extensions on https://open-vsx.org/
    extensions = [
      "vscodevim.vim"
      "ms-python.python"
      "ms-python.debugpy"
    ];
    workspace = {
      #onCreate = {
      #}
      onStart = {
        poetry-project = ''
        poetry config --local virtualenvs.in-project true
        poetry install
        '';
      };
    };
    # Enable previews and customize configuration
    previews = {};
  };
}
