package dev

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
	"log"
	"path/filepath"
	"runtime"
	"strings"
)

// Install and configure Git.
func git(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "linux":
			if utils.IsDebianSeries() {
				command := utils.Format("{prefix} apt-get update && {prefix} apt-get install {yes_s} git git-lfs", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
				// TODO: leverage from_github to download git-delta and install it to /usr/local/bin!!!
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} yum install {yes_s} git", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
					"yes_s": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
		case "darwin":
			utils.BrewInstallSafe([]string{"git", "git-lfs", "bash-completion@2"})
		default:
		}
		utils.RunCmd("git lfs install")
	}
	if utils.GetBoolFlag(cmd, "config") {
		network.SshClient(cmd, args)
		home := utils.UserHomeDir()
		gitConfig := filepath.Join(home, ".gitconfig")
		utils.CopyEmbedFile("data/git/gitconfig", gitConfig, 0o600)
		log.Printf("%s is copied to %s", "data/git/gitconfig", gitConfig)
		switch runtime.GOOS {
		case "darwin":
			file := "/usr/local/etc/bash_completion.d/git-completion.bash"
			script := utils.Format("\n# Git completion\n[ -f {file} ] &&  . {file}", map[string]string{
				"file": file,
			})
			utils.AppendToTextFile(filepath.Join(home, ".bash_profile"), script)
			log.Printf("Bash completion is enabled for Git.")
		default:
		}
		configureGitIgnore(cmd)
		proxy := utils.GetStringFlag(cmd, "proxy")
		if proxy != "" {
			command := utils.Format("git config --global http.proxy {proxy} && git config --global https.proxy {proxy}", map[string]string{
				"proxy": proxy,
			})
			utils.RunCmd(command)
		}
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		utils.RunCmd("git lfs uninstall")
		switch runtime.GOOS {
		case "darwin":
			utils.RunCmd("brew uninstall git git-lfs")
		case "linux":
			if utils.IsDebianSeries() {
				command := utils.Format("{prefix} apt-get purge {yes_s} git git-lfs", map[string]string{
					"prefix": utils.GetCommandPrefix(
						true,
						map[string]uint32{},
					),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} yum remove git", map[string]string{
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
	dstDir := utils.GetStringFlag(cmd, "dst-dir")
	dstFile := filepath.Join(dstDir, ".gitignore")
	if utils.GetBoolFlag(cmd, "append") {
		utils.AppendToTextFile(dstFile, utils.ReadEmbedFileAsString(srcFile))
		log.Printf("%s is appended into %s.", srcFile, dstFile)
	} else {
		utils.CopyEmbedFile(srcFile, dstFile, 0o600)
		log.Printf("%s is copied to %s.", srcFile, dstFile)
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
	GitCmd.Flags().String("proxy", "", "Configure Git to use the specified proxy.")
	GitCmd.Flags().StringP("dest-dir", "d", ".", "The destination directory (current directory, by default) to copy gitignore files to.")
	GitCmd.Flags().BoolP("append", "a", false, "Append to the .gitignore instead of oveerwriting it.")
	GitCmd.Flags().StringP("lang", "l", "", "The language to configure .gitignore for.")
	// rootCmd.AddCommand(gitCmd)
}
