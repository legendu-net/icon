package dev

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
)

func installGitUi(cmd *cobra.Command) {
	if utils.GetBoolFlag(cmd, "gitui") {
		tmpdir := utils.CreateTempDir("")
		defer os.RemoveAll(tmpdir)
		file := filepath.Join(tmpdir, "gitui.tar.gz")
		network.DownloadGitHubRelease("extrawurst/gitui", "", map[string][]string{
			"common": {"tar.gz"},
			"linux":  {"linux"},
			"darwin": {"mac"},
			"amd64":  {"musl"},
			"arm64":  {"aarch64"},
		}, []string{}, file)
		command := utils.Format(`{prefix} tar -zxvf {file} -C /usr/local/bin/`, map[string]string{
			"prefix": utils.GetCommandPrefix(
				true,
				map[string]uint32{},
			),
			"file": file,
		})
		utils.RunCmd(command)
	}
}

func linkGitUiFiles(baseDir string, backup bool, copy bool) {
	utils.SymlinkIntoDir("~/.config/icon-data/git/gitui/key_bindings.ron", filepath.Join(baseDir, "gitui"),
		backup, copy)
}

func configGitUi(cmd *cobra.Command) {
	if utils.GetBoolFlag(cmd, "gitui") {
		linkGitUiFiles("~/.config", !utils.GetBoolFlag(cmd, "no-backup"), utils.GetBoolFlag(cmd, "copy"))
		if utils.IsLinux() {
			baseDir := os.Getenv("XDG_CONFIG_HOME")
			if baseDir != "" {
				linkGitUiFiles(baseDir,
					!utils.GetBoolFlag(cmd, "no-backup"), utils.GetBoolFlag(cmd, "copy"))
			}
		}
	}
}

func installGitDelta(cmd *cobra.Command) {
	tmpdir := utils.CreateTempDir("")
	defer os.RemoveAll(tmpdir)
	file := filepath.Join(tmpdir, "git-delta.tar.gz")
	network.DownloadGitHubRelease("dandavison/delta", "", map[string][]string{
		"common": {},
		"amd64":  {"x86_64"},
		"arm64":  {"aarch64"},
		"linux":  {"linux", "gnu"},
		"darwin": {"apple", "darwin"},
	}, []string{}, file)
	command := utils.Format(`{prefix} tar -zxvf {file} -C /usr/local/bin/ --wildcards --no-anchored delta --strip=1 \
		&& rm {file}`, map[string]string{
		"prefix": utils.GetCommandPrefix(
			true,
			map[string]uint32{},
		),
		"yes_s": utils.BuildYesFlag(cmd),
		"file":  file,
	})
	utils.RunCmd(command)
}

func configGitProxy(cmd *cobra.Command) {
	git := utils.GetStringFlag(cmd, "git")
	proxy := utils.GetStringFlag(cmd, "proxy")
	if proxy != "" {
		command := utils.Format("{git} config --global http.proxy {proxy} && {git} config --global https.proxy {proxy}", map[string]string{
			"proxy": proxy,
			"git":   git,
		})
		utils.RunCmd(command)
	}
}

// Install and configure Git.
func git(cmd *cobra.Command, args []string) {
	git := utils.GetStringFlag(cmd, "git")
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "linux":
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format(`{prefix} apt-get update && {prefix} apt-get install {yes_s} git git-lfs`, map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yes_s} install git", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
			installGitDelta(cmd)
			installGitUi(cmd)
		case "darwin":
			utils.BrewInstallSafe([]string{"git", "git-lfs"})
		default:
		}
		command := utils.Format("{git} lfs install", map[string]string{
			"git": git,
		})
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
		icon.FetchConfigData(false, "")
		network.SshClient(cmd, args)
		utils.Symlink("~/.config/icon-data/git/gitconfig", "~/.gitconfig",
			!utils.GetBoolFlag(cmd, "no-backup"), utils.GetBoolFlag(cmd, "copy"))
		configGitProxy(cmd)
		configGitUi(cmd)
	}
	configureGitIgnore(cmd)
	if utils.GetBoolFlag(cmd, "uninstall") {
		command := utils.Format("{git} lfs uninstall", map[string]string{
			"git": git,
		})
		utils.RunCmd(command)
		switch runtime.GOOS {
		case "darwin":
			utils.RunCmd("brew uninstall git git-lfs")
		case "linux":
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format("{prefix} apt-get purge {yes_s} git git-lfs", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yes_s} remove git", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
				})
				utils.RunCmd(command)
			}
		default:
		}
	}
}

// Insert patterns to ingore into .gitignore in the current directory.
func configureGitIgnore(cmd *cobra.Command) {
	lang := strings.ToLower(utils.GetStringFlag(cmd, "lang"))
	if lang == "" {
		return
	}
	srcFile := "~/.config/icon-data/git/gitignore_" + lang
	dstDir := utils.GetStringFlag(cmd, "dest-dir")
	dstFile := filepath.Join(dstDir, ".gitignore")
	if utils.GetBoolFlag(cmd, "append") {
		utils.AppendToTextFile(dstFile, utils.ReadFileAsString(srcFile), true)
		log.Printf("%s is appended into %s.", srcFile, dstFile)
	} else {
		utils.CopyFile(srcFile, dstFile)
	}
}

var GitCmd = &cobra.Command{
	Use:     "git",
	Aliases: []string{},
	Short:   "Install and configure Git.",
	//Args:  cobra.ExactArgs(1),
	Run: git,
}

func init() {
	GitCmd.Flags().BoolP("install", "i", false, "Install Git.")
	GitCmd.Flags().Bool("uninstall", false, "Uninstall Git.")
	GitCmd.Flags().BoolP("config", "c", false, "Configure Git.")
	GitCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	GitCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	GitCmd.Flags().String("git", "git", "Path to the Git command.")
	GitCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	GitCmd.Flags().Bool("gitui", false, "Install and configure gitui too.")
	GitCmd.Flags().String("proxy", "", "Configure Git to use the specified proxy.")
	GitCmd.Flags().StringP("dest-dir", "d", ".", "The destination directory (current directory, by default) to copy gitignore files to.")
	GitCmd.Flags().BoolP("append", "a", false, "Append to the .gitignore instead of oveerwriting it.")
	GitCmd.Flags().StringP("lang", "l", "", "The language to configure .gitignore for.")
}
