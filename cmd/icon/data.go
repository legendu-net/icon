package icon

import (
	"fmt"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

const GIT_URL = "https://github.com/legendu-net/icon-data.git"

func FetchConfigData(force bool, gitUrl string) {
	if gitUrl == "" {
		gitUrl = GIT_URL
	}

	dir := "~/.config/icon-data"
	if !force && utils.ExistsDir(dir+"/.git") {
		fmt.Printf("Using existing data in %s.\n", dir)
		return
	}

	utils.Backup(dir, "")
	utils.MkdirAll(dir, 0o700)

	command := utils.Format(`git clone {gitUrl} {dir} \
			&& cd {dir} && git submodule init && git submodule update --remote`, map[string]string{
		"gitUrl": gitUrl,
		"dir":    dir,
	})
	utils.RunCmd(command)
	fmt.Printf("Data for icon has been pulled into %s.\n", dir)
}

// Pull data for icon from GitHub into ~/.config/icon-data.
func data(cmd *cobra.Command, _ []string) {
	FetchConfigData(utils.GetBoolFlag(cmd, "force"), utils.GetStringFlag(cmd, "git-url"))
}

var DataCmd = &cobra.Command{
	Use:     "data",
	Aliases: []string{"d"},
	Short:   "Pull data for icon from GitHub into ~/.config/icon-data.",
	Run:     data,
}

func init() {
	DataCmd.Flags().StringP("git-url", "g", GIT_URL, "The Git repo URL for icon-data.")
	DataCmd.Flags().Bool("force", false, "Force pulling data if it alreay exists.")
}
