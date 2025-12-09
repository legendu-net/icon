package cmd

import (
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"legendu.net/icon/utils"
)

// Update icon.
func update(cmd *cobra.Command, _ []string) {
	dir := utils.GetStringFlag(cmd, "install-dir")
	command := utils.Format(`curl -sSL https://raw.githubusercontent.com/legendu-net/icon/main/install_icon.sh \
			| {prefix} bash -s -- -d {dir}`, map[string]string{
		"prefix": utils.GetCommandPrefix(false, map[string]uint32{
			dir: unix.W_OK | unix.R_OK,
		}),
		"dir": dir,
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
	updateCmd.Flags().StringP("install-dir", "d", "/usr/local/bin", "The directory for installing icon.")
	rootCmd.AddCommand(updateCmd)
}
