package shell

import (
	"runtime"

	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
)

func downloadFishFromGitHub(version string) string {
	output := "/tmp/_fish_shell.tar.xz"
	network.DownloadGitHubRelease("fish-shell/fish-shell", version, map[string][]string{
		"common":             {"fish", "tar.xz"},
		"x86_64":             {"x86_64"},
		"arm64":              {"aarch64"},
		"DebianUbuntuSeries": {},
		"FedoraSeries":       {},
		"OtherLinux":         {},
	}, []string{"app", "zip", "pkg", "asc"}, output)
	return output
}

// Install and config the fish shell.
func fish(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "linux":
			file := downloadFishFromGitHub(utils.GetStringFlag(cmd, "version"))
			command := utils.Format(`{prefix} tar --xz -xvf {file} -C /usr/bin/`, map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"file":   file,
				})
			utils.RunCmd(command)
		case "darwin":
			utils.RunCmd("brew install fish")
		default:
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		utils.CopyEmbeddedDir("data/fish", utils.NormalizePath("~/.config/fish"), true)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		switch runtime.GOOS {
		case "linux":
			command := utils.Format(`{prefix} rm /usr/bin/fish`, map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
				})
			utils.RunCmd(command)
		case "darwin":
			utils.RunCmd("brew uninstall fish")
		default:
		}
	}
}

var FishCmd = &cobra.Command{
	Use:     "fish",
	Aliases: []string{},
	Short:   "Install and configure the fish shell.",
	Run: fish,
}

func init() {
	FishCmd.Flags().BoolP("install", "i", false, "If specified, install the fish shell.")
	FishCmd.Flags().Bool("uninstall", false, "If specified, uninstall the fish shell.")
	FishCmd.Flags().BoolP("config", "c", false, "If specified, configure the fish shell.")
	FishCmd.Flags().StringP("version", "v", "", "The version of the release.")
}
