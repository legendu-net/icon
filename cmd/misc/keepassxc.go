package misc

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
	"log"
	"runtime"
)

// Install and configure the KeepassXC terminal.
func keepassxc(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "linux":
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format("{prefix} apt-get update && {prefix} apt-get install {yes_s} keepassxc", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yes_s":  utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yes_s} install keepassxc", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yes_s":  utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
		case "darwin":
			utils.RunCmd("brew install --cask keepassxc")
		default:
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		path := "~/.local/bin/"
		command := utils.Format(`mkdir -p {path} \
			&& ln -s /Applications/KeePassXC.app/Contents/MacOS/keepassxc-cli {path}`, map[string]string{
			"path": path,
		})
		utils.RunCmd(command)
		log.Printf("The command keepassxc-cli has been symlinked into %s.\n", path)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		switch runtime.GOOS {
		case "linux":
			command := utils.Format("{prefix} apt-get purge {yes_s} keepassxc", map[string]string{
				"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
				"yes_s":  utils.BuildYesFlag(cmd),
			})
			utils.RunCmd(command)
		case "darwin":
			utils.RunCmd("brew uninstall --cask keepassxc")
		default:
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
	KeepassXCCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	KeepassXCCmd.Flags().StringP("version", "v", "", "The version of the release.")
	// rootCmd.AddCommand(keepassxcCmd)
}
