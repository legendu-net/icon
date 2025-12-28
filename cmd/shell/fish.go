package shell

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/utils"
)

func generateCompletions() {
	dir := "~/.config/fish/completions/"
	var cmdMap map[string]string
	err := yaml.Unmarshal(utils.ReadFile(dir+"commands.yaml"), &cmdMap)
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
		file := "~/.local/bin/uvx"
		if !utils.ExistsCommand(file) {
			utils.RunCmd("curl -LsSf https://astral.sh/uv/install.sh | sh")
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
			cmd := utils.Format(`{uvx} --python '>=3.10' --with pyyaml \
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
func fish(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		if utils.IsLinux() {
			if utils.IsDebianUbuntuSeries() {
				if utils.IsUbuntuSeries() {
					cmd := utils.Format("{prefix} add-apt-repository ppa:fish-shell/release-4", map[string]string{
						"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					})
					utils.RunCmd(cmd)
				}
				cmd := utils.Format(`{prefix} apt-get {yesStr} update \
				&& {prefix} apt-get {yesStr} install fish`, map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(cmd)
			} else if utils.IsFedoraSeries() {
				cmd := utils.Format(`{prefix} dnf {yesStr} install fish`, map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(cmd)
			}
			log.Printf("Successfully installed the fish shell.\n")
		} else {
			utils.RunCmd("brew install fish")
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		icon.FetchConfigData(false, "")

		dir := "~/.config/fish"
		utils.Symlink("~/.config/icon-data/fish", dir,
			!utils.GetBoolFlag(cmd, "no-backup"), utils.GetBoolFlag(cmd, "copy"))

		generateCompletions()
		generateCrazyCompletions()
	}
	if utils.GetBoolFlag(cmd, "trust") {
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
	}
}

var fishCmd = &cobra.Command{
	Use:     "fish",
	Aliases: []string{},
	Short:   "Install and configure the fish shell.",
	Run:     fish,
}

func ConfigFishCmd(rootCmd *cobra.Command) {
	fishCmd.Flags().BoolP("install", "i", false, "If specified, install the fish shell.")
	fishCmd.Flags().Bool("uninstall", false, "If specified, uninstall the fish shell.")
	fishCmd.Flags().BoolP("config", "c", false, "If specified, configure the fish shell.")
	fishCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	fishCmd.Flags().BoolP("trust", "t", false, "Add the fish shell into /etc/shells.")
	fishCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	fishCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	fishCmd.Flags().StringP("version", "v", "", "The version of the release.")
	rootCmd.AddCommand(fishCmd)
}
