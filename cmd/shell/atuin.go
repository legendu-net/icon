package shell

import (
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install atuin.
func atuin(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		command := "bash <(curl https://raw.githubusercontent.com/ellie/atuin/main/install.sh)"
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
		utils.ConfigBash()
		switch runtime.GOOS {
		case "darwin":
			atuinBash := `
[[ -f ~/.bash-preexec.sh ]] && source ~/.bash-preexec.sh
eval "$(atuin init bash)"
`
			utils.AppendToTextFile(
				filepath.Join(utils.UserHomeDir(), ".bash_profile"),
				atuinBash,
				true,
			)
		case "linux":
		default:
		}
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
	}
}

var AtuinCmd = &cobra.Command{
	Use:     "atuin",
	Aliases: []string{"atuin"},
	Short:   "Install and configure atuin.",
	//Args:  cobra.ExactArgs(1),
	Run: atuin,
}

func init() {
	AtuinCmd.Flags().BoolP("install", "i", false, "If specified, install atuin.")
	AtuinCmd.Flags().Bool("uninstall", false, "If specified, uninstall atuin.")
	AtuinCmd.Flags().BoolP("config", "c", false, "If specified, configure atuin.")
	AtuinCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	// rootCmd.AddCommand(AtuinCmd)
}
