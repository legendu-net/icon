package ide

import (
	"log"

	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
)

// Install and configure Firenvim.
func firenvim(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		Neovim(true, true, false, "-y",
			!utils.GetBoolFlag(cmd, "no-backup"), utils.GetBoolFlag(cmd, "copy"))
		network.InstallChromeExtension("egpjdkipkomnmjhjmdamaniclmdlobbo", "Firenvim")
		utils.RunCmd(`nvim --headless +"call firenvim#install(0)" +qall`)
		if utils.IsLinux() {
			url := "https://www.legendu.net/misc/blog/firenvim-brings-neovim-into-your-browser/#installation"
			log.Printf("\nPlease follow step 5 in %s to configure a shortcut!\n", url)
		} else {
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
	}
}

var FirenvimCmd = &cobra.Command{
	Use:     "firenvim",
	Aliases: []string{"fvim"},
	Short:   "Install and configure Firenvim.",
	//Args:  cobra.ExactArgs(1),
	Run: firenvim,
}

func init() {
	FirenvimCmd.Flags().BoolP("install", "i", false, "Install Firenvim.")
	FirenvimCmd.Flags().Bool("uninstall", false, "Uninstall Firenvim.")
	FirenvimCmd.Flags().BoolP("config", "c", false, "Configure Firenvim.")
	FirenvimCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	FirenvimCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
}
