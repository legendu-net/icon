package dev

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
	"log"
	"path/filepath"
	"runtime"
)

// Install and configure pylint.
func pylint(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		command := utils.Format("{pip_install} pylint", map[string]string{
			"pip_install": utils.BuildPipInstall(cmd),
		})
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
        srcFile = "data/pylint/pyproject.toml"
        dictSrc = tomlkit.loads(src_file.read_text())
        des_file = args.dst_dir / "pyproject.toml"
        if des_file.is_file(){
            dictDes = tomlkit.loads(des_file.read_text())
		} else {
            dictDes = {}
		}
        update_dict(dic_des, dic_src, recursive=True)
        des_file.write_text(tomlkit.dumps(dic_des))
        logging.info("pylint is configured via %s.", des_file)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		run_cmd(f"{args.pip_uninstall} pylint")
	}
}

var PylintCmd = &cobra.Command{
	Use:     "pylint",
	Aliases: []string{},
	Short:   "Install and configure pylint.",
	//Args:  cobra.ExactArgs(1),
	Run: pylint,
}

func init() {
	PylintCmd.Flags().BoolP("install", "i", false, "Install Python Poetry.")
	PylintCmd.Flags().Bool("uninstall", false, "Uninstall Python Poetry.")
	PylintCmd.Flags().BoolP("config", "c", false, "Configure Python Poetry.")
	PylintCmd.Flags().StringP("dest-dir", "d", ".", "The destination directory to copy the pylint configuration file to.")
	// rootCmd.AddCommand(PylintCmd)
}
