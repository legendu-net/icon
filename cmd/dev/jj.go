package dev

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
)

// jjBinDir returns the directory into which the jj tools are installed, which is
// /usr/local/bin when global is true and ~/.local/bin otherwise.
func jjBinDir(global bool) string {
	if global {
		return "/usr/local/bin"
	}
	return "~/.local/bin"
}

// installJj downloads the prebuilt jj binary from its GitHub releases and
// installs it. When global is true the binary is installed into /usr/local/bin;
// otherwise it is installed into ~/.local/bin (no privilege escalation).
// jj is not reliably packaged in the Debian/Ubuntu and Fedora series, so the
// official static binary is used.
func installJj(global bool) {
	tmpdir := utils.CreateTempDir("")
	defer os.RemoveAll(tmpdir)
	file := filepath.Join(tmpdir, "jj.tar.gz")
	network.DownloadGitHubRelease("jj-vcs/jj", "", map[string][]string{
		"common": {"tar.gz"},
		"amd64":  {"x86_64"},
		"arm64":  {"aarch64"},
		"linux":  {"linux", "musl"},
		"darwin": {"apple", "darwin"},
	}, []string{}, file)
	prefix := ""
	binDir := jjBinDir(global)
	if global {
		prefix = utils.GetCommandPrefix(
			true,
			map[string]uint32{},
		)
	}
	command := utils.Format(`{prefix} mkdir -p {binDir} \
		&& {prefix} tar -zxvf {file} -C {binDir} --strip-components=1 \
			--exclude=LICENSE --exclude='README.*' --exclude=CHANGELOG.md`, map[string]string{
		"prefix": prefix,
		"binDir": binDir,
		"file":   file,
	})
	utils.RunCmd(command)
}

// installJjui downloads the prebuilt jjui binary from its GitHub releases and
// installs it as {binDir}/jjui. jjui is a terminal UI for jj and is not
// packaged in the Debian/Ubuntu and Fedora series, so the official prebuilt
// binary is used. The release archive contains a single binary whose name
// carries the version and the platform (e.g. jjui-0.10.9-linux-amd64), so it is
// extracted into a temporary directory and copied to its final name from there.
// Privilege escalation for a global installation is derived by CopyFile and
// Chmod from the write permissions of the destination.
func installJjui(global bool) {
	tmpdir := utils.CreateTempDir("")
	defer os.RemoveAll(tmpdir)
	file := filepath.Join(tmpdir, "jjui.zip")
	network.DownloadGitHubRelease("idursun/jjui", "", map[string][]string{
		"common": {".zip"},
		"amd64":  {"amd64"},
		"arm64":  {"arm64"},
		"linux":  {"linux"},
		"darwin": {"darwin"},
	}, []string{}, file)
	extractDir := filepath.Join(tmpdir, "jjui")
	utils.RunCmd(utils.Format("unzip -o -j {file} 'jjui*' -d {extractDir}", map[string]string{
		"file":       file,
		"extractDir": extractDir,
	}))
	dst := jjBinDir(global) + "/jjui"
	utils.CopyFile(extractedJjuiBin(extractDir), dst)
	utils.Chmod(dst, "755")
}

// extractedJjuiBin returns the path of the jjui binary extracted into
// extractDir, which is the only jjui* entry the release archive of jjui
// contains.
func extractedJjuiBin(extractDir string) string {
	entries := utils.ReadDir(extractDir)
	if len(entries) != 1 {
		log.Fatalf("Expected exactly 1 jjui binary in the release archive of jjui but got %d.", len(entries))
	}
	return filepath.Join(extractDir, entries[0].Name())
}

// installJjBinaries installs the jj and jjui binaries into the directory chosen
// by global. unzip is checked upfront (rather than in installJjui) so that a
// machine without it does not end up with jj installed but jjui missing.
func installJjBinaries(global bool) {
	if !utils.ExistsCommand("unzip") {
		log.Fatal("The command unzip is required to extract the release archive of jjui but is not available.")
	}
	installJj(global)
	installJjui(global)
}

// InstallJjTools installs jj together with jjui, the terminal UI for jj, into the
// location permitted by the current platform. It is exported because jjui is
// required at runtime by the Yazi plugin Adda0/jjui, so that `icon yazi
// --install` installs the tools exactly the way `icon jj --install` does.
func InstallJjTools(global bool) {
	if !utils.IsLinux() {
		if global {
			log.Print("WARNING: --global is not respected on macOS; jj and jjui are installed into ~/.local/bin.")
		}
		installJjBinaries(false)
		return
	}
	switch {
	case utils.IsUniversalBlue():
		if global {
			log.Print("WARNING: --global is not respected on Universal Blue; jj and jjui are installed into ~/.local/bin.")
		}
		installJjBinaries(false)
	case utils.IsDebianUbuntuSeries(), utils.IsFedoraSeries():
		installJjBinaries(global)
	default:
		log.Print("WARNING: the Linux distribution is not supported, so jj and jjui are not installed.")
	}
}

// uninstallJj removes the jj and jjui binaries installed by InstallJjTools. It
// searches the candidate installation directories (~/.local/bin and
// /usr/local/bin) and removes the binaries wherever they are found. RemoveAll
// skips non-existent paths and derives privilege escalation from the target
// path's write permissions, so the system location is handled with sudo only
// when necessary.
func uninstallJj() {
	for _, binDir := range []string{"~/.local/bin", "/usr/local/bin"} {
		for _, bin := range []string{"jj", "jjui"} {
			utils.RemoveAll(binDir + "/" + bin)
		}
	}
}

// resolveJj returns the command used to invoke jj. It prefers the jj found on
// PATH; if jj is not on PATH (e.g. just installed into ~/.local/bin which is not
// yet in the current shell's PATH), it falls back to the absolute path in the
// candidate installation directories.
func resolveJj() string {
	if path := utils.LookPath("jj"); path != "" {
		return path
	}
	for _, dir := range []string{"~/.local/bin", "/usr/local/bin"} {
		path := dir + "/jj"
		if utils.ExistsFile(path) {
			return utils.NormalizePath(path)
		}
	}
	return "jj"
}

// Install and configure jj (Jujutsu) and jjui.
func jj(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		// unzip is required to extract the release archive of jjui.
		utils.InstallPackages(utils.GetBoolFlag(cmd, "yes"), utils.PkgUnzip)
		InstallJjTools(utils.GetBoolFlag(cmd, "global"))
	}
	if utils.GetBoolFlag(cmd, "config") {
		icon.FetchConfigData(false, "")
		cfg := utils.ReadUserConfig()
		jjBin := resolveJj()
		utils.RunCmd(utils.Format(
			`{jjBin} config set --user user.name "{userName}"`,
			map[string]string{"jjBin": jjBin, "userName": cfg.UserName},
		))
		utils.RunCmd(utils.Format(
			`{jjBin} config set --user user.email "{userEmail}"`,
			map[string]string{"jjBin": jjBin, "userEmail": cfg.UserEmail},
		))
		utils.RunCmd(utils.Format(
			`{jjBin} config set --user ui.diff-editor :builtin`,
			map[string]string{"jjBin": jjBin},
		))
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		uninstallJj()
	}
}

var jjCmd = &cobra.Command{
	Use:     "jj",
	Aliases: []string{},
	Short:   "Install and configure jj (Jujutsu) and jjui.",
	//Args:  cobra.ExactArgs(1),
	Run: jj,
}

func ConfigJjCmd(rootCmd *cobra.Command) {
	jjCmd.Flags().BoolP("install", "i", false, "Install jj and jjui.")
	jjCmd.Flags().Bool("uninstall", false, "Uninstall jj and jjui.")
	jjCmd.Flags().BoolP("config", "c", false, "Configure jj.")
	jjCmd.Flags().Bool("global", false, "Install jj and jjui into /usr/local/bin instead of ~/.local/bin.")
	jjCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	rootCmd.AddCommand(jjCmd)
}
