package filesystem

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install fd. fdfind (the name of fd in the Debian/Ubuntu series) is made
// available as fd by symlinking it into ~/.local/bin.
func fd(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		utils.InstallPackages(utils.GetBoolFlag(cmd, "yes"), utils.PkgFd)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		utils.UninstallPackages(utils.GetBoolFlag(cmd, "yes"), utils.PkgFd)
	}
}

var fdCmd = &cobra.Command{
	Use:     "fd",
	Aliases: []string{"fdfind"},
	Short:   "Install fd.",
	//Args:  cobra.ExactArgs(1),
	Run: fd,
}

func ConfigFdCmd(rootCmd *cobra.Command) {
	fdCmd.Flags().BoolP("install", "i", false, "Install fd.")
	fdCmd.Flags().BoolP("uninstall", "u", false, "Uninstall fd.")
	fdCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	rootCmd.AddCommand(fdCmd)
}
