package shell

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install and configure the Alacritty terminal.
func alacritty(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		if utils.IsLinux() {
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format(`{prefix} apt-get update && {prefix} apt-get install {yesStr} \
					cmake pkg-config python3 \
					libfreetype6-dev libfontconfig1-dev libxcb-xfixes0-dev libxkbcommon-dev
					`, map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format(`{prefix} dnf {yesStr} install \ 
					cmake g++ \
					freetype-devel fontconfig-devel libxcb-devel libxkbcommon-devel 
					`, map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
			utils.RunCmd("cargo install alacritty")
			command := utils.Format(`{prefix} curl -sSL -o /usr/share/pixmaps/Alacritty.svg \
					https://raw.githubusercontent.com/alacritty/alacritty/master/extra/logo/alacritty-term.svg \
				&& curl -sSL -o /tmp/Alacritty.desktop \
					https://raw.githubusercontent.com/alacritty/alacritty/master/extra/linux/Alacritty.desktop \
				&& {prefix} mv ~/.cargo/bin/alacritty /usr/local/bin/ \
				&& {prefix} desktop-file-install /tmp/Alacritty.desktop \
				&& {prefix} update-desktop-database
				`, map[string]string{
				"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
			})
			utils.RunCmd(command)
		} else {
			utils.RunCmd("brew install --cask alacritty")
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		if utils.IsLinux() {
		} else {
			utils.RunCmd("brew uninstall --cask alacritty")
		}
	}
}

var alacrittyCmd = &cobra.Command{
	Use:     "alacritty",
	Aliases: []string{"alac"},
	Short:   "Install and configure the Alacritty terminal.",
	//Args:  cobra.ExactArgs(1),
	Run: alacritty,
}

func ConfigAlacrittyCmd(rootCmd *cobra.Command) {
	alacrittyCmd.Flags().BoolP("install", "i", false, "Install the Alacritty terminal.")
	alacrittyCmd.Flags().Bool("uninstall", false, "Uninstall Alacritty terminal.")
	alacrittyCmd.Flags().BoolP("config", "c", false, "Configure the Alacritty terminal.")
	alacrittyCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	alacrittyCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	alacrittyCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	rootCmd.AddCommand(alacrittyCmd)
}
