package cmd

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"

	//"log"
	"path/filepath"
	//"runtime"
)

func configureSpaceVimForFirenvim() {
	text := `"----------------------------------------------------------------------
if exists('g:started_by_firenvim')
	set guifont=Monaco:h16
endif
	`
	utils.AppendToTextFile(filepath.Join(utils.UserHomeDir(), ".SpaceVim/init.vim"), text)
}

// Enable/disable true color for SpaceVim.
func configureSpaceVimTrueColor(trueColor bool) {
	path := filepath.Join(utils.UserHomeDir(), ".SpaceVim.d/init.toml")
	lines := strings.Split(utils.ReadTextFile(path), "\n")
	for idx, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "enable_guicolors") {
			if trueColor {
				lines[idx] = strings.Replace(line, "false", "true", 1)
			} else {
				lines[idx] = strings.Replace(line, "true", "false", 1)
			}
		}
	}
	text := strings.Join(lines, "\n")
	utils.WriteTextFile(path, text)
}

func stripSpaceVim() {
	dir := filepath.Join(utils.UserHomeDir(), ".SpaceVim/")
	paths := []string{
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
	}
	for _, path := range paths {
		path = filepath.Join(dir, path)
		utils.RemoveAll(path)
	}
}

// Install and configure SpaceVim.
func spaceVim(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		pipInstall := utils.BuildPipInstall(cmd)
		utils.RunCmd("curl -sLf https://spacevim.org/install.sh | bash")
		utils.RunCmd(utils.Format("{pip_install} python-lsp-server", map[string]string{
			"pip_install": pipInstall,
		}))
		log.Print("The Python package python-lsp-server is installed! Please make sure pylsp is on the search path!\n")
		// npm install -g bash-language-server javascript-typescript-langserver
		if utils.GetBoolFlag(cmd, "strip") {
			stripSpaceVim()
		}
		if utils.ExistsCommand("nvim") {
			utils.RunCmd(utils.Format(`nvim --headless +"call dein#install()" +qall && {pip_install} pynvim`, map[string]string{
				"pip_install": pipInstall,
			}))
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		home := utils.UserHomeDir()
		// configure .SpaceVim
		desDir := filepath.Join(home, ".SpaceVim")
		utils.MkdirAll(desDir, 0700)
		utils.CopyEmbedFileToDir("data/SpaceVim/SpaceVim/init.vim", desDir)
		// configure .SpaceVim.d
		desDir = filepath.Join(home, ".SpaceVim.d")
		utils.MkdirAll(desDir, 0700)
		utils.CopyEmbedFileToDir("data/SpaceVim/SpaceVim.d/init.toml", desDir)
		utils.CopyEmbedFileToDir("data/SpaceVim/SpaceVim.d/vimrc", desDir)
		// -----------------------------------------------------------
		if utils.GetBoolFlag(cmd, "enable-true-color") {
			configureSpaceVimTrueColor(true)
		}
		if utils.GetBoolFlag(cmd, "disable-true-color") {
			configureSpaceVimTrueColor(false)
		}
		configureSpaceVimForFirenvim()
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		utils.RunCmd("curl -sLf https://spacevim.org/install.sh | bash -s -- --uninstall")
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
	spaceVimCmd.Flags().BoolP("install", "i", false, "Install SpaceVim.")
	spaceVimCmd.Flags().Bool("uninstall", false, "Uninstall SpaceVim.")
	spaceVimCmd.Flags().BoolP("config", "c", false, "Configure SpaceVim.")
	spaceVimCmd.Flags().Bool("enable-true-color", false, "Enable true color support in SpaceVim.")
	spaceVimCmd.Flags().Bool("disable-true-color", false, "Disable true color support in SpaceVim.")
	spaceVimCmd.Flags().Bool("strip", false, "Strip unnecessary files from '~/.SpaceVim/'.")
	utils.AddPythonFlags(spaceVimCmd)
	rootCmd.AddCommand(spaceVimCmd)
}
