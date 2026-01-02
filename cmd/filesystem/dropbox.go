package filesystem

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install and configure Dropbox.
func dropbox(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		utils.RunCmd("flatpak install flathub com.dropbox.Client")
	}
	if utils.GetBoolFlag(cmd, "config") {
		if utils.IsAtomicLinux() {
			utils.WriteTextFile(
				"~/.local/share/flatpak/overrides",
				`[Context]
filesystems=/var/home/dclong

[Environment]
HOME=/var/home/dclong
`, 0o644) //nolint:mnd // readable
		}
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		utils.RunCmd("flatpak uninstall com.dropbox.Client")
	}
}

var dropboxCmd = &cobra.Command{
	Use:     "dropbox",
	Aliases: []string{},
	Short:   "Install and configure Dropbox.",
	Run:     dropbox,
}

func ConfigDropboxCmd(rootCmd *cobra.Command) {
	dropboxCmd.Flags().BoolP("install", "i", false, "Install Dropbox.")
	dropboxCmd.Flags().Bool("uninstall", false, "Uninstall Dropbox.")
	dropboxCmd.Flags().BoolP("config", "c", false, "Configure Dropbox.")
	dropboxCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	dropboxCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	dropboxCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	rootCmd.AddCommand(dropboxCmd)
}
