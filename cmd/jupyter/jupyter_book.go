package jupyter

import (
	"log"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install and configure jupyter_book.
func jupyter_book(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		command := utils.Format("{pip_install} jupyter-book", map[string]string{
			"pip_install": utils.BuildPipInstall(cmd),
		})
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
		srcFile := "jupyter-book/_config.yml"
		utils.CopyEmbedFileToDir(srcFile, ".", 0o600, true)
		log.Printf("%s is copied to the current directory.", srcFile)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		command := utils.Format("{pip_uninstall} jupyter_book", map[string]string{
			"pip_uninstall": utils.BuildPipUninstall(cmd),
		})
		utils.RunCmd(command)
	}
}

var JupyterBookCmd = &cobra.Command{
	Use:     "jupyter_book",
	Aliases: []string{"jb", "jbook"},
	Short:   "Install and configure jupyter_book.",
	//Args:  cobra.ExactArgs(1),
	Run: jupyter_book,
}

func init() {
	JupyterBookCmd.Flags().BoolP("install", "i", false, "Install jupyter_book.")
	JupyterBookCmd.Flags().Bool("uninstall", false, "Uninstall jupyter_book.")
	JupyterBookCmd.Flags().BoolP("config", "c", false, "Configure jupyter_book.")
	utils.AddPythonFlags(JupyterBookCmd)
	// rootCmd.AddCommand(jupyterBookCmd)
}
