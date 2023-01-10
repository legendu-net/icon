package dev

import (
	"github.com/elliotchance/orderedmap/v2"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
	"log"
	"path/filepath"
)

// Install and configure pytype.
func pytype(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		command := utils.Format("{pip_install} pytype", map[string]string{
			"pip_install": utils.BuildPipInstall(cmd),
		})
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
		srcFile := "data/pytype/pyproject.toml"
		var srcMap orderedmap.OrderedMap[string, any]
		toml.Unmarshal(utils.ReadEmbedFile(srcFile), &srcMap)
		destFile := filepath.Join(utils.GetStringFlag(cmd, "dest-dir"), "pyproject.toml")
		var destMap orderedmap.OrderedMap[string, any]
		if utils.ExistsFile(destFile) {
			toml.Unmarshal(utils.ReadFile(destFile), &destMap)
		}
		utils.UpdateMap(destMap, srcMap)
		bytes, err := toml.Marshal(destMap)
		if err != nil {
			log.Fatal("ERROR - ", err)
		}
		utils.WriteFile(destFile, bytes, 0o600)
		log.Printf("pytype is configured via %s.", destFile)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		command := utils.Format("{pip_uninstall} pytype", map[string]string{
			"pip_uninstall": utils.BuildPipUninstall(cmd),
		})
		utils.RunCmd(command)
	}
}

var PytypeCmd = &cobra.Command{
	Use:     "pytype",
	Aliases: []string{},
	Short:   "Install and configure pytype.",
	//Args:  cobra.ExactArgs(1),
	Run: pytype,
}

func init() {
	PytypeCmd.Flags().BoolP("install", "i", false, "Install Python Poetry.")
	PytypeCmd.Flags().Bool("uninstall", false, "Uninstall Python Poetry.")
	PytypeCmd.Flags().BoolP("config", "c", false, "Configure Python Poetry.")
	PytypeCmd.Flags().StringP("dest-dir", "d", ".", "The destination directory to copy the pylint configuration file to.")
	utils.AddPythonFlags(PytypeCmd)
}
