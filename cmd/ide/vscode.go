package ide

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
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format("{prefix} snap install --classic code", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yes_s} install vscode", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
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
		utils.MkdirAll(userDir, 0o700)
		utils.CopyEmbeddedFileToDir(srcFile, userDir, 0600, true)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		switch runtime.GOOS {
		case "darwin":
			command := "brew cask uninstall visual-studio-code"
			utils.RunCmd(command)
		case "linux":
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format("{prefix} apt-get purge {yes_s} vscode", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yes_s} remove vscode", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
				})
				utils.RunCmd(command)
			}
		default:
		}
	}
}

var VscodeCmd = &cobra.Command{
	Use:     "visual_studio_code",
	Aliases: []string{"vscode", "code"},
	Short:   "Install and configure Visual Studio Code.",
	//Args:  cobra.ExactArgs(1),
	Run: vscode,
}

func init() {
	VscodeCmd.Flags().BoolP("install", "i", false, "Install Visual Studio Code.")
	VscodeCmd.Flags().Bool("uninstall", false, "Uninstall Visual Studio Code.")
	VscodeCmd.Flags().BoolP("config", "c", false, "Configure Visual Studio Code.")
	VscodeCmd.Flags().StringP("user-dir", "d", "", "The configuration directory for Visual Studio Code.")
	VscodeCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	// rootCmd.AddCommand(vscodeCmd)
}
