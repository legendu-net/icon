package bigdata

import (
	//"embed"
	"path/filepath"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

func linkArrowDbProfileFromHost() {
	srcProfile := filepath.Join("/home_host", utils.GetCurrentUser().Name, ".arrowdb_profile")
	dstProfile := filepath.Join(utils.UserHomeDir(), ".arrowdb_profile")
	if utils.ExistsFile(srcProfile) {
		// inside a Docker container, link profile from host
		utils.Symlink(srcProfile, dstProfile, false)
	}
}

// Install and configure the Python library arrowdb.
func arrowDb(cmd *cobra.Command, _ []string) {
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
		linkArrowDbProfileFromHost()
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

var ArrowDbCmd = &cobra.Command{
	Use:     "arrowdb",
	Aliases: []string{},
	Short:   "Install and configure arrowdb.",
	//Args:  cobra.ExactArgs(1),
	Run: arrowDb,
}

func init() {
	ArrowDbCmd.Flags().BoolP("install", "i", false, "Install Spark.")
	ArrowDbCmd.Flags().BoolP("uninstall", "u", false, "Uninstall Spark.")
	ArrowDbCmd.Flags().BoolP("config", "c", false, "Configure Spark.")
	utils.AddPythonFlags(ArrowDbCmd)
	// rootCmd.AddCommand(ArrowDbCmd)
}
