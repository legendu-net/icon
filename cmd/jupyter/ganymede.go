package jupyter

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
)

// Install and configure Ganymede.
func ganymede(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		tmpdir := utils.CreateTempDir("")
		defer os.RemoveAll(tmpdir)
		file := filepath.Join(tmpdir, "ganymede.jar")
		network.DownloadGitHubRelease(
			"allen-ball/ganymede",
			"",
			map[string][]string{"common": {"jar"}},
			[]string{"asc"},
			file,
		)
		command := utils.Format(`{prefix} java -jar {file} -i --sys-prefix \
				&& {prefix} cp -r /usr/share/jupyter/kernels/ganymede-*-java-* /usr/local/share/jupyter/kernels/ \
				&& {prefix} sed -i 's_/usr/share/jupyter/kernels/_/usr/local/share/jupyter/kernels/_g' /usr/local/share/jupyter/kernels/ganymede*/kernel.json \
			`, map[string]string{
			"prefix": utils.GetCommandPrefix(false, map[string]uint32{
				"/usr/share/jupyter/":       unix.W_OK | unix.R_OK,
				"/usr/local/share/jupyter/": unix.W_OK | unix.R_OK,
			}),
			"file": file,
		})
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
	}
}

var GanymedeCmd = &cobra.Command{
	Use:     "ganymede",
	Aliases: []string{"gmd"},
	Short:   "Install and configure Ganymede.",
	//Args:  cobra.ExactArgs(1),
	Run: ganymede,
}

func init() {
	GanymedeCmd.Flags().BoolP("install", "i", false, "Install Ganymede.")
	GanymedeCmd.Flags().Bool("uninstall", false, "Uninstall Ganymede.")
	GanymedeCmd.Flags().BoolP("config", "c", false, "Configure Ganymede.")
	GanymedeCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	GanymedeCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	GanymedeCmd.Flags().Bool("sudo", false, "Force using sudo.")
	GanymedeCmd.Flags().String("profile-dir", filepath.Join(utils.UserHomeDir(), ".ipython"), "The directory for storing IPython configuration files.")
	utils.AddPythonFlags(GanymedeCmd)
}
