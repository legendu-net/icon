package shell

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
)

const ghosttyAppImageRepo = "pkgforge-dev/ghostty-appimage"

// installGhosttyAppImage downloads the Ghostty AppImage from the
// pkgforge-dev/ghostty-appimage GitHub releases and installs it into
// ~/Applications. The AppImage is used on all Linux distributions because
// Ghostty has no single official package shared across the Debian/Ubuntu and
// Fedora series.
func installGhosttyAppImage() {
	tmpdir := utils.CreateTempDir("")
	defer os.RemoveAll(tmpdir)
	file := filepath.Join(tmpdir, "ghostty.AppImage")
	network.DownloadGitHubRelease(ghosttyAppImageRepo, "", map[string][]string{
		"common": {".AppImage"},
		"amd64":  {"x86_64"},
		"arm64":  {"aarch64"},
	}, []string{".zsync"}, file)
	dst := "~/Applications/ghostty.AppImage"
	utils.CopyFile(file, dst)
	utils.Chmod(dst, "+x")
}

// Install and configure the Ghostty terminal.
func ghostty(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		if utils.IsLinux() {
			installGhosttyAppImage()
		} else {
			utils.RunCmd("brew install --cask ghostty")
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		icon.FetchConfigData(false, "")
		src := "~/.config/icon-data/ghostty/config.ghostty"
		if !utils.ExistsFile(src) {
			log.Fatalf("The Ghostty configuration file %s does not exist.", src)
		}
		dst := "~/.config/ghostty/config"
		utils.BackupOrRemove(dst, utils.ShouldBackup(cmd))
		utils.CopyOrSymlink(src, dst, utils.GetBoolFlag(cmd, "copy"))
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		if utils.IsLinux() {
			utils.RemoveAll("~/Applications/ghostty.AppImage")
		} else {
			utils.RunCmd("brew uninstall --cask ghostty")
		}
	}
}

var ghosttyCmd = &cobra.Command{
	Use:     "ghostty",
	Aliases: []string{"ghost"},
	Short:   "Install and configure the Ghostty terminal.",
	//Args:  cobra.ExactArgs(1),
	Run: ghostty,
}

func ConfigGhosttyCmd(rootCmd *cobra.Command) {
	ghosttyCmd.Flags().BoolP("install", "i", false, "Install the Ghostty terminal.")
	ghosttyCmd.Flags().Bool("uninstall", false, "Uninstall the Ghostty terminal.")
	ghosttyCmd.Flags().BoolP("config", "c", false, "Configure the Ghostty terminal.")
	ghosttyCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	ghosttyCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	ghosttyCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	rootCmd.AddCommand(ghosttyCmd)
}
