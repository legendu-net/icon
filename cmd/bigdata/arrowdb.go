package bigdata

import (
	//"embed"
	"path/filepath"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

func linkArrowDbProfileFromHost(backup, copy bool) {
	srcProfile := filepath.Join("/home_host", utils.GetCurrentUser().Name, ".arrowdb_profile")
	dstProfile := filepath.Join(utils.UserHomeDir(), ".arrowdb_profile")
	if utils.ExistsFile(srcProfile) {
		// inside a Docker container, link profile from host
		utils.Symlink(srcProfile, dstProfile, backup, copy)
	}
}

// Install and configure the Python library arrowdb.
func arrowDB(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		command := utils.Format("{prefix} {pip_install} arrowdb", map[string]string{
			"prefix": utils.GetCommandPrefix(
				utils.GetBoolFlag(cmd, "sudo"),
				map[string]uint32{},
			),
			"pip_install": utils.BuildPipInstall(cmd),
		})
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
		linkArrowDbProfileFromHost(!utils.GetBoolFlag(cmd, "no-backup"), utils.GetBoolFlag(cmd, "copy"))
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		command := utils.Format("{prefix} {pip_uninstall} arrowdb", map[string]string{
			"prefix": utils.GetCommandPrefix(
				utils.GetBoolFlag(cmd, "sudo"),
				map[string]uint32{},
			),
			"pip_uninstall": utils.BuildPipUninstall(cmd),
		})
		utils.RunCmd(command)
	}
}

var ArrowDBCmd = &cobra.Command{
	Use:     "arrowdb",
	Aliases: []string{},
	Short:   "Install and configure arrowdb.",
	//Args:  cobra.ExactArgs(1),
	Run: arrowDB,
}

func init() {
	ArrowDBCmd.Flags().BoolP("install", "i", false, "Install Spark.")
	ArrowDBCmd.Flags().BoolP("uninstall", "u", false, "Uninstall Spark.")
	ArrowDBCmd.Flags().BoolP("config", "c", false, "Configure Spark.")
	ArrowDBCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	ArrowDBCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	utils.AddPythonFlags(ArrowDBCmd)
}
