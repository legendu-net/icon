package shell

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install atuin.
func atuin(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		command := "bash <(curl https://raw.githubusercontent.com/ellie/atuin/main/install.sh)"
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
		utils.ConfigBash()
		if utils.IsLinux() {
			utils.ReplacePattern(utils.GetBashConfigFile(), `eval "$(atuin init bash)"`, "eval \"$(atuin init bash --disable-up-arrow)\"\n")
		} else {
			atuinBash := `
[[ -f ~/.bash-preexec.sh ]] && source ~/.bash-preexec.sh
eval "$(atuin init bash --disable-up-arrow)"
`
			utils.AppendToTextFile(
				filepath.Join(utils.UserHomeDir(), ".bash_profile"),
				atuinBash,
				true,
			)
		}
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
	}
}

var atuinCmd = &cobra.Command{
	Use:     "atuin",
	Aliases: []string{"atuin"},
	Short:   "Install and configure atuin.",
	//Args:  cobra.ExactArgs(1),
	Run: atuin,
}

func ConfigAtuinCmd(rootCmd *cobra.Command) {
	atuinCmd.Flags().BoolP("install", "i", false, "If specified, install atuin.")
	atuinCmd.Flags().Bool("uninstall", false, "If specified, uninstall atuin.")
	atuinCmd.Flags().BoolP("config", "c", false, "If specified, configure atuin.")
	atuinCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	atuinCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	atuinCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	rootCmd.AddCommand(atuinCmd)
}
