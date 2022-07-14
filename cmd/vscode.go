package cmd

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
	"log"
	"path/filepath"
	"runtime"
)

// Install and configure Visual Studio Code.
func vscode(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "linux":
			if utils.IsDebianSeries() {
				command := utils.Format("{prefix} apt-get update && {prefix} apt-get install {yes_s} vscode", map[string]string{
					"prefix": utils.GetCommandPrefix(
						utils.GetBoolFlag(cmd, "sudo"),
						map[string]uint32{},
						"ls",
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} yum {yes_s} install vscode", map[string]string{
					"prefix": utils.GetCommandPrefix(
						utils.GetBoolFlag(cmd, "sudo"),
						map[string]uint32{},
						"ls",
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
		case "darwin":
			command := "brew cask install visual-studio-code"
			utils.RunCmd(command)
		default:
			log.Fatal("ERROR - The OS ", runtime.GOOS, " is not supported!")
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		srcFile := "data/vscode/settings.json"
		userDir := utils.GetStringFlag(cmd, "user-dir")
		if userDir == "" {
			home := utils.UserHomeDir()
			switch runtime.GOOS {
			case "darwin":
				userDir = filepath.Join(home, "Library/Application Support/Code/User")
			default:
				userDir = filepath.Join(home, ".config/Code/User")
			}
		}
		utils.MkdirAll(userDir, 0700)
		utils.CopyEmbedFileToDir(srcFile, userDir)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		switch runtime.GOOS {
		case "darwin":
			command := "brew cask uninstall visual-studio-code"
			utils.RunCmd(command)
		case "linux":
			if utils.IsDebianSeries() {
				command := utils.Format("{prefix} apt-get purge {yes_s} vscode", map[string]string{
					"prefix": utils.GetCommandPrefix(
						utils.GetBoolFlag(cmd, "sudo"),
						map[string]uint32{},
						"ls",
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} yum remove vscode", map[string]string{
					"prefix": utils.GetCommandPrefix(
						utils.GetBoolFlag(cmd, "sudo"),
						map[string]uint32{},
						"ls",
					),
				})
				utils.RunCmd(command)
			}
		default:
		}
	}
}

var vscodeCmd = &cobra.Command{
	Use:     "visual_studio_code",
	Aliases: []string{"vscode", "code"},
	Short:   "Install and configure Visual Studio Code.",
	//Args:  cobra.ExactArgs(1),
	Run: vscode,
}

func init() {
	vscodeCmd.Flags().BoolP("install", "i", false, "Install Visual Studio Code.")
	vscodeCmd.Flags().Bool("uninstall", false, "Uninstall Visual Studio Code.")
	vscodeCmd.Flags().BoolP("config", "c", false, "Configure Visual Studio Code.")
	vscodeCmd.Flags().StringP("user-dir", "d", "", "The configuration directory for Visual Studio Code.")
	rootCmd.AddCommand(vscodeCmd)
}
