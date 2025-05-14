package shell

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install bash-it, a community Bash framework.
// For more details, please refer to https://github.com/Bash-it/bash-it#installation.
func bashIt(cmd *cobra.Command, args []string) {
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
		configBashIt()
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		utils.RunCmd("~/.bash_it/uninstall.sh && rm -rf ~/.bash_it/")
	}
}

func copyBashitConfigFiles(files []string, dir string) {
	for _, file := range files {
		utils.CopyEmbedFileToDir(file, dir, 0600, true)
	}
}

func configBashIt() {
	copyBashitConfigFiles([]string{
		"data/bash-it/custom.bash",
	}, utils.NormalizePath("~/.bash_it/lib"))
	copyBashitConfigFiles([]string{
		"data/bash-it/custom.plugins.bash",
	}, utils.NormalizePath("~/.bash_it/plugins"))
	copyBashitConfigFiles(
		[]string{"data/bash-it/custom.aliases.bash"},
		utils.NormalizePath("~/.bash_it/aliases"))
	copyBashitConfigFiles(
		[]string{
			"data/bash-it/icon.completion.bash",
			"data/bash-it/ldc.completion.bash",
			"data/bash-it/custom.completion.bash",
		}, utils.NormalizePath("~/.bash_it/completion"))
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
