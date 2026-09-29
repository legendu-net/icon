package cmd

import (
	"log"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/ai"
	"legendu.net/icon/cmd/bigdata"
	"legendu.net/icon/cmd/dev"
	"legendu.net/icon/cmd/filesystem"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/cmd/ide"
	"legendu.net/icon/cmd/jupyter"
	"legendu.net/icon/cmd/misc"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/cmd/shell"
	"legendu.net/icon/cmd/virtualization"
)

var rootCmd = &cobra.Command{
	Use:              "icon",
	Short:            "Install and configure tools.",
	TraverseChildren: true,
}

// commandConfigs are the functions registering all subcommands of icon.
var commandConfigs = []func(*cobra.Command){
	ai.ConfigPyTorchCmd,
	bigdata.ConfigArrowDBCmd,
	bigdata.ConfigSparkCmd,
	dev.ConfigBytehoundCmd,
	dev.ConfigGitCmd,
	dev.ConfigGolangCmd,
	dev.ConfigJjCmd,
	dev.ConfigPerfCmd,
	dev.ConfigPytypeCmd,
	dev.ConfigRustCmd,
	dev.ConfigDenoCmd,
	filesystem.ConfigRipCmd,
	filesystem.ConfigDropboxCmd,
	filesystem.ConfigFdCmd,
	filesystem.ConfigYaziCmd,
	icon.ConfigCompletionCmd,
	icon.ConfigDataCmd,
	icon.ConfigUpdateCmd,
	icon.ConfigVersionCmd,
	ide.ConfigFirenvimCmd,
	ide.ConfigNeovimCmd,
	ide.ConfigVscodeCmd,
	ide.ConfigHelixCmd,
	jupyter.ConfigGanymedeCmd,
	jupyter.ConfigIpythonCmd,
	jupyter.ConfigJupyterBookCmd,
	jupyter.ConfigJLabVimCmd,
	dev.ConfigHomebrewCmd,
	misc.ConfigGopassCmd,
	misc.ConfigKeepassXCCmd,
	misc.ConfigKeyboardCmd,
	network.ConfigDownloadGitHubReleaseCmd,
	network.ConfigSSHClientCmd,
	network.ConfigSSHServerCmd,
	shell.ConfigAlacrittyCmd,
	shell.ConfigAtuinCmd,
	shell.ConfigBashItCmd,
	shell.ConfigFishCmd,
	shell.ConfigGhosttyCmd,
	shell.ConfigNushellCmd,
	shell.ConfigTelevisionCmd,
	shell.ConfigWavetermCmd,
	shell.ConfigZellijCmd,
	virtualization.ConfigKVMCmd,
	virtualization.ConfigDockerCmd,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	switch runtime.GOOS {
	case "darwin", "linux":
	default:
		log.Fatal("The OS ", runtime.GOOS, " is not supported!")
	}

	for _, config := range commandConfigs {
		config(rootCmd)
	}

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
