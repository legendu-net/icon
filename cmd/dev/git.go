package dev

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
)

func installGitUI(cmd *cobra.Command) {
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

func linkGitUIFiles(baseDir string, backup, copyPath bool) {
	utils.SymlinkIntoDir("~/.config/icon-data/git/gitui/key_bindings.ron", filepath.Join(baseDir, "gitui"),
		backup, copyPath)
}

func configGitUI(cmd *cobra.Command) {
	if utils.GetBoolFlag(cmd, "gitui") {
		linkGitUIFiles("~/.config", !utils.GetBoolFlag(cmd, "no-backup"), utils.GetBoolFlag(cmd, "copy"))
		if utils.IsLinux() {
			baseDir := os.Getenv("XDG_CONFIG_HOME")
			if baseDir != "" {
				linkGitUIFiles(baseDir,
					!utils.GetBoolFlag(cmd, "no-backup"), utils.GetBoolFlag(cmd, "copy"))
			}
		}
	}
}

func installGitDelta() {
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
	command := utils.Format(`{prefix} tar -zxvf {file} \
			-C /usr/local/bin/ --wildcards --no-anchored delta --strip=1 \
		&& rm {file}`, map[string]string{
		"prefix": utils.GetCommandPrefix(
			true,
			map[string]uint32{},
		),
		"file": file,
	})
	utils.RunCmd(command)
}

func configGitProxy(cmd *cobra.Command) {
	git := utils.GetStringFlag(cmd, "git")
	proxy := utils.GetStringFlag(cmd, "proxy")
	if proxy != "" {
		command := utils.Format(`{git} config --global http.proxy {proxy} \
				&& {git} config --global https.proxy {proxy}`, map[string]string{
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
		if utils.IsLinux() {
			if utils.IsUniversalBlue() {
				utils.BrewInstallSafe([]string{"git-delta", "gitui"})
			} else if utils.IsDebianUbuntuSeries() {
				command := utils.Format(`{prefix} apt-get {yesStr} update \
						&& {prefix} apt-get {yesStr} install git git-lfs`, map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
				installGitDelta()
				installGitUI(cmd)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yesStr} install git", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
				installGitDelta()
				installGitUI(cmd)
			}
		} else {
			utils.BrewInstallSafe([]string{"git", "git-lfs", "git-delta", "gitui"})
		}
		command := utils.Format("{git} lfs install", map[string]string{
			"git": git,
		})
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
		icon.FetchConfigData(false, "")
		network.SSHClient(cmd, args)
		utils.Symlink("~/.config/icon-data/git/gitconfig", "~/.gitconfig",
			!utils.GetBoolFlag(cmd, "no-backup"), utils.GetBoolFlag(cmd, "copy"))
		configGitProxy(cmd)
		configGitUI(cmd)
	}
	configureGitIgnore(cmd)
	if utils.GetBoolFlag(cmd, "uninstall") {
		command := utils.Format("{git} lfs uninstall", map[string]string{
			"git": git,
		})
		utils.RunCmd(command)
		if utils.IsLinux() {
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format("{prefix} apt-get {yesStr} purge git git-lfs", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yesStr} remove git", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
		} else {
			utils.RunCmd("brew uninstall git git-lfs")
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

var gitCmd = &cobra.Command{
	Use:     "git",
	Aliases: []string{},
	Short:   "Install and configure Git.",
	//Args:  cobra.ExactArgs(1),
	Run: git,
}

func ConfigGitCmd(rootCmd *cobra.Command) {
	gitCmd.Flags().BoolP("install", "i", false, "Install Git.")
	gitCmd.Flags().Bool("uninstall", false, "Uninstall Git.")
	gitCmd.Flags().BoolP("config", "c", false, "Configure Git.")
	gitCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	gitCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	gitCmd.Flags().String("git", "git", "Path to the Git command.")
	gitCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	gitCmd.Flags().Bool("gitui", false, "Install and configure gitui too.")
	gitCmd.Flags().String("proxy", "", "Configure Git to use the specified proxy.")
	gitCmd.Flags().StringP("dest-dir", "d", ".", "The destination directory (current directory, by default) to copy gitignore files to.")
	gitCmd.Flags().BoolP("append", "a", false, "Append to the .gitignore instead of oveerwriting it.")
	gitCmd.Flags().StringP("lang", "l", "", "The language to configure .gitignore for.")
	rootCmd.AddCommand(gitCmd)
}
