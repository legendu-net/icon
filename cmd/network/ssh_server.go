package network

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install and configure SSH server.
func SshServer(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		if utils.IsDebianUbuntuSeries() {
			command := utils.Format("{prefix} apt-get update && {prefix} apt-get install {yes_s} openssh-server fail2ban", map[string]string{
				"prefix": utils.GetCommandPrefix(
					true,
					map[string]uint32{},
				),
				"yes_s": utils.BuildYesFlag(cmd),
			})
			utils.RunCmd(command)
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		if utils.IsDebianUbuntuSeries() {
			command := utils.Format("{prefix} apt-get purge {yes_s} openssh-server fail2ban", map[string]string{
				"prefix": utils.GetCommandPrefix(
					true,
					map[string]uint32{},
				),
				"yes_s": utils.BuildYesFlag(cmd),
			})
			utils.RunCmd(command)
		}
	}
}

var SshServerCmd = &cobra.Command{
	Use:     "ssh_server",
	Aliases: []string{"sshs"},
	Short:   "Install and configure SSH server.",
	//Args:  cobra.ExactArgs(1),
	Run: SshServer,
}

func init() {
	SshServerCmd.Flags().BoolP("install", "i", false, "Install Git.")
	SshServerCmd.Flags().Bool("uninstall", false, "Uninstall Git.")
	SshServerCmd.Flags().BoolP("config", "c", false, "Configure Git.")
	SshServerCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	SshServerCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	SshServerCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
}
