package shell

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
	"log"
	"path/filepath"
	"runtime"
)

func downloadHyperFromGitHub(version string) string {
	output := "/tmp/_hyper_js_terminal"
	network.DownloadGitHubRelease("vercel/hyper", version, map[string][]string{
		"common":             {},
		"x86_64":             {"amd64"},
		"arm64":              {"arm64"},
		"DebianUbuntuSeries": {"deb"},
		"FedoraSeries":       {"rpm"},
		"OtherLinux":         {"appimage"},
	}, []string{}, output)
	return output
}

// Install and configure the Hyper terminal.
func hyper(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "linux":
			file := downloadHyperFromGitHub(utils.GetStringFlag(cmd, "version"))
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format("{prefix} apt-get update && {prefix} apt-get install {yes_s} {file}", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yes_s":  utils.BuildYesFlag(cmd),
					"file":   file,
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yes_s} install {file}", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yes_s":  utils.BuildYesFlag(cmd),
					"file":   file,
				})
				utils.RunCmd(command)
			}
		case "darwin":
			utils.RunCmd("brew install --cask hyper")
		default:
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		utils.RunCmd("hyper i hypercwd")
		utils.RunCmd("hyper i hyper-search")
		utils.RunCmd("hyper i hyper-pane")
		utils.RunCmd("hyper i hyperpower")
		log.Printf("Hyper plugins hypercwd, hyper-search, hyper-pane and hyperpower are installed.\n")
		utils.CopyEmbeddedFile("data/hyper/hyper.js", filepath.Join(utils.UserHomeDir(), ".hyper.js"), 0o600, true)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		switch runtime.GOOS {
		case "linux":
		case "darwin":
			utils.RunCmd("brew uninstall --cask hyper")
		default:
		}
	}
}

var HyperCmd = &cobra.Command{
	Use:     "hyper",
	Aliases: []string{},
	Short:   "Install and configure the Hyper terminal.",
	//Args:  cobra.ExactArgs(1),
	Run: hyper,
}

func init() {
	HyperCmd.Flags().BoolP("install", "i", false, "Install the Hyper terminal.")
	HyperCmd.Flags().Bool("uninstall", false, "Uninstall Hyper terminal.")
	HyperCmd.Flags().BoolP("config", "c", false, "Configure the Hyper terminal.")
	HyperCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	HyperCmd.Flags().StringP("version", "v", "", "The version of the release.")
	// rootCmd.AddCommand(HyperCmd)
}
