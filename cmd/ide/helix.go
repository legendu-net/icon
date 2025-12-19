package ide

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install and configure helix.
func helix(cmd *cobra.Command, _ []string) {
	Helix(
		utils.GetBoolFlag(cmd, "install"),
		utils.GetBoolFlag(cmd, "config"),
		utils.GetBoolFlag(cmd, "uninstall"),
		utils.BuildYesFlag(cmd),
	)
}

func Helix(install, config, uninstall bool, yesStr string) {
	if install {
		if utils.IsLinux() {
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
				command := utils.Format("{prefix} apt-get update && {prefix} apt-get install {yesStr} helix", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": yesStr,
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf copr enable varlad/helix && {prefix} dnf {yesStr} install helix", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": yesStr,
				})
				utils.RunCmd(command)
			}
		} else {
			utils.BrewInstallSafe([]string{"helix"})
		}
	}
	if config {
	}
	if uninstall {
		if utils.IsLinux() {
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format("{prefix} apt-get purge {yesStr} helix", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": yesStr,
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yesStr} remove helix", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
				})
				utils.RunCmd(command)
			}
		} else {
			utils.RunCmd("brew uninstall helix")
		}
	}
}

var helixCmd = &cobra.Command{
	Use:     "helix",
	Aliases: []string{"nvim"},
	Short:   "Install and configure helix.",
	//Args:  cobra.ExactArgs(1),
	Run: helix,
}

func ConfigHelixCmd(rootCmd *cobra.Command) {
	helixCmd.Flags().BoolP("install", "i", false, "Install helix.")
	helixCmd.Flags().Bool("uninstall", false, "Uninstall helix.")
	helixCmd.Flags().BoolP("config", "c", false, "Configure helix.")
	helixCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	helixCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	helixCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	rootCmd.AddCommand(helixCmd)
}
