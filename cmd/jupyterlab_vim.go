package cmd

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install and configure the jupyterlab_vim extension for JupyterLab.
func jupyterlab_vim(cmd *cobra.Command, args []string) {
	prefix := utils.GetCommandPrefix(
		utils.GetBoolFlag(cmd, "sudo"),
		map[string]uint32{},
		"ls",
	)
	if utils.GetBoolFlag(cmd, "install") {
		command := utils.Format("{prefix} {pip_install} jupyterlab_vim", map[string]string{
			"prefix":      prefix,
			"pip_install": utils.BuildPipInstall(cmd),
		})
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") || utils.GetBoolFlag(cmd, "enable") || utils.GetBoolFlag(cmd, "disable") {
		if utils.GetBoolFlag(cmd, "enable") {
			command := utils.Format("{prefix} jupyter labextension enable @axlair/jupyterlab_vim", map[string]string{
				"prefix": prefix,
			})
			utils.RunCmd(command)
		}
		if utils.GetBoolFlag(cmd, "disable") {
			command := utils.Format("{prefix} jupyter labextension disable @axlair/jupyterlab_vim", map[string]string{
				"prefix": prefix,
			})
			utils.RunCmd(command)
		}
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		command := utils.Format("{prefix} {python} -m pip uninstall jupyterlab_vim", map[string]string{
			"prefix": prefix,
			"python": utils.GetStringFlag(cmd, "python"),
		})
		utils.RunCmd(command)
	}
}

var jvimCmd = &cobra.Command{
	Use:     "jupyterlab_vim",
	Aliases: []string{"jlab_vim", "jlabvim", "jvim"},
	Short:   "Install and configure the jupyterlab_vim extension for JupyterLab.",
	//Args:  cobra.ExactArgs(1),
	Run: jupyterlab_vim,
}

func init() {
	jvimCmd.Flags().BoolP("install", "i", false, "Install the jupyterlab_vim extension for JupyterLab.")
	jvimCmd.Flags().Bool("uninstall", false, "Uninstall the jupyterlab_vim extension for JupyterLab.")
	jvimCmd.Flags().BoolP("config", "c", false, "Configure the jupyterlab_vim extension for JupyterLab.")
	jvimCmd.Flags().Bool("sudo", false, "Force using sudo.")
	jvimCmd.Flags().Bool("enable", false, "Enable the jupyterlab_vim extension for JupyterLab.")
	jvimCmd.Flags().Bool("disable", false, "Disable the jupyterlab_vim extension for JupyterLab.")
	utils.AddPythonFlags(jvimCmd)
	rootCmd.AddCommand(jvimCmd)
}
