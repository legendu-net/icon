package shell

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
)

// televisionConfigDir returns the configuration directory of Television,
// which is overridden by the environment variables TELEVISION_CONFIG and
// XDG_CONFIG_HOME (in this order of precedence).
func televisionConfigDir() string {
	if dir := os.Getenv("TELEVISION_CONFIG"); dir != "" {
		return dir
	}
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "television")
	}
	return "~/.config/television"
}

// installTelevision downloads the prebuilt Television binary (tv) from its
// GitHub releases and installs it into dirBin. It is used on Linux only.
// The static musl build is used on x86_64 since the gnu build requires a recent
// glibc (2.39), which older distributions (e.g., Ubuntu 22.04) lack.
// The gnu build is the only one available on aarch64.
func installTelevision(cmd *cobra.Command, dirBin string) {
	tmpdir := utils.CreateTempDir("")
	defer os.RemoveAll(tmpdir)
	file := filepath.Join(tmpdir, "tv.tar.gz")
	network.DownloadGitHubRelease(
		"alexpasmantier/television",
		"",
		map[string][]string{
			"common": {"tar.gz"},
			"amd64":  {"x86_64-unknown-linux-musl"},
			"arm64":  {"aarch64-unknown-linux-gnu"},
		},
		[]string{},
		file,
	)
	// The archive contains tv under a top-level directory named after the release.
	// --no-same-owner: the archive stores tv as owned by the release builder's uid.
	command := utils.Format(`{prefix} mkdir -p {dirBin} \
			&& {prefix} tar --no-same-owner --wildcards --strip-components=1 -zxvf {file} -C {dirBin} '*/tv'`, map[string]string{
		"file":   file,
		"dirBin": dirBin,
		"prefix": utils.GetCommandPrefix(utils.GetBoolFlag(cmd, "sudo"), map[string]uint32{
			dirBin: unix.W_OK | unix.R_OK,
		}),
	})
	utils.RunCmd(command)
}

// configTelevision symlinks (or copies) the configuration of Television tracked
// in icon-data into the configuration directory of Television. The entries of
// ~/.config/icon-data/television are handled one by one (instead of the
// directory as a whole) so that the cable directory, which
// `tv update-channels` writes into, stays local to the machine.
func configTelevision(cmd *cobra.Command) {
	icon.FetchConfigData(false, "")
	src := "~/.config/icon-data/television"
	if !utils.ExistsDir(src) {
		log.Fatalf("The Television configuration directory %s does not exist.", src)
	}
	configDir := televisionConfigDir()
	for _, entry := range utils.ReadDir(src) {
		dst := filepath.Join(configDir, entry.Name())
		utils.BackupOrRemove(dst, utils.ShouldBackup(cmd))
		utils.CopyOrSymlink(filepath.Join(src, entry.Name()), dst, utils.GetBoolFlag(cmd, "copy"))
	}
}

// Install and configure Television.
func television(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		// fd, rg and bat are required by the default channels of Television.
		utils.InstallPackages(utils.GetBoolFlag(cmd, "yes"), utils.PkgFd, utils.PkgRipgrep, utils.PkgBat)
		if utils.IsLinux() {
			installTelevision(cmd, utils.GetStringFlag(cmd, "bin-dir"))
		} else {
			utils.BrewInstallSafe([]string{"television"})
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		configTelevision(cmd)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		if utils.IsLinux() {
			utils.RemoveAll(filepath.Join(utils.GetStringFlag(cmd, "bin-dir"), "tv"))
		} else {
			utils.RunCmd("brew uninstall television")
		}
	}
}

var televisionCmd = &cobra.Command{
	Use:     "television",
	Aliases: []string{"tv"},
	Short:   "Install and configure Television.",
	//Args:  cobra.ExactArgs(1),
	Run: television,
}

func ConfigTelevisionCmd(rootCmd *cobra.Command) {
	televisionCmd.Flags().BoolP("install", "i", false, "Install Television.")
	televisionCmd.Flags().BoolP("uninstall", "u", false, "Uninstall Television.")
	televisionCmd.Flags().BoolP("config", "c", false, "Configure Television.")
	televisionCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	televisionCmd.Flags().Bool("sudo", false, "Force using sudo.")
	televisionCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	televisionCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	televisionCmd.Flags().String("bin-dir", "/usr/local/bin", "The directory for installing the Television executable (tv).")
	rootCmd.AddCommand(televisionCmd)
}
