package dev

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install and configure perf.
func perf(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		if utils.IsLinux() {
			if utils.IsDebianSeries() {
				command := utils.Format(`{prefix} apt-get {yesStr} update \
						&& {prefix} apt-get {yesStr} install linux-perf`, map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
				// TODO: leverage from_github to download git-delta and install it to /usr/local/bin!!!
			} else if utils.IsUbuntuSeries() {
				command := utils.Format(`{prefix} apt-get {yesStr} update \
						&& {prefix} apt-get {yesStr} install \
							linux-tools-common \
							linux-tools-generic \
							linux-tools-$(uname -r)`, map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yesStr} install perf", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
		} else {
			utils.BrewInstallSafe([]string{"gperftools"})
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		if utils.IsLinux() {
			command := utils.Format("{prefix} sysctl -w kernel.perf_event_paranoid=-1", map[string]string{
				"prefix": utils.GetCommandPrefix(
					true,
					map[string]uint32{},
				),
			})
			utils.RunCmd(command)
		} else {
		}
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		if utils.IsLinux() {
			if utils.IsDebianSeries() {
				command := utils.Format("{prefix} apt-get {yesStr} purge linux-perf", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsUbuntuSeries() {
				command := utils.Format(`{prefix} apt-get {yesStr} purge \
						linux-tools-common linux-tools-generic linux-tools-$(uname -r)`, map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yesStr} remove perf", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
		} else {
			utils.RunCmd("brew uninstall gperftools")
		}
	}
}

var perfCmd = &cobra.Command{
	Use:     "perf",
	Aliases: []string{},
	Short:   "Install and configure perf.",
	//Args:  cobra.ExactArgs(1),
	Run: perf,
}

func ConfigPerfCmd(rootCmd *cobra.Command) {
	perfCmd.Flags().BoolP("install", "i", false, "Install Git.")
	perfCmd.Flags().Bool("uninstall", false, "Uninstall Git.")
	perfCmd.Flags().BoolP("config", "c", false, "Configure Git.")
	perfCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	perfCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	perfCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	rootCmd.AddCommand(perfCmd)
}
