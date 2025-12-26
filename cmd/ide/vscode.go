package ide

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/utils"
)

// Install and configure Visual Studio Code.
func vscode(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		if utils.IsLinux() {
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format("{prefix} snap install --classic code", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yesStr} install vscode", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
		} else {
			command := "brew cask install visual-studio-code"
			utils.RunCmd(command)
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		icon.FetchConfigData(false, "")

		userDir := utils.GetStringFlag(cmd, "user-dir")
		if userDir == "" {
			if utils.IsLinux() {
				userDir = "~/.config/Code/User"
			} else {
				userDir = "~/Library/Application Support/Code/User"
			}
		}
		utils.SymlinkIntoDir("~/.config/icon-data/vscode/settings.json", userDir,
			!utils.GetBoolFlag(cmd, "no-backup"), utils.GetBoolFlag(cmd, "copy"))
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		if utils.IsLinux() {
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format("{prefix} apt-get {yesStr} purge vscode", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yesStr} remove vscode", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
		} else {
			command := "brew cask uninstall visual-studio-code"
			utils.RunCmd(command)
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

func ConfigVscodeCmd(rootCmd *cobra.Command) {
	vscodeCmd.Flags().BoolP("install", "i", false, "Install Visual Studio Code.")
	vscodeCmd.Flags().Bool("uninstall", false, "Uninstall Visual Studio Code.")
	vscodeCmd.Flags().BoolP("config", "c", false, "Configure Visual Studio Code.")
	vscodeCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	vscodeCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	vscodeCmd.Flags().StringP("user-dir", "d", "", "The configuration directory for Visual Studio Code.")
	vscodeCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	rootCmd.AddCommand(vscodeCmd)
}
