package ide

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
	"runtime"
)

// Install and configure helix.
func helix(cmd *cobra.Command, args []string) {
	Helix(
		utils.GetBoolFlag(cmd, "install"),
		utils.GetBoolFlag(cmd, "config"),
		utils.GetBoolFlag(cmd, "uninstall"),
		utils.BuildYesFlag(cmd),
	)
}

func Helix(install bool, config bool, uninstall bool, yes_s string) {
	if install {
		switch runtime.GOOS {
		case "darwin":
			utils.BrewInstallSafe([]string{"helix"})
		case "linux":
			if utils.IsDebianUbuntuSeries() {
				if utils.IsUbuntuSeries() {
					command := utils.Format(`{prefix} add-apt-repository ppa:maveonair/helix-editor`, map[string]string{
						"prefix": utils.GetCommandPrefix(
							true,
							map[string]uint32{},
						),
					})
					utils.RunCmd(command)
				}
				command := utils.Format("{prefix} apt-get update && {prefix} apt-get install {yes_s} helix", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": yes_s,
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf copr enable varlad/helix && {prefix} dnf {yes_s} install helix", map[string]string{
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
	}
	if uninstall {
		switch runtime.GOOS {
		case "darwin":
			utils.RunCmd("brew uninstall helix")
		case "linux":
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format("{prefix} apt-get purge {yes_s} helix", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": yes_s,
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yes_s} remove helix", map[string]string{
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

var HelixCmd = &cobra.Command{
	Use:     "helix",
	Aliases: []string{"nvim"},
	Short:   "Install and configure helix.",
	//Args:  cobra.ExactArgs(1),
	Run: helix,
}

func init() {
	HelixCmd.Flags().BoolP("install", "i", false, "Install helix.")
	HelixCmd.Flags().Bool("uninstall", false, "Uninstall helix.")
	HelixCmd.Flags().BoolP("config", "c", false, "Configure helix.")
	HelixCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	// rootCmd.AddCommand(HelixCmd)
}
