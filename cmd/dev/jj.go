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
	binDir := "~/.local/bin"
	if global {
		prefix = utils.GetCommandPrefix(
			true,
			map[string]uint32{},
		)
		binDir = "/usr/local/bin"
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

// uninstallJj removes the jj binary installed by installJj. It searches the
// candidate installation directories (~/.local/bin and /usr/local/bin) and
// removes the binary wherever it is found, using privilege escalation only
// for the system location.
func uninstallJj() {
	for _, dir := range []struct {
		binDir string
		global bool
	}{
		{"~/.local/bin", false},
		{"/usr/local/bin", true},
	} {
		path := dir.binDir + "/jj"
		if !utils.ExistsFile(path) {
			continue
		}
		prefix := ""
		if dir.global {
			prefix = utils.GetCommandPrefix(
				true,
				map[string]uint32{},
			)
		}
		command := utils.Format("{prefix} rm -f {path}", map[string]string{
			"prefix": prefix,
			"path":   path,
		})
		utils.RunCmd(command)
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

// Install and configure jj (Jujutsu).
func jj(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		if utils.IsLinux() {
			if utils.IsUniversalBlue() {
				if utils.GetBoolFlag(cmd, "global") {
					log.Print("WARNING: --global is not respected on Universal Blue; jj is installed into ~/.local/bin.")
				}
				installJj(false)
			} else if utils.IsDebianUbuntuSeries() {
				installJj(utils.GetBoolFlag(cmd, "global"))
			} else if utils.IsFedoraSeries() {
				installJj(utils.GetBoolFlag(cmd, "global"))
			}
		} else {
			if utils.GetBoolFlag(cmd, "global") {
				log.Print("WARNING: --global is not respected on macOS; jj is installed into ~/.local/bin.")
			}
			installJj(false)
		}
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
	Short:   "Install and configure jj (Jujutsu).",
	//Args:  cobra.ExactArgs(1),
	Run: jj,
}

func ConfigJjCmd(rootCmd *cobra.Command) {
	jjCmd.Flags().BoolP("install", "i", false, "Install jj.")
	jjCmd.Flags().Bool("uninstall", false, "Uninstall jj.")
	jjCmd.Flags().BoolP("config", "c", false, "Configure jj.")
	jjCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	jjCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	jjCmd.Flags().Bool("global", false, "Install jj into /usr/local/bin instead of ~/.local/bin.")
	jjCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	rootCmd.AddCommand(jjCmd)
}
