package dev

import (
	"github.com/elliotchance/orderedmap/v2"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
	"log"
	"path/filepath"
)

// Update map1 using map2.
func UpdateMap(map1 orderedmap.OrderedMap[string, any], map2 orderedmap.OrderedMap[string, any]) {
	for _, key2 := range map2.Keys() {
		val2, _ := map2.Get(key2)
		val1, map1HasKey2 := map1.Get(key2)
		if !map1HasKey2 {
			map1.Set(key2, val2)
			continue
		}
		switch val2.(type) {
		case orderedmap.OrderedMap[string, any]:
			switch val1.(type) {
			case orderedmap.OrderedMap[string, any]:
				UpdateMap(val1.(orderedmap.OrderedMap[string, any]), val2.(orderedmap.OrderedMap[string, any]))
			default:
				map1.Set(key2, val2)
			}
		default:
			map1.Set(key2, val2)
		}
	}
}

// Install and configure pylint.
func pylint(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		command := utils.Format("{pip_install} pylint", map[string]string{
			"pip_install": utils.BuildPipInstall(cmd),
		})
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
		srcFile := "data/pylint/pyproject.toml"
		var srcMap orderedmap.OrderedMap[string, any]
		toml.Unmarshal(utils.ReadEmbedFile(srcFile), &srcMap)
		destFile := filepath.Join(utils.GetStringFlag(cmd, "dest-dir"), "pyproject.toml")
		var destMap orderedmap.OrderedMap[string, any]
		if utils.ExistsFile(destFile) {
			toml.Unmarshal(utils.ReadFile(destFile), &destMap)
		}
		UpdateMap(destMap, srcMap)
		bytes, err := toml.Marshal(destMap)
		if err != nil {
			log.Fatal("ERROR - ", err)
		}
		utils.WriteFile(destFile, bytes, 0o600)
		log.Printf("pylint is configured via %s.", destFile)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		command := utils.Format("{pip_uninstall} pylint", map[string]string{
			"pip_uninstall": utils.BuildPipUninstall(cmd),
		})
		utils.RunCmd(command)
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
	utils.AddPythonFlags(PylintCmd)
	// rootCmd.AddCommand(PylintCmd)
}
