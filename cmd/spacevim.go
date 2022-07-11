package cmd

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
	//"log"
	"path/filepath"
	//"runtime"
)

func configureSpaceVimForFirenvim(){
	text = `"----------------------------------------------------------------------
if exists('g:started_by_firenvim')
	set guifont=Monaco:h16
endif
	` 
	utils.AppendToTextFile(filepath.Join(utils.UserHomeDir(), ".SpaceVim/init.vim"), text)
}

// Enable/disable true color for SpaceVim.
func configureSpaceVimTrueColor(trueColor bool) {
    file = HOME / ".SpaceVim.d/init.toml"
    with file.open() as fin:
        lines = fin.readlines()
    for idx, line in enumerate(lines):
        if line.strip().startswith("enable_guicolors"):
            if true_color:
                lines[idx] = line.replace("false", "true")
            else:
                lines[idx] = line.replace("true", "false")
    with file.open("w") as fout:
        fout.writelines(lines)
}

def _strip_spacevim(args: Namespace) -> None:
    if not args.strip:
        return
    dir_ = Path.home() / ".SpaceVim/"
    paths = [
        ".git",
        ".SpaceVim.d/",
        ".ci/",
        ".github/",
        "docker/",
        "docs/",
        "wiki/",
        ".editorconfig",
        ".gitignore",
        "CODE_OF_CONDUCT.md",
        "CONTRIBUTING.cn.md",
        "CONTRIBUTING.md",
        "Makefile",
        "README.cn.md",
        "README.md",
        "codecov.yml",
    ]
    for path in paths:
        path = dir_ / path
        if path.is_file():
            try:
                path.unlink()
            except FileNotFoundError:
                pass
        else:
            try:
                shutil.rmtree(path)
            except FileNotFoundError:
                pass




// Install and configure SpaceVim.
func spaceVim(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
        utils.RunCmd("curl -sLf https://spacevim.org/install.sh | bash")
        _strip_spacevim(args)
        if shutil.which("nvim"):
            run_cmd('nvim --headless +"call dein#install()" +qall')
        if not args.no_lsp:
            cmd = f"{args.pip_install} python-language-server[all] pyls-mypy"
            # npm install -g bash-language-server javascript-typescript-langserver
            run_cmd(cmd)
	}
	if utils.GetBoolFlag(cmd, "config") {
        # configure .SpaceVim
        des_dir = HOME / ".SpaceVim"
        os.makedirs(des_dir, exist_ok=True)
        shutil.copy2(BASE_DIR / "SpaceVim/SpaceVim/init.vim", des_dir)
        # configure .SpaceVim.d
        des_dir = HOME / ".SpaceVim.d"
        os.makedirs(des_dir, exist_ok=True)
        shutil.copy2(BASE_DIR / "SpaceVim/SpaceVim.d/init.toml", des_dir)
        shutil.copy2(BASE_DIR / "SpaceVim/SpaceVim.d/vimrc", des_dir)
        # -----------------------------------------------------------
        _svim_true_color(args.true_colors)
        #_svim_for_firenvim()
	}
	if utils.GetBoolFlag(cmd, "config") {
        run_cmd("curl -sLf https://spacevim.org/install.sh | bash -s -- --uninstall")
	}
}

var spaceVimCmd = &cobra.Command{
	Use:     "spacevim",
	Aliases: []string{"svim"},
	Short:   "Install and configure SpaceVim.",
	//Args:  cobra.ExactArgs(1),
	Run: spaceVim,
}

func init() {
	spaceVimCmd.Flags().BoolP("install", "i", false, "If specified, install SpaceVim.")
	spaceVimCmd.Flags().Bool("uninstall", false, "If specified, uninstall SpaceVim.")
	spaceVimCmd.Flags().BoolP("config", "c", false, "If specified, configure SpaceVim.")
	rootCmd.AddCommand(spaceVimCmd)
    subparser.add_argument(
        "--enable-true-colors",
        dest="true_colors",
        action="store_true",
        default=None,
        help="Enable true color (default true) for SpaceVim."
    )
    subparser.add_argument(
        "--disable-true-colors",
        dest="true_colors",
        action="store_false",
        help="Disable true color (default true) for SpaceVim."
    )
    subparser.add_argument(
        "--no-lsp",
        dest="no_lsp",
        action="store_true",
        help="Disable true color (default true) for SpaceVim."
    )
    subparser.add_argument(
        "--strip",
        dest="strip",
        action="store_true",
        help='Strip unnecessary files from "~/.SpaceVim".'
    )
    option_pip_bundle(subparser)
}
