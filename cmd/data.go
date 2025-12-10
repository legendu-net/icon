package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Pull data for icon from GitHub into ~/.config/icon-data.
func data(cmd *cobra.Command, _ []string) {
	dir := "~/.config/icon-data"
	if !utils.GetBoolFlag(cmd, "force") && utils.ExistsDir(dir+"/.git") {
		fmt.Println("Using existing data in ~/.config/icon-data.")
		return
	}

	utils.Backup(dir, "")
	utils.MkdirAll(dir, 0o700)

	command := utils.Format(`git clone {gitUrl} {dir} \
			&& cd {dir} && git submodule init && git submodule update --remote`, map[string]string{
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
	dataCmd.Flags().StringP("git-url", "g", "https://github.com/legendu-net/icon-data.git", "The Git repo URL for icon-data.")
	dataCmd.Flags().Bool("force", false, "Force pulling data if it alreay exists.")
	rootCmd.AddCommand(dataCmd)
}
