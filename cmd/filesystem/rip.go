package filesystem

import (
	"runtime"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install and configure rip (rm-improved).
func rip(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "linux":
			utils.RunCmd("cargo install rm-improved")
		case "darwin":
			utils.BrewInstallSafe([]string{"rm-improved"})
		default:
		}
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		switch runtime.GOOS {
		case "linux":
			utils.RunCmd("rm ~/.cargo/bin/rip")
		case "darwin":
			utils.RunCmd("brew uninstall rm-improved")
		default:
		}
	}
}

var RipCmd = &cobra.Command{
	Use:     "rip",
	Aliases: []string{},
	Short:   "Install and configure rip (rm-improved).",
	//Args:  cobra.ExactArgs(1),
	Run: rip,
}

func init() {
	RipCmd.Flags().BoolP("install", "i", false, "Install rip (rm-improved).")
	RipCmd.Flags().Bool("uninstall", false, "Uninstall rip (rm-improved).")
	RipCmd.Flags().BoolP("config", "c", false, "Configure rip (rm-improved).")
	RipCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	// rootCmd.AddCommand(RipCmd)
}
