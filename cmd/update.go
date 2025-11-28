package cmd

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Update icon.
func update(_ *cobra.Command, args []string) {
	command := utils.Format(`curl -sSL https://raw.githubusercontent.com/legendu-net/icon/main/install_icon.sh \
			| {prefix} bash -`, map[string]string{
		"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
	})
	utils.RunCmd(command)
}

var updateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"upd"},
	Short:   "Update icon.",
	Run:     update,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
