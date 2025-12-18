package jupyter

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/utils"
)

// Install and configure IPython.
func ipython(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		command := utils.Format("{prefix} {pip_install} ipython", map[string]string{
			"prefix": utils.GetCommandPrefix(
				utils.GetBoolFlag(cmd, "sudo"),
				map[string]uint32{},
			),
			"pip_install": utils.BuildPipInstall(cmd),
		})
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
		icon.FetchConfigData(false, "")
		profileDir := utils.GetStringFlag(cmd, "profile-dir")
		profileDefault := filepath.Join(profileDir, "profile_default")
		utils.Symlink(
			"~/.config/icon-data/ipython/startup.ipy",
			filepath.Join(profileDefault, "startup/startup.ipy"),
			!utils.GetBoolFlag(cmd, "no-backup"), utils.GetBoolFlag(cmd, "copy"))
		utils.SymlinkIntoDir(
			"~/.config/icon-data/ipython/ipython_config.py",
			profileDefault,
			!utils.GetBoolFlag(cmd, "no-backup"), utils.GetBoolFlag(cmd, "copy"))
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		command := utils.Format("{prefix} {pip_uninstall} ipython", map[string]string{
			"prefix": utils.GetCommandPrefix(
				utils.GetBoolFlag(cmd, "sudo"),
				map[string]uint32{},
			),
			"pip_uninstall": utils.BuildPipUninstall(cmd),
		})
		utils.RunCmd(command)
	}
}

var IpythonCmd = &cobra.Command{
	Use:     "ipython",
	Aliases: []string{"ipy"},
	Short:   "Install and configure IPython.",
	//Args:  cobra.ExactArgs(1),
	Run: ipython,
}

func init() {
	IpythonCmd.Flags().BoolP("install", "i", false, "Install IPython.")
	IpythonCmd.Flags().Bool("uninstall", false, "Uninstall IPython.")
	IpythonCmd.Flags().BoolP("config", "c", false, "Configure IPython.")
	IpythonCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	IpythonCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	IpythonCmd.Flags().Bool("sudo", false, "Force using sudo.")
	IpythonCmd.Flags().String("profile-dir", filepath.Join(utils.UserHomeDir(), ".ipython"), "The directory for storing IPython configuration files.")
	utils.AddPythonFlags(IpythonCmd)
}
