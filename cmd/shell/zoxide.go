package shell

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install zoxide.
func zoxide(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		command := "curl -sS https://raw.githubusercontent.com/ajeetdsouza/zoxide/main/install.sh | bash"
		utils.RunCmd(command)
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
	// rootCmd.AddCommand(zoxideCmd)
}
