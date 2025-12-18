package ide

import (
	"runtime"

	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/utils"
)

// Install and configure neovim.
func neovim(cmd *cobra.Command, _ []string) {
	Neovim(
		utils.GetBoolFlag(cmd, "install"),
		utils.GetBoolFlag(cmd, "config"),
		utils.GetBoolFlag(cmd, "uninstall"),
		utils.BuildYesFlag(cmd),
		!utils.GetBoolFlag(cmd, "no-backup"), utils.GetBoolFlag(cmd, "copy"))
}

func Neovim(install, config, uninstall bool, yesStr string, backup, copyPath bool) {
	if install {
		switch runtime.GOOS {
		case "darwin":
			utils.BrewInstallSafe([]string{"neovim"})
		case "linux":
			if utils.IsDebianUbuntuSeries() {
				if utils.IsUbuntuSeries() {
					command := utils.Format(`{prefix} apt-get update && {prefix} apt-get install {yesStr} gnupg \
						&& {prefix} add-apt-repository {yesStr} ppa:neovim-ppa/unstable`, map[string]string{
						"prefix": utils.GetCommandPrefix(
							true,
							map[string]uint32{},
						),
						"yesStr": yesStr,
					})
					utils.RunCmd(command)
				}
				command := utils.Format("{prefix} apt-get update && {prefix} apt-get install {yesStr} neovim", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": yesStr,
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yesStr} install neovim", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": yesStr,
				})
				utils.RunCmd(command)
			}
		}
	}
	if config {
		icon.FetchConfigData(false, "")
		dir := "~/.config/nvim"
		utils.Symlink("~/.config/icon-data/nvim", dir, backup, copyPath)
	}
	if uninstall {
		switch runtime.GOOS {
		case "darwin":
			utils.RunCmd("brew uninstall neovim")
		case "linux":
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format("{prefix} apt-get purge {yesStr} neovim", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": yesStr,
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yesStr} remove neovim", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
				})
				utils.RunCmd(command)
			}
		}
	}
}

var NeovimCmd = &cobra.Command{
	Use:     "neovim",
	Aliases: []string{"nvim"},
	Short:   "Install and configure neovim.",
	//Args:  cobra.ExactArgs(1),
	Run: neovim,
}

func init() {
	NeovimCmd.Flags().BoolP("install", "i", false, "Install neovim.")
	NeovimCmd.Flags().Bool("uninstall", false, "Uninstall neovim.")
	NeovimCmd.Flags().BoolP("config", "c", false, "Configure neovim.")
	NeovimCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	NeovimCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	NeovimCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
}
