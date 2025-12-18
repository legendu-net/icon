package icon

import (
	"fmt"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

const GitURL = "https://github.com/legendu-net/icon-data.git"

func FetchConfigData(force bool, gitURL string) {
	if gitURL == "" {
		gitURL = GitURL
	}

	dir := "~/.config/icon-data"
	if !force && utils.ExistsDir(dir+"/.git") {
		fmt.Printf("Using existing data in %s.\n", dir)
		return
	}

	utils.Backup(dir, "")
	utils.MkdirAll(dir, "700")

	command := utils.Format(`git clone {gitUrl} {dir} \
			&& cd {dir} && git submodule init && git submodule update --remote`, map[string]string{
		"gitUrl": gitURL,
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
	DataCmd.Flags().StringP("git-url", "g", GitURL, "The Git repo URL for icon-data.")
	DataCmd.Flags().Bool("force", false, "Force pulling data if it alreay exists.")
}
