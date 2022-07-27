package cmd

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
	"path/filepath"
)

// Install and configure IPython.
func ipython(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		command := utils.Format("{prefix} {pip_install} ipython", map[string]string{
			"prefix": utils.GetCommandPrefix(
				utils.GetBoolFlag(cmd, "sudo"),
				map[string]uint32{},
				"ls",
			),
			"pip_install": utils.BuildPipInstall(cmd),
		})
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
		profile_dir := utils.GetStringFlag(cmd, "profile-dir")
		profile_default := filepath.Join(profile_dir, "profile_default")
		utils.CopyEmbedFile("data/ipython/startup.ipy", filepath.Join(profile_default, "startup/startup.ipy"), 0600)
		utils.CopyEmbedFileToDir("data/ipython/ipython_config.py", profile_default, 0600)
	}
	if utils.GetBoolFlag(cmd, "config") {
	}
}

var ipythonCmd = &cobra.Command{
	Use:     "ipython",
	Aliases: []string{"ipy"},
	Short:   "Install and configure IPython.",
	//Args:  cobra.ExactArgs(1),
	Run: ipython,
}

func init() {
	ipythonCmd.Flags().BoolP("install", "i", false, "Install IPython.")
	ipythonCmd.Flags().Bool("uninstall", false, "Uninstall IPython.")
	ipythonCmd.Flags().BoolP("config", "c", false, "Configure IPython.")
	ipythonCmd.Flags().Bool("sudo", false, "Force using sudo.")
	ipythonCmd.Flags().String("profile-dir", filepath.Join(utils.UserHomeDir(), ".ipython"), "The directory for storing IPython configuration files.")
	utils.AddPythonFlags(ipythonCmd)
	rootCmd.AddCommand(ipythonCmd)
}
