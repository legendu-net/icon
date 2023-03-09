package shell

import (
	"log"
	"runtime"

	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
)

func downloadNushellFromGitHub(version string) string {
	output := "/tmp/_nu.tar.gz"
	network.DownloadGitHubRelease("nushell/nushell", version, 
	map[string][]string{
		"common": {"tar.gz"},
		"linux": {"unknown", "linux", "gnu"},
		"darwin": {"apple", "darwin"},
		"x86_64": {"x86_64"},
		"arm64": {"aarch64"},
	}, []string{}, output)
	return output
}

// Install nushell.
func nushell(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "windows":
			utils.RunCmd("winget install nushell")
		case "darwin":
			utils.BrewInstallSafe([]string{"nushell"})
		case "linux":
			file := downloadNushellFromGitHub(utils.GetStringFlag(cmd, "version"))
			dir := utils.GetStringFlag(cmd, "dir")
			utils.Format(`mkdir -p {dir} \
					&& tar -zxvf {file} -C {dir} --strip=1 --exclude=LICENSE --exclude='README.*'`, map[string]string{
				"file": file,
				"dir":  dir,
			})
			log.Printf("Nushell has been installed into %s.\n", dir)
		default:
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
	}
}

var NushellCmd = &cobra.Command{
	Use:     "nushell",
	Aliases: []string{"nu"},
	Short:   "Install and configure nushell.",
	//Args:  cobra.ExactArgs(1),
	Run: nushell,
}

func init() {
	NushellCmd.Flags().BoolP("install", "i", false, "If specified, install nushell.")
	NushellCmd.Flags().Bool("uninstall", false, "If specified, uninstall nushell.")
	NushellCmd.Flags().BoolP("config", "c", false, "If specified, configure nushell.")
	// rootCmd.AddCommand(nushellCmd)
}
