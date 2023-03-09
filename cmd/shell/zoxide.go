package shell

import (
	"runtime"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install zoxide.
func zoxide(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		command := "curl -sS https://raw.githubusercontent.com/ajeetdsouza/zoxide/main/install.sh | bash"
		utils.RunCmd(command)
		switch runtime.GOOS {
		case "darwin":
			utils.BrewInstallSafe([]string{"fzf"})
		case "linux":
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format("{prefix} apt-get update && {prefix} apt-get install {yes_s} fzf", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yes_s} install fzf", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
		default:
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		utils.ConfigBash()
		utils.AppendToTextFile(utils.GetBashConfigFile(), "eval \"$(zoxide init bash)\"\n", true)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
	}
}

var ZoxideCmd = &cobra.Command{
	Use:     "zoxide",
	Aliases: []string{"zoxide", "z"},
	Short:   "Install and configure zoxide.",
	//Args:  cobra.ExactArgs(1),
	Run: zoxide,
}

func init() {
	ZoxideCmd.Flags().BoolP("install", "i", false, "If specified, install zoxide.")
	ZoxideCmd.Flags().Bool("uninstall", false, "If specified, uninstall zoxide.")
	ZoxideCmd.Flags().BoolP("config", "c", false, "If specified, configure zoxide.")
	ZoxideCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	// rootCmd.AddCommand(zoxideCmd)
}
