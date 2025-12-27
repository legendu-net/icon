package shell

import (
	"log"
	"runtime"

	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
)

func downloadHyperFromGitHub(version string) string {
	output := "/tmp/_hyper_js_terminal"
	network.DownloadGitHubRelease("vercel/hyper", version, map[string][]string{
		"common":             {},
		"amd64":              {"amd64"},
		"arm64":              {"arm64"},
		"DebianUbuntuSeries": {"deb"},
		"FedoraSeries":       {"rpm"},
		"OtherLinux":         {"appimage"},
	}, []string{}, output)
	return output
}

// Install and configure the Hyper terminal.
func hyper(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "linux":
			file := downloadHyperFromGitHub(utils.GetStringFlag(cmd, "version"))
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format(`{prefix} apt-get {yesStr} update \
						&& {prefix} apt-get {yesStr} install {file}`, map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yesStr": utils.BuildYesFlag(cmd),
					"file":   file,
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yesStr} install {file}", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yesStr": utils.BuildYesFlag(cmd),
					"file":   file,
				})
				utils.RunCmd(command)
			}
		case "darwin":
			utils.RunCmd("brew install --cask hyper")
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		icon.FetchConfigData(false, "")

		utils.RunCmd("hyper i hypercwd")
		utils.RunCmd("hyper i hyper-search")
		utils.RunCmd("hyper i hyper-pane")
		utils.RunCmd("hyper i hyperpower")
		log.Printf("Hyper plugins hypercwd, hyper-search, hyper-pane and hyperpower are installed.\n")
		utils.Symlink(
			"~/.config/icon-data/hyper/hyper.js",
			"~/.hyper.js",
			!utils.GetBoolFlag(cmd, "no-backup"), utils.GetBoolFlag(cmd, "copy"))
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		switch runtime.GOOS {
		case "linux":
		case "darwin":
			utils.RunCmd("brew uninstall --cask hyper")
		}
	}
}

var hyperCmd = &cobra.Command{
	Use:     "hyper",
	Aliases: []string{},
	Short:   "Install and configure the Hyper terminal.",
	//Args:  cobra.ExactArgs(1),
	Run: hyper,
}

func ConfigHyperCmd(rootCmd *cobra.Command) {
	hyperCmd.Flags().BoolP("install", "i", false, "Install the Hyper terminal.")
	hyperCmd.Flags().Bool("uninstall", false, "Uninstall Hyper terminal.")
	hyperCmd.Flags().BoolP("config", "c", false, "Configure the Hyper terminal.")
	hyperCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	hyperCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	hyperCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	hyperCmd.Flags().StringP("version", "v", "", "The version of the release.")
	rootCmd.AddCommand(hyperCmd)
}
