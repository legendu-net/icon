package ide

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
	"log"
	"path/filepath"
	"strings"
)

func configureSpaceVimForFirenvim() {
	script := `
"-------------------------- FireNVim Configurations --------------------------
if exists('g:started_by_firenvim')
  set guifont=Monaco:h16
endif
let g:firenvim_config = { 
    \ 'globalSettings': {
        \ 'alt': 'all',
    \  },
    \ 'localSettings': {
        \ '.*': {
            \ 'cmdline': 'neovim',
            \ 'content': 'text',
            \ 'priority': 0,
            \ 'selector': 'textarea',
            \ 'takeover': 'never',
        \ },
    \ }
\ }
`
	utils.AppendToTextFile(filepath.Join(utils.UserHomeDir(), ".SpaceVim/init.vim"), script, true)
}

// Enable/disable true color for SpaceVim.
func configureSpaceVimTrueColor(trueColor bool) {
	path := filepath.Join(utils.UserHomeDir(), ".SpaceVim.d/init.toml")
	lines := strings.Split(utils.ReadFileAsString(path), "\n")
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
	utils.WriteTextFile(path, text, 0o600)
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
	SpaceVim(
		utils.GetBoolFlag(cmd, "install"),
		utils.GetCommandPrefix(true, map[string]uint32{}),
		utils.BuildYesFlag(cmd),
		utils.GetBoolFlag(cmd, "config"),
		utils.GetBoolFlag(cmd, "strip"),
		utils.GetBoolFlag(cmd, "enable-true-color"),
		utils.GetBoolFlag(cmd, "disable-true-color"),
		utils.GetBoolFlag(cmd, "uninstall"),
		utils.GetStringFlag(cmd, "version"),
		utils.BuildPipInstall(cmd),
	)
}

func SpaceVim(install bool, prefix string, yes_s string, config bool, strip bool, enableTrueColor bool, disableTrueColor bool, uninstall bool, version string, pipInstall string) {
	if install {
		if utils.IsDebianUbuntuSeries() {
			cmd := utils.Format("{prefix} apt-get update && {prefix} apt-get install {yes_s} xfonts-utils", map[string]string{
				"prefix": prefix,
				"yes_s":  yes_s,
			})
			utils.RunCmd(cmd)
		}
		if version == "" {
			version = network.GetLatestRelease(network.GetReleaseUrl("SpaceVim/SpaceVim")).TagName
		} else if !strings.HasPrefix(version, "v") {
			version = "v" + version
		}
		utils.RemoveAll(filepath.Join(utils.UserHomeDir(), ".SpaceVim"))
		command := utils.Format(`curl -sLf https://spacevim.org/install.sh | bash \
			&& mkdir -p ~/.config && ln -svf ~/.SpaceVim ~/.config/nvim \
			&& cd ~/.SpaceVim && git checkout {version}`, map[string]string{
			"version": version,
		})
		utils.RunCmd(command)
		utils.RunCmd(utils.Format("{pip_install} python-lsp-server", map[string]string{
			"pip_install": pipInstall,
		}))
		log.Print("The Python package python-lsp-server is installed! Please make sure pylsp is on the search path!\n")
		// npm install -g bash-language-server javascript-typescript-langserver
		if strip {
			stripSpaceVim()
		}
		if utils.ExistsCommand("nvim") {
			utils.RunCmd(utils.Format(`nvim --headless +"call dein#install()" +qall && {pip_install} pynvim`, map[string]string{
				"pip_install": pipInstall,
			}))
		}
	}
	if config {
		home := utils.UserHomeDir()
		// configure .SpaceVim.d
		desDir := filepath.Join(home, ".SpaceVim.d")
		utils.MkdirAll(desDir, 0700)
		utils.CopyEmbedFileToDir("data/SpaceVim/SpaceVim.d/init.toml", desDir, 0600, true)
		utils.CopyEmbedFileToDir("data/SpaceVim/SpaceVim.d/vimrc", desDir, 0600, true)
		// -----------------------------------------------------------
		if enableTrueColor {
			configureSpaceVimTrueColor(true)
		}
		if disableTrueColor {
			configureSpaceVimTrueColor(false)
		}
		configureSpaceVimForFirenvim()
	}
	if uninstall {
		utils.RunCmd("curl -sLf https://spacevim.org/install.sh | bash -s -- --uninstall")
	}
}

var SpaceVimCmd = &cobra.Command{
	Use:     "spacevim",
	Aliases: []string{"svim"},
	Short:   "Install and configure SpaceVim.",
	//Args:  cobra.ExactArgs(1),
	Run: spaceVim,
}

func init() {
	SpaceVimCmd.Flags().BoolP("install", "i", false, "Install SpaceVim.")
	SpaceVimCmd.Flags().Bool("uninstall", false, "Uninstall SpaceVim.")
	SpaceVimCmd.Flags().BoolP("config", "c", false, "Configure SpaceVim.")
	SpaceVimCmd.Flags().StringP("version", "v", "", "The version (latest release by default) of SpaceVim to install.")
	SpaceVimCmd.Flags().Bool("enable-true-color", false, "Enable true color support in SpaceVim.")
	SpaceVimCmd.Flags().Bool("disable-true-color", false, "Disable true color support in SpaceVim.")
	SpaceVimCmd.Flags().Bool("strip", false, "Strip unnecessary files from '~/.SpaceVim/'.")
	SpaceVimCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	utils.AddPythonFlags(SpaceVimCmd)
	// rootCmd.AddCommand(spaceVimCmd)
}
