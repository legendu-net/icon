package ide

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
	"runtime"
)

// Install and configure neovim.
func neovim(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "darwin":
			utils.BrewInstallSafe([]string{"neovim"})
		case "linux":
			if utils.IsDebianSeries() {
				command := utils.Format("{prefix} apt-get update && {prefix} apt-get install {yes_s} neovim", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} yum {yes_s} install neovim", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
		default:
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		switch runtime.GOOS {
		case "darwin":
			utils.RunCmd("brew uninstall neovim")
		case "linux":
			if utils.IsDebianSeries() {
				command := utils.Format("{prefix} apt-get purge {yes_s} neovim", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} yum remove neovim", map[string]string{
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
	// rootCmd.AddCommand(spaceVimCmd)
}
