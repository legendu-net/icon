package shell

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
)

// Install and configure Ganymede.
func zellij(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		tmpdir := utils.CreateTempDir("")
		defer os.RemoveAll(tmpdir)
		file := filepath.Join(tmpdir, "zellij.tar.gz")
		network.DownloadGitHubRelease(
			"zellij-org/zellij",
			"",
			map[string][]string{"common": {"tar.gz"}},
			[]string{"sha256sum"},
			file,
		)
		dir_bin := utils.GetStringFlag(cmd, "bin-dir")
		command := utils.Format(`{prefix} tar -zxvf {file} -C {dir_bin}`, map[string]string{
			"file":    file,
			"dir_bin": dir_bin,
			"prefix": utils.GetCommandPrefix(false, map[string]uint32{
				dir_bin: unix.W_OK | unix.R_OK,
			}),
		})
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
		icon.FetchConfigData(false, "")
		utils.Symlink("~/.config/icon-data/zellij/config.kdl", "~/.config/zellij/config.kdl", true)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
	}
}

var ZellijCmd = &cobra.Command{
	Use:     "zellij",
	Aliases: []string{"zj", "z"},
	Short:   "Install and configure Zellij.",
	//Args:  cobra.ExactArgs(1),
	Run: zellij,
}

func init() {
	ZellijCmd.Flags().BoolP("install", "i", false, "Install Ganymede.")
	ZellijCmd.Flags().Bool("uninstall", false, "Uninstall Ganymede.")
	ZellijCmd.Flags().BoolP("config", "c", false, "Configure Ganymede.")
	ZellijCmd.Flags().Bool("sudo", false, "Force using sudo.")
	ZellijCmd.Flags().String("bin-dir", "/usr/local/bin", "The directory for installing Zellij executable.")
	utils.AddPythonFlags(ZellijCmd)
	// rootCmd.AddCommand(ZellijCmd)
}
