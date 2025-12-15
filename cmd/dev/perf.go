package dev

import (
	"runtime"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install and configure perf.
func perf(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "linux":
			if utils.IsDebianSeries() {
				command := utils.Format("{prefix} apt-get update && {prefix} apt-get install {yes_s} linux-perf", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
				// TODO: leverage from_github to download git-delta and install it to /usr/local/bin!!!
			} else if utils.IsUbuntuSeries() {
				command := utils.Format(`{prefix} apt-get update \
					&& {prefix} apt-get install {yes_s} linux-tools-common linux-tools-generic linux-tools-$(uname -r)`, map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yes_s} install perf", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
		case "darwin":
			utils.BrewInstallSafe([]string{"gperftools"})
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		switch runtime.GOOS {
		case "linux":
			command := utils.Format("{prefix} sysctl -w kernel.perf_event_paranoid=-1", map[string]string{
				"prefix": utils.GetCommandPrefix(
					true,
					map[string]uint32{},
				),
			})
			utils.RunCmd(command)
		case "darwin":
		}
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		switch runtime.GOOS {
		case "linux":
			if utils.IsDebianSeries() {
				command := utils.Format("{prefix} apt-get purge {yes_s} linux-perf", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsUbuntuSeries() {
				command := utils.Format("{prefix} apt-get purge {yes_s} linux-tools-common linux-tools-generic linux-tools-$(uname -r)", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yes_s} remove perf", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
		case "darwin":
			utils.RunCmd("brew uninstall gperftools")
		}
	}
}

var PerfCmd = &cobra.Command{
	Use:     "perf",
	Aliases: []string{},
	Short:   "Install and configure perf.",
	//Args:  cobra.ExactArgs(1),
	Run: perf,
}

func init() {
	PerfCmd.Flags().BoolP("install", "i", false, "Install Git.")
	PerfCmd.Flags().Bool("uninstall", false, "Uninstall Git.")
	PerfCmd.Flags().BoolP("config", "c", false, "Configure Git.")
	PerfCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	PerfCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	PerfCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
}
