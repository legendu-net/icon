package shell

import (
	"log"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
)

func downloadFishFromGitHub(version string) string {
	output := "/tmp/_fish_shell.tar.xz"
	network.DownloadGitHubRelease("fish-shell/fish-shell", version, map[string][]string{
		"common":             {"fish", "tar.xz"},
		"amd64":              {"x86_64"},
		"arm64":              {"aarch64"},
		"linux":              {"linux"},
		"DebianUbuntuSeries": {},
		"FedoraSeries":       {},
		"OtherLinux":         {},
	}, []string{"app", "zip", "pkg", "asc"}, output)
	return output
}

func generateCompletions() {
	dir := "~/.config/fish/completions/"
	var cmdMap map[string]string
	err := yaml.Unmarshal(utils.ReadFile(utils.NormalizePath(dir+"commands.yaml")), &cmdMap)
	if err != nil {
		log.Fatalf("Error unmarshaling data: %v", err)
	}

	for cmd, cmdCompletion := range cmdMap {
		if utils.ExistsCommand(cmd) {
			script := dir + cmd + ".fish"
			utils.RunCmd(cmdCompletion + " > " + script)
		}
	}
}

func generateCrazyCompletions() {
	var uvx string
	if utils.ExistsCommand("uvx") {
		uvx = "uvx"
	} else {
		file := utils.NormalizePath("~/.local/bin/uvx")
		if !utils.ExistsCommand(file) {
			utils.RunCmd("icon uv -ic")
		}
		uvx = file
	}
	dir := "~/.config/fish/completions/"
	dirCrazy := utils.NormalizePath(dir + "crazy_complete")
	if utils.ExistsPath(dirCrazy) {
		for _, entry := range utils.ReadDir(dirCrazy) {
			fileName := entry.Name()
			srcFile := filepath.Join(dirCrazy, fileName)
			fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName)) + ".fish"
			destFile := dir + fileName
			cmd := utils.Format(`{uvx} --with pyyaml \
				--from git+https://github.com/dclong/crazy-complete \
				crazy-complete --input-type=yaml fish {srcFile} > {destFile}`,
				map[string]string{
					"uvx":      uvx,
					"srcFile":  srcFile,
					"destFile": destFile,
				})
			utils.RunCmd(cmd)
		}
	}
}

// Install and config the fish shell.
func fish(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "linux":
			file := downloadFishFromGitHub(utils.GetStringFlag(cmd, "version"))
			command := utils.Format(`{prefix} tar --xz -xvf {file} -C /usr/bin/`, map[string]string{
				"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
				"file":   file,
			})
			utils.RunCmd(command)
		case "darwin":
			utils.RunCmd("brew install fish")
		default:
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		dir := "~/.config/fish"
		dir_go := utils.NormalizePath(dir)
		utils.BackupDir(dir_go, "")

		utils.MkdirAll(dir_go, 0o700)
		utils.RunCmd("git clone https://github.com/legendu-net/fish " + dir)

		generateCompletions()
		generateCrazyCompletions()
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		switch runtime.GOOS {
		case "linux":
			command := utils.Format(`{prefix} rm /usr/bin/fish`, map[string]string{
				"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
			})
			utils.RunCmd(command)
		case "darwin":
			utils.RunCmd("brew uninstall fish")
		default:
		}
	}
}

var FishCmd = &cobra.Command{
	Use:     "fish",
	Aliases: []string{},
	Short:   "Install and configure the fish shell.",
	Run:     fish,
}

func init() {
	FishCmd.Flags().BoolP("install", "i", false, "If specified, install the fish shell.")
	FishCmd.Flags().Bool("uninstall", false, "If specified, uninstall the fish shell.")
	FishCmd.Flags().BoolP("config", "c", false, "If specified, configure the fish shell.")
	FishCmd.Flags().StringP("version", "v", "", "The version of the release.")
}
