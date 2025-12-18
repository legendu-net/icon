package misc

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install and configure the KeepassXC terminal.
func keepassxc(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		if utils.IsLinux() {
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format("{prefix} apt-get update && {prefix} apt-get install {yesStr} keepassxc", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yesStr} install keepassxc", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
		} else {
			utils.RunCmd("brew install --cask keepassxc")
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		if utils.IsLinux() {
		} else {
			utils.SymlinkIntoDir("/Applications/KeePassXC.app/Contents/MacOS/keepassxc-cli", "~/.local/bin", false, false)
		}
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		if utils.IsLinux() {
			command := utils.Format("{prefix} apt-get purge {yesStr} keepassxc", map[string]string{
				"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
				"yesStr": utils.BuildYesFlag(cmd),
			})
			utils.RunCmd(command)
		} else {
			utils.RunCmd("brew uninstall --cask keepassxc")
		}
	}
}

var KeepassXCCmd = &cobra.Command{
	Use:     "keepassxc",
	Aliases: []string{},
	Short:   "Install and configure the KeepassXC terminal.",
	//Args:  cobra.ExactArgs(1),
	Run: keepassxc,
}

func init() {
	KeepassXCCmd.Flags().BoolP("install", "i", false, "Install the keepassxc terminal.")
	KeepassXCCmd.Flags().Bool("uninstall", false, "Uninstall keepassxc terminal.")
	KeepassXCCmd.Flags().BoolP("config", "c", false, "Configure the keepassxc terminal.")
	KeepassXCCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	KeepassXCCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	KeepassXCCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
}
