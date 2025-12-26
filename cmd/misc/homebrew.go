package misc

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install and configure Homebrew.
func homebrew(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		url := "https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh"
		command := utils.Format(`NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL {url})"`, map[string]string{
			"url": url,
		})
		utils.RunCmd(command)
		if utils.IsLinux() {
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format(`{prefix} apt-get update \
						&& {prefix} apt-get install {yesStr} build-essential procps curl file git`, map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format(`{prefix} dnf {yesStr} group install development-tools \
						&& {prefix} dnf {yesStr} install procps-ng curl file`, map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		if utils.IsLinux() {
			cmd := utils.Format(`{prefix} grep -q -E 'Defaults\s+secure_path\s*=.+/home/linuxbrew/.linuxbrew/bin.*' {file} \
			|| {prefix} sed -i '/^Defaults\s\+secure_path\s*=/s/"$/:\/home\/linuxbrew\/.linuxbrew\/bin"/g' {file}`, map[string]string{
				"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
				"file":   "/etc/sudoers",
			})
			utils.RunCmd(cmd)
		}
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
	}
}

var homebrewCmd = &cobra.Command{
	Use:     "keepassxc",
	Aliases: []string{"brew"},
	Short:   "Install and configure the KeepassXC terminal.",
	Run:     homebrew,
}

func ConfigHomebrewCmd(rootCmd *cobra.Command) {
	homebrewCmd.Flags().BoolP("install", "i", false, "Install Homebrew.")
	homebrewCmd.Flags().Bool("uninstall", false, "Uninstall Homebrew.")
	homebrewCmd.Flags().BoolP("config", "c", false, "Configure Homebrew.")
	homebrewCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	homebrewCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	homebrewCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	rootCmd.AddCommand(homebrewCmd)
}
