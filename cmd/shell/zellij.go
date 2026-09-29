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

// zellijConfigDir returns the configuration directory of Zellij,
// which is overridden by the environment variable ZELLIJ_CONFIG_DIR.
func zellijConfigDir() string {
	if dir := os.Getenv("ZELLIJ_CONFIG_DIR"); dir != "" {
		return dir
	}
	return "~/.config/zellij"
}

// Install and configure Zellij.
func zellij(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		tmpdir := utils.CreateTempDir("")
		defer os.RemoveAll(tmpdir)
		file := filepath.Join(tmpdir, "zellij.tar.gz")
		network.DownloadGitHubRelease(
			"zellij-org/zellij",
			"",
			map[string][]string{
				"common": {"tar.gz"},
				"amd64":  {"x86_64"},
				"arm64":  {"aarch64"},
				"linux":  {"linux", "musl"},
				"darwin": {"apple", "darwin"},
			},
			[]string{"sha256sum", "no-web"},
			file,
		)
		dirBin := utils.GetStringFlag(cmd, "bin-dir")
		// --no-same-owner: the archive stores zellij as owned by the release builder's uid.
		command := utils.Format(`{prefix} mkdir -p {dirBin} \
				&& {prefix} tar --no-same-owner -zxvf {file} -C {dirBin}`, map[string]string{
			"file":   file,
			"dirBin": dirBin,
			"prefix": utils.GetCommandPrefix(utils.GetBoolFlag(cmd, "sudo"), map[string]uint32{
				dirBin: unix.W_OK | unix.R_OK,
			}),
		})
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
		icon.FetchConfigData(false, "")
		src := "~/.config/icon-data/zellij"
		dst := zellijConfigDir()
		utils.BackupOrRemove(dst, utils.ShouldBackup(cmd))
		utils.CopyOrSymlink(src, dst, utils.GetBoolFlag(cmd, "copy"))
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		utils.RemoveAll(filepath.Join(utils.GetStringFlag(cmd, "bin-dir"), "zellij"))
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
	zellijCmd.Flags().BoolP("install", "i", false, "Install Zellij.")
	zellijCmd.Flags().Bool("uninstall", false, "Uninstall Zellij.")
	zellijCmd.Flags().BoolP("config", "c", false, "Configure Zellij.")
	zellijCmd.Flags().Bool("sudo", false, "Force using sudo.")
	zellijCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	zellijCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	zellijCmd.Flags().String("bin-dir", "/usr/local/bin", "The directory for installing Zellij executable.")
	rootCmd.AddCommand(zellijCmd)
}
