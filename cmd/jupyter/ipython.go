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
		backup := utils.ShouldBackup(cmd)
		doCopy := utils.GetBoolFlag(cmd, "copy")
		src1 := "~/.config/icon-data/ipython/startup.ipy"
		dst1 := filepath.Join(profileDefault, "startup", "startup.ipy")
		utils.BackupOrRemove(dst1, backup)
		utils.CopyOrSymlink(src1, dst1, doCopy)
		src2 := "~/.config/icon-data/ipython/ipython_config.py"
		dst2 := filepath.Join(profileDefault, filepath.Base(src2))
		utils.BackupOrRemove(dst2, backup)
		utils.CopyOrSymlink(src2, dst2, doCopy)
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

var ipythonCmd = &cobra.Command{
	Use:     "ipython",
	Aliases: []string{"ipy"},
	Short:   "Install and configure IPython.",
	//Args:  cobra.ExactArgs(1),
	Run: ipython,
}

func ConfigIpythonCmd(rootCmd *cobra.Command) {
	ipythonCmd.Flags().BoolP("install", "i", false, "Install IPython.")
	ipythonCmd.Flags().Bool("uninstall", false, "Uninstall IPython.")
	ipythonCmd.Flags().BoolP("config", "c", false, "Configure IPython.")
	ipythonCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	ipythonCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	ipythonCmd.Flags().Bool("sudo", false, "Force using sudo.")
	ipythonCmd.Flags().String("profile-dir", filepath.Join(utils.UserHomeDir(), ".ipython"),
		"The directory for storing IPython configuration files.")
	utils.AddPythonFlags(ipythonCmd)
	rootCmd.AddCommand(ipythonCmd)
}
