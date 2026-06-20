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

const wavetermRepo = "wavetermdev/waveterm"

// installWavetermDeb downloads the Wave terminal .deb package from its GitHub
// releases and installs it on the Debian/Ubuntu series.
func installWavetermDeb(cmd *cobra.Command) {
	tmpdir := utils.CreateTempDir("")
	defer os.RemoveAll(tmpdir)
	file := filepath.Join(tmpdir, "waveterm.deb")
	network.DownloadGitHubRelease(wavetermRepo, "", map[string][]string{
		"common": {".deb"},
		"amd64":  {"amd64"},
		"arm64":  {"arm64"},
	}, []string{}, file)
	command := utils.Format("{prefix} apt-get {yesStr} install {file}", map[string]string{
		"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
		"yesStr": utils.BuildYesFlag(cmd),
		"file":   file,
	})
	utils.RunCmd(command)
}

// installWavetermRpm downloads the Wave terminal .rpm package from its GitHub
// releases and installs it on the Fedora series.
func installWavetermRpm(cmd *cobra.Command) {
	tmpdir := utils.CreateTempDir("")
	defer os.RemoveAll(tmpdir)
	file := filepath.Join(tmpdir, "waveterm.rpm")
	network.DownloadGitHubRelease(wavetermRepo, "", map[string][]string{
		"common": {".rpm"},
		"amd64":  {"x86_64"},
		"arm64":  {"aarch64"},
	}, []string{}, file)
	command := utils.Format("{prefix} dnf {yesStr} install {file}", map[string]string{
		"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
		"yesStr": utils.BuildYesFlag(cmd),
		"file":   file,
	})
	utils.RunCmd(command)
}

// installWavetermAppImage downloads the Wave terminal AppImage from its GitHub
// releases and installs it into ~/Applications. AppImage is used on image-based
// Universal Blue distributions where layering deb/rpm packages is undesirable.
func installWavetermAppImage() {
	tmpdir := utils.CreateTempDir("")
	defer os.RemoveAll(tmpdir)
	file := filepath.Join(tmpdir, "waveterm.AppImage")
	network.DownloadGitHubRelease(wavetermRepo, "", map[string][]string{
		"common": {".AppImage"},
		"amd64":  {"x86_64"},
		"arm64":  {"arm64"},
	}, []string{}, file)
	appDir := "~/Applications"
	command := utils.Format(`mkdir -p {appDir} \
		&& cp {file} {appDir}/waveterm.AppImage \
		&& chmod +x {appDir}/waveterm.AppImage`, map[string]string{
		"appDir": appDir,
		"file":   file,
	})
	utils.RunCmd(command)
}

// uninstallWavetermDeb removes the Wave terminal package on the Debian/Ubuntu series.
func uninstallWavetermDeb(cmd *cobra.Command) {
	command := utils.Format("{prefix} apt-get {yesStr} purge waveterm", map[string]string{
		"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
		"yesStr": utils.BuildYesFlag(cmd),
	})
	utils.RunCmd(command)
}

// uninstallWavetermRpm removes the Wave terminal package on the Fedora series.
func uninstallWavetermRpm(cmd *cobra.Command) {
	command := utils.Format("{prefix} dnf {yesStr} remove waveterm", map[string]string{
		"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
		"yesStr": utils.BuildYesFlag(cmd),
	})
	utils.RunCmd(command)
}

// Install and configure the Wave terminal.
func waveterm(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		if utils.IsLinux() {
			if utils.IsUniversalBlue() {
				installWavetermAppImage()
			} else if utils.IsDebianUbuntuSeries() {
				installWavetermDeb(cmd)
			} else if utils.IsFedoraSeries() {
				installWavetermRpm(cmd)
			}
		} else {
			utils.RunCmd("brew install --cask wave")
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		icon.FetchConfigData(false, "")
		src := "~/.config/icon-data/waveterm/settings.json"
		if !utils.ExistsFile(src) {
			log.Fatalf("The Wave terminal configuration file %s does not exist.", src)
		}
		dst := "~/.config/waveterm/settings.json"
		utils.BackupOrRemove(dst, utils.ShouldBackup(cmd))
		utils.CopyOrSymlink(src, dst, utils.GetBoolFlag(cmd, "copy"))
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		if utils.IsLinux() {
			if utils.IsUniversalBlue() {
				utils.RemoveAll("~/Applications/waveterm.AppImage")
			} else if utils.IsDebianUbuntuSeries() {
				uninstallWavetermDeb(cmd)
			} else if utils.IsFedoraSeries() {
				uninstallWavetermRpm(cmd)
			}
		} else {
			utils.RunCmd("brew uninstall --cask wave")
		}
	}
}

var wavetermCmd = &cobra.Command{
	Use:     "waveterm",
	Aliases: []string{"wave"},
	Short:   "Install and configure the Wave terminal.",
	//Args:  cobra.ExactArgs(1),
	Run: waveterm,
}

func ConfigWavetermCmd(rootCmd *cobra.Command) {
	wavetermCmd.Flags().BoolP("install", "i", false, "Install the Wave terminal.")
	wavetermCmd.Flags().Bool("uninstall", false, "Uninstall the Wave terminal.")
	wavetermCmd.Flags().BoolP("config", "c", false, "Configure the Wave terminal.")
	wavetermCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	wavetermCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	wavetermCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	rootCmd.AddCommand(wavetermCmd)
}
