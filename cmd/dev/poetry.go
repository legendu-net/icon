package dev

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
	"log"
	"path/filepath"
	"runtime"
)

// Install and configure Python Poetry.
func poetry(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		url := "https://install.python-poetry.org"
		command := utils.Format("curl -sSL {url} | {python}", map[string]string{
			"url":    url,
			"python": utils.GetStringFlag(cmd, "python"),
		})
		version := utils.GetStringFlag(cmd, "version")
		if version != "" {
			command += " - --version " + version
		}
		utils.RunCmd(command)
	}
	poetry_bin := filepath.Join(utils.UserHomeDir(), ".local/bin/poetry")
	if utils.GetBoolFlag(cmd, "config") {
		// make poetry always create virtual environment in the root directory of the project
		command := utils.Format("{poetry_bin} config virtualenvs.in-project true", map[string]string{
			"poetry_bin": poetry_bin,
		})
		utils.RunCmd(command)
		log.Printf("Python poetry has been configured to create virtual environments inside projects!\n")
		// bash completion
		if utils.GetBoolFlag(cmd, "bash-completion") {
			switch runtime.GOOS {
			case "linux":
				command := utils.Format(`{poetry_bin} completions bash | tee \
                    /etc/bash_completion.d/poetry.bash-completion > /dev/null`, map[string]string{
					"poetry_bin": poetry_bin,
				})
				utils.RunCmd(command)
			case "darwin":
				command := utils.Format(`{poetry_bin} completions bash > \
                    $(brew --prefix)/etc/bash_completion.d/poetry.bash-completion`, map[string]string{
					"poetry_bin": poetry_bin,
				})
				utils.RunCmd(command)
			default:
			}
			log.Printf("Bash completion is enabled for poetry.\n")
		}
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		command := utils.Format("{poetry_bin} self:uninstall", map[string]string{
			"poetry_bin": poetry_bin,
		})
		utils.RunCmd(command)
	}
}

var PoetryCmd = &cobra.Command{
	Use:     "poetry",
	Aliases: []string{"pt"},
	Short:   "Install and configure Python Poetry.",
	//Args:  cobra.ExactArgs(1),
	Run: poetry,
}

func init() {
	PoetryCmd.Flags().BoolP("install", "i", false, "Install Python Poetry.")
	PoetryCmd.Flags().Bool("uninstall", false, "Uninstall Python Poetry.")
	PoetryCmd.Flags().BoolP("config", "c", false, "Configure Python Poetry.")
	PoetryCmd.Flags().BoolP("bash-completion", "b", false, "Configure Bash completion for Python Poetry.")
	PoetryCmd.Flags().StringP("version", "v", "", "The version of Python Poetry to install.")
	PoetryCmd.Flags().String("python", "python3", "Path to the python3 command.")
	// rootCmd.AddCommand(PoetryCmd)
}
