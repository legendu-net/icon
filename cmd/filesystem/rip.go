package filesystem

import (
	"runtime"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install and configure rip (rm-improved).
func rip(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "linux":
			utils.RunCmd("cargo install rip2")
		case "darwin":
			utils.BrewInstallSafe([]string{"rip2"})
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		switch runtime.GOOS {
		case "linux":
			utils.RunCmd("rm ~/.cargo/bin/rip")
		case "darwin":
			utils.RunCmd("brew uninstall rip2")
		}
	}
}

var ripCmd = &cobra.Command{
	Use:     "rip",
	Aliases: []string{},
	Short:   "Install and configure rip2 (rm-improved).",
	//Args:  cobra.ExactArgs(1),
	Run: rip,
}

func ConfigRipCmd(rootCmd *cobra.Command) {
	ripCmd.Flags().BoolP("install", "i", false, "Install rip2 (rm-improved).")
	ripCmd.Flags().Bool("uninstall", false, "Uninstall rip2 (rm-improved).")
	ripCmd.Flags().BoolP("config", "c", false, "Configure rip2 (rm-improved).")
	ripCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	ripCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	ripCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	rootCmd.AddCommand(ripCmd)
}
