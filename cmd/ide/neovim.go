package ide

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
	"runtime"
)

// Install and configure neovim.
func neovim(cmd *cobra.Command, args []string) {
	Neovim(
		utils.GetBoolFlag(cmd, "install"),
		utils.GetBoolFlag(cmd, "config"),
		utils.GetBoolFlag(cmd, "uninstall"),
		utils.BuildYesFlag(cmd),
	)
}

func Neovim(install bool, config bool, uninstall bool, yes_s string) {
	if install {
		switch runtime.GOOS {
		case "darwin":
			utils.BrewInstallSafe([]string{"neovim"})
		case "linux":
			if utils.IsDebianUbuntuSeries() {
				if utils.IsUbuntuSeries() {
					command := utils.Format(`{prefix} apt-get update && {prefix} apt-get install {yes_s} gnupg \
						&& {prefix} add-apt-repository {yes_s} ppa:neovim-ppa/stable`, map[string]string{
						"prefix": utils.GetCommandPrefix(
							true,
							map[string]uint32{},
						),
						"yes_s": yes_s,
					})
					utils.RunCmd(command)
				}
				command := utils.Format("{prefix} apt-get update && {prefix} apt-get install {yes_s} neovim", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": yes_s,
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yes_s} install neovim", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": yes_s,
				})
				utils.RunCmd(command)
			}
		default:
		}
	}
	if config {
		utils.MkdirAll("~/.config/nvim", 0o700)
		utils.RunCmd("git clone https://github.com/legendu-net/nvim ~/.config/nvim")
	}
	if uninstall {
		switch runtime.GOOS {
		case "darwin":
			utils.RunCmd("brew uninstall neovim")
		case "linux":
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format("{prefix} apt-get purge {yes_s} neovim", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": yes_s,
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yes_s} remove neovim", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
				})
				utils.RunCmd(command)
			}
		default:
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
	NeovimCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	// rootCmd.AddCommand(NeovimCmd)
}
