package cmd

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
	"log"
	"path/filepath"
	"runtime"
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
		profile := ".bash_profile"
		if runtime.GOOS == "linux" {
			profile = ".bashrc"
		}
		profile = filepath.Join(home, profile)
		utils.AddPathShell([]string{"$HOME/*/bin"}, profile)
		log.Printf("%s is configured to insert $HOME/*/bin into $PATH.", profile)
		if runtime.GOOS == "linux" {
			text := `
# source in ~/.bashrc
if [[ -f $HOME/.bashrc ]]; then
	. $HOME/.bashrc
fi
`
			utils.AppendToTextFile(filepath.Join(home, ".bash_profile"), text)
		}
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		utils.RunCmd("~/.bash_it/uninstall.sh && rm -rf ~/.bash_it/")
	}
}

var bashItCmd = &cobra.Command{
	Use:     "bash_it",
	Aliases: []string{"bashit", "bit"},
	Short:   "Install and configure bash-it.",
	//Args:  cobra.ExactArgs(1),
	Run: bashIt,
}

func init() {
	bashItCmd.Flags().BoolP("install", "i", false, "If specified, install bash-it.")
	bashItCmd.Flags().BoolP("config", "c", false, "If specified, configure bash-it.")
	rootCmd.AddCommand(bashItCmd)
}
