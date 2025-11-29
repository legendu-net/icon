package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Pull data for icon from GitHub into ~/.config/icon-data.
func data(cmd *cobra.Command, args []string) {
	dir := "~/.config/icon-data"
	utils.BackupDir(dir, "")
	utils.MkdirAll(dir, 0o700)

	command := utils.Format(`git clone {gitUrl} {dir} \
			&& cd {dir} && git submodule init && git submodule update`, map[string]string{
		"gitUrl": utils.GetStringFlag(cmd, "git-url"),
		"dir":    dir,
	})
	utils.RunCmd(command)
	fmt.Printf("Data for icon has been pulled into %s.\n", dir)
}

var dataCmd = &cobra.Command{
	Use:     "data",
	Aliases: []string{"d"},
	Short:   "Pull data for icon from GitHub into ~/.config/icon-data.",
	Run:     data,
}

func init() {
	dataCmd.Flags().StringP("git-url", "g", "git@github.com:legendu-net/icon-data.git", "The Git repo URL for icon-data.")
	rootCmd.AddCommand(dataCmd)
}
