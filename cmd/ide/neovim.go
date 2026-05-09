package ide

import (
	"runtime"

	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/utils"
)

// Install and configure Neovim.
func neovim(cmd *cobra.Command, _ []string) {
	Neovim(
		utils.GetBoolFlag(cmd, "install"),
		utils.GetBoolFlag(cmd, "config"),
		utils.GetBoolFlag(cmd, "uninstall"),
		utils.GetBoolFlag(cmd, "brew"),
		utils.BuildYesFlag(cmd),
		!utils.GetBoolFlag(cmd, "no-backup"), utils.GetBoolFlag(cmd, "copy"))
}

func Neovim(install, config, uninstall, brew bool, yesStr string, backup, copyPath bool) {
	if install {
		if runtime.GOOS == "darwin" || brew || utils.IsUniversalBlue() {
			utils.BrewInstallSafe([]string{"neovim"})
		} else if utils.IsDebianUbuntuSeries() {
			command := utils.Format(`{prefix} apt-get {yesStr} update \
					&& {prefix} apt-get {yesStr} install neovim`, map[string]string{
				"prefix": utils.GetCommandPrefix(
					true,
					map[string]uint32{},
				),
				"yesStr": yesStr,
			})
			utils.RunCmd(command)
		} else if utils.IsFedoraSeries() {
			command := utils.Format("{prefix} dnf {yesStr} install neovim", map[string]string{
				"prefix": utils.GetCommandPrefix(
					true,
					map[string]uint32{},
				),
				"yesStr": yesStr,
			})
			utils.RunCmd(command)
		}
	}
	if config {
		icon.FetchConfigData(false, "")
		dir := "~/.config/nvim"
		utils.Symlink("~/.config/icon-data/nvim", dir, backup, copyPath)
	}
	if uninstall {
		if runtime.GOOS == "darwin" || brew || utils.IsUniversalBlue() {
			utils.RunCmd("brew uninstall neovim")
		} else if utils.IsDebianUbuntuSeries() {
			command := utils.Format("{prefix} apt-get {yesStr} purge neovim", map[string]string{
				"prefix": utils.GetCommandPrefix(
					true,
					map[string]uint32{},
				),
				"yesStr": yesStr,
			})
			utils.RunCmd(command)
		} else if utils.IsFedoraSeries() {
			command := utils.Format("{prefix} dnf {yesStr} remove neovim", map[string]string{
				"prefix": utils.GetCommandPrefix(
					true,
					map[string]uint32{},
				),
				"yesStr": yesStr,
			})
			utils.RunCmd(command)
		}
	}
}

var neovimCmd = &cobra.Command{
	Use:     "neovim",
	Aliases: []string{"nvim"},
	Short:   "Install and configure Neovim.",
	//Args:  cobra.ExactArgs(1),
	Run: neovim,
}

func ConfigNeovimCmd(rootCmd *cobra.Command) {
	neovimCmd.Flags().BoolP("install", "i", false, "Install Neovim.")
	neovimCmd.Flags().Bool("uninstall", false, "Uninstall Neovim.")
	neovimCmd.Flags().BoolP("config", "c", false, "Configure Neovim.")
	neovimCmd.Flags().Bool("brew", false, "Install Neovim using Homebrew.")
	neovimCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	neovimCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	neovimCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	rootCmd.AddCommand(neovimCmd)
}
