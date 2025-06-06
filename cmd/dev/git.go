package dev

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
)

var USER = utils.GetCurrentUser().Username

func getGitUserName(cmd *cobra.Command) string {
	user := utils.GetStringFlag(cmd, "user-name")
	if user != "" {
		return user
	}
	if utils.GetBoolFlag(cmd, "yes") {
		return USER
	}
	fmt.Printf("Please enter the user name for Git: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	err := scanner.Err()
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return scanner.Text()
}

func getGitUserEmail(cmd *cobra.Command) string {
	email := utils.GetStringFlag(cmd, "user-email")
	if email != "" {
		return email
	}
	if utils.GetBoolFlag(cmd, "yes") {
		return USER + "@example.com"
	}
	fmt.Printf("Please enter the user email for Git: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	err := scanner.Err()
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return scanner.Text()
}

func installGitUi(cmd *cobra.Command) {
	if utils.GetBoolFlag(cmd, "gitui") {
		tmpdir := utils.CreateTempDir("")
		defer os.RemoveAll(tmpdir)
		file := filepath.Join(tmpdir, "gitui.tar.gz")
		network.DownloadGitHubRelease("extrawurst/gitui", "", map[string][]string{
			"common": {"tar.gz"},
			"linux":  {"linux"},
			"darwin": {"mac"},
			"x86_64": {"musl"},
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

func configGitUiHelper(baseDir string) {
	dstDir := filepath.Join(baseDir, "gitui/")
	utils.MkdirAll(dstDir, 0o700)
	utils.CopyEmbeddedFileToDir("data/git/gitui/key_bindings.ron", dstDir, 0o600, true)
}

func configGitUi(cmd *cobra.Command) {
	if utils.GetBoolFlag(cmd, "gitui") {
		home := utils.UserHomeDir()
		configGitUiHelper(filepath.Join(home, ".config/"))
		if utils.IsLinux() {
			baseDir := os.Getenv("XDG_CONFIG_HOME")
			if baseDir != "" {
				configGitUiHelper(filepath.Join(baseDir))
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
		"x86_64": {"x86_64"},
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

func configGitBashCompletion(cmd *cobra.Command) {
	switch runtime.GOOS {
	case "darwin":
		file := "/usr/local/etc/bash_completion.d/git-completion.bash"
		script := utils.Format("\n# Git completion\n[ -f {file} ] &&  . {file}", map[string]string{
			"file": file,
		})
		home := utils.UserHomeDir()
		utils.AppendToTextFile(filepath.Join(home, ".bash_profile"), script, true)
		log.Printf("Bash completion is enabled for Git.")
	default:
	}
}

func configGitUser(cmd *cobra.Command) {
	// user.name and user.email
	git := utils.GetStringFlag(cmd, "git")
	command := utils.Format(`{git} config --global user.name "{name}" \
		&& {git} config --global user.email "{email}"`, map[string]string{
		"name":  getGitUserName(cmd),
		"email": getGitUserEmail(cmd),
		"git":   git,
	})
	utils.RunCmd(command)
}

func createGitConfig() {
	home := utils.UserHomeDir()
	gitConfig := filepath.Join(home, ".gitconfig")
	utils.CopyEmbeddedFile("data/git/gitconfig", gitConfig, 0o600, true)
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
			utils.BrewInstallSafe([]string{"git", "git-lfs", "bash-completion@2"})
		default:
		}
		command := utils.Format("{git} lfs install", map[string]string{
			"git": git,
		})
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
		network.SshClient(cmd, args)
		createGitConfig()
		configGitUser(cmd)
		configGitBashCompletion(cmd)
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
	srcFile := "data/git/gitignore_" + lang
	dstDir := utils.GetStringFlag(cmd, "dest-dir")
	dstFile := filepath.Join(dstDir, ".gitignore")
	if utils.GetBoolFlag(cmd, "append") {
		utils.AppendToTextFile(dstFile, utils.ReadEmbeddedFileAsString(srcFile), true)
		log.Printf("%s is appended into %s.", srcFile, dstFile)
	} else {
		utils.CopyEmbeddedFile(srcFile, dstFile, 0o600, true)
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
	GitCmd.Flags().String("git", "git", "Path to the Git command.")
	GitCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	GitCmd.Flags().Bool("gitui", false, "Install and configure gitui too.")
	GitCmd.Flags().StringP("user-name", "n", "", "The user name for Git.")
	GitCmd.Flags().StringP("user-email", "e", "", "The user name for Git.")
	GitCmd.Flags().String("proxy", "", "Configure Git to use the specified proxy.")
	GitCmd.Flags().StringP("dest-dir", "d", ".", "The destination directory (current directory, by default) to copy gitignore files to.")
	GitCmd.Flags().BoolP("append", "a", false, "Append to the .gitignore instead of oveerwriting it.")
	GitCmd.Flags().StringP("lang", "l", "", "The language to configure .gitignore for.")
	// rootCmd.AddCommand(gitCmd)
}
