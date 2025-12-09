package shell

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install bash-it, a community Bash framework.
// For more details, please refer to https://github.com/Bash-it/bash-it#installation.
func bashIt(cmd *cobra.Command, _ []string) {
	home := utils.UserHomeDir()
	if utils.GetBoolFlag(cmd, "install") {
		dir := filepath.Join(home, ".bash_it")
		utils.RemoveAll(dir)
		command := utils.Format(`git clone --depth=1 https://github.com/Bash-it/bash-it.git {dir} \
			&& {dir}/install.sh --silent -f`,
			map[string]string{
				"dir": dir,
			})
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
		utils.ConfigBash()
		utils.Symlink("~/.config/icon-data/bash-it", "~/.bash_it", true)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		utils.RunCmd("~/.bash_it/uninstall.sh && rm -rf ~/.bash_it/")
	}
}

var BashItCmd = &cobra.Command{
	Use:     "bash_it",
	Aliases: []string{"bashit", "bit"},
	Short:   "Install and configure bash-it.",
	//Args:  cobra.ExactArgs(1),
	Run: bashIt,
}

func init() {
	BashItCmd.Flags().BoolP("install", "i", false, "If specified, install bash-it.")
	BashItCmd.Flags().Bool("uninstall", false, "If specified, uninstall bash-it.")
	BashItCmd.Flags().BoolP("config", "c", false, "If specified, configure bash-it.")
	// rootCmd.AddCommand(bashItCmd)
}
