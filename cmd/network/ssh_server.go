package network

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install and configure SSH server.
func SSHServer(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		if utils.IsDebianUbuntuSeries() {
			command := utils.Format(`{prefix} apt-get {yesStr} update \
					&& {prefix} apt-get {yesStr} install openssh-server fail2ban`, map[string]string{
				"prefix": utils.GetCommandPrefix(
					true,
					map[string]uint32{},
				),
				"yesStr": utils.BuildYesFlag(cmd),
			})
			utils.RunCmd(command)
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		if utils.IsDebianUbuntuSeries() {
			command := utils.Format("{prefix} apt-get {yesStr} purge openssh-server fail2ban", map[string]string{
				"prefix": utils.GetCommandPrefix(
					true,
					map[string]uint32{},
				),
				"yesStr": utils.BuildYesFlag(cmd),
			})
			utils.RunCmd(command)
		}
	}
}

var sshServerCmd = &cobra.Command{
	Use:     "ssh_server",
	Aliases: []string{"sshs"},
	Short:   "Install and configure SSH server.",
	//Args:  cobra.ExactArgs(1),
	Run: SSHServer,
}

func ConfigSSHServerCmd(rootCmd *cobra.Command) {
	sshServerCmd.Flags().BoolP("install", "i", false, "Install Git.")
	sshServerCmd.Flags().Bool("uninstall", false, "Uninstall Git.")
	sshServerCmd.Flags().BoolP("config", "c", false, "Configure Git.")
	sshServerCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	sshServerCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	sshServerCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	rootCmd.AddCommand(sshServerCmd)
}
