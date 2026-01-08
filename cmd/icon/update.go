package icon

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"legendu.net/icon/utils"
)

// Update icon.
func update(cmd *cobra.Command, _ []string) {
	dir := utils.GetStringFlag(cmd, "install-dir")
	if dir == "" {
		dir = filepath.Dir(utils.LookPath("icon"))
	}
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

func ConfigUpdateCmd(rootCmd *cobra.Command) {
	updateCmd.Flags().StringP("install-dir", "d", "", "The directory for installing icon.")
	rootCmd.AddCommand(updateCmd)
}
