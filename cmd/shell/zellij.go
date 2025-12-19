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
		dirBin := utils.GetStringFlag(cmd, "bin-dir")
		command := utils.Format(`{prefix} tar -zxvf {file} -C {dirBin}`, map[string]string{
			"file":   file,
			"dirBin": dirBin,
			"prefix": utils.GetCommandPrefix(false, map[string]uint32{
				dirBin: unix.W_OK | unix.R_OK,
			}),
		})
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
		icon.FetchConfigData(false, "")
		utils.Symlink("~/.config/icon-data/zellij/config.kdl", "~/.config/zellij/config.kdl",
			!utils.GetBoolFlag(cmd, "no-backup"), utils.GetBoolFlag(cmd, "copy"))
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
	}
}

var zellijCmd = &cobra.Command{
	Use:     "zellij",
	Aliases: []string{"zj", "z"},
	Short:   "Install and configure Zellij.",
	//Args:  cobra.ExactArgs(1),
	Run: zellij,
}

func ConfigZellijCmd(rootCmd *cobra.Command) {
	zellijCmd.Flags().BoolP("install", "i", false, "Install Ganymede.")
	zellijCmd.Flags().Bool("uninstall", false, "Uninstall Ganymede.")
	zellijCmd.Flags().BoolP("config", "c", false, "Configure Ganymede.")
	zellijCmd.Flags().Bool("sudo", false, "Force using sudo.")
	zellijCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	zellijCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	zellijCmd.Flags().String("bin-dir", "/usr/local/bin", "The directory for installing Zellij executable.")
	utils.AddPythonFlags(zellijCmd)
	rootCmd.AddCommand(zellijCmd)
}
