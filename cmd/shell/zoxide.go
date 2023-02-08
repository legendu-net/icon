package shell

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install bash-it, a community Bash framework.
// For more details, please refer to https://github.com/Bash-it/bash-it#installation.
func zoxide(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		command := "curl -sS https://raw.githubusercontent.com/ajeetdsouza/zoxide/main/install.sh | bash"
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
		utils.ConfigBash()
		utils.AppendToTextFile(utils.GetBashConfigFile(), `eval "$(zoxide init bash)"`, true)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		utils.RunCmd("~/.bash_it/uninstall.sh && rm -rf ~/.bash_it/")
	}
}

var ZoxideCmd = &cobra.Command{
	Use:     "bash_it",
	Aliases: []string{"zoxide", "bit"},
	Short:   "Install and configure bash-it.",
	//Args:  cobra.ExactArgs(1),
	Run: zoxide,
}

func init() {
	ZoxideCmd.Flags().BoolP("install", "i", false, "If specified, install bash-it.")
	ZoxideCmd.Flags().Bool("uninstall", false, "If specified, uninstall bash-it.")
	ZoxideCmd.Flags().BoolP("config", "c", false, "If specified, configure bash-it.")
	// rootCmd.AddCommand(zoxideCmd)
}
