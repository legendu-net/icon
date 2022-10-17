package shell

import (
	"github.com/shirou/gopsutil/host"
	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
	"log"
	"path/filepath"
	"runtime"
)

func buildGithubCmd() *cobra.Command {
	var downloadGitHubReleaseCmd = &cobra.Command{}
	downloadGitHubReleaseCmd.Flags().StringP("repo", "r", "vercel/hyper", "")
	keywords := []string{"linux"}
	if utils.IsDebianSeries() {
		keywords = append(keywords, "deb")
	} else if utils.IsFedoraSeries() {
		keywords = append(keywords, "rpm")
	} else {
		keywords = append(keywords, "appimage")
	}
	info, err := host.Info()
	if err != nil {
		log.Fatal(err)
	}
	switch info.KernelArch {
	case "x86_64":
		keywords = append(keywords, "x64")
	case "arm":
		keywords = append(keywords, "arm64")
	default:
	}
	downloadGitHubReleaseCmd.Flags().StringSliceP("kwd", "k", keywords, "")
	downloadGitHubReleaseCmd.Flags().StringSliceP("KWD", "K", []string{}, "")
	downloadGitHubReleaseCmd.Flags().StringP("output", "o", "/tmp/_hyper_js_terminal", "")
	return downloadGitHubReleaseCmd
}

// Install and configure the Hyper terminal.
func hyper(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "linux":
			gitHubCmd := buildGithubCmd()
			if utils.IsDebianSeries() {
				network.DownloadGitHubReleaseArgs(gitHubCmd, args)
				command := utils.Format("{prefix} apt-get update && {prefix} apt-get install {yes_s} {path}", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yes_s":  utils.BuildYesFlag(cmd),
					"path":   utils.GetStringFlag(cmd, "output"),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				network.DownloadGitHubReleaseArgs(gitHubCmd, args)
				command := utils.Format("{prefix} yum install {yes_s} {path}", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yes_s":  utils.BuildYesFlag(cmd),
					"path":   utils.GetStringFlag(cmd, "output"),
				})
				utils.RunCmd(command)
			}
		case "darwin":
			utils.RunCmd("brew install --cask hyper")
		default:
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		utils.RunCmd("hyper i hypercwd")
		utils.RunCmd("hyper i hyper-search")
		utils.RunCmd("hyper i hyper-pane")
		utils.RunCmd("hyper i hyperpower")
		log.Printf("Hyper plugins hypercwd, hyper-search, hyper-pane and hyperpower are installed.\n")
		utils.CopyEmbedFile("data/hyper/hyper.js", filepath.Join(utils.UserHomeDir(), ".hyper.js"), 0o600)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		switch runtime.GOOS {
		case "linux":
		case "darwin":
			utils.RunCmd("brew uninstall --cask hyper")
		default:
		}
	}
}

var HyperCmd = &cobra.Command{
	Use:     "hyper",
	Aliases: []string{},
	Short:   "Install and configure the Hyper terminal.",
	//Args:  cobra.ExactArgs(1),
	Run: hyper,
}

func init() {
	HyperCmd.Flags().BoolP("install", "i", false, "Install the Hyper terminal.")
	HyperCmd.Flags().Bool("uninstall", false, "Uninstall Hyper terminal.")
	HyperCmd.Flags().BoolP("config", "c", false, "Configure the Hyper terminal.")
	HyperCmd.Flags().StringP("repo", "r", "", "A GitHub repo of the form 'user_name/repo_name'.")
	err := HyperCmd.MarkFlagRequired("repo")
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	HyperCmd.Flags().StringP("version", "v", "", "The version of the release.")
	HyperCmd.Flags().StringSliceP("kwd", "k", []string{}, "Keywords that the assert's name contains.")
	HyperCmd.Flags().StringSliceP("KWD", "K", []string{}, "Keywords that the assert's name contains.")
	err = HyperCmd.MarkFlagRequired("kwd")
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	HyperCmd.Flags().StringP("output", "o", "", "The output path for the downloaded asset.")
	err = HyperCmd.MarkFlagRequired("output")
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	// rootCmd.AddCommand(HyperCmd)
}
