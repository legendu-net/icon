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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	switch runtime.GOOS {
	case "darwin", "linux":
	default:
		log.Fatal("The OS ", runtime.GOOS, " is not supported!")
	}

	ai.ConfigPyTorchCmd(rootCmd)
	bigdata.ConfigArrowDBCmd(rootCmd)
	bigdata.ConfigSparkCmd(rootCmd)
	dev.ConfigBytehoundCmd(rootCmd)
	dev.ConfigGitCmd(rootCmd)
	dev.ConfigGolangCmd(rootCmd)
	dev.ConfigPerfCmd(rootCmd)
	dev.ConfigPytypeCmd(rootCmd)
	dev.ConfigRustCmd(rootCmd)
	dev.ConfigDenoCmd(rootCmd)
	filesystem.ConfigRipCmd(rootCmd)
	icon.ConfigCompletionCmd(rootCmd)
	icon.ConfigDataCmd(rootCmd)
	icon.ConfigUpdateCmd(rootCmd)
	icon.ConfigVersionCmd(rootCmd)
	ide.ConfigFirenvimCmd(rootCmd)
	ide.ConfigNeovimCmd(rootCmd)
	ide.ConfigVscodeCmd(rootCmd)
	ide.ConfigHelixCmd(rootCmd)
	jupyter.ConfigGanymedeCmd(rootCmd)
	jupyter.ConfigIpythonCmd(rootCmd)
	jupyter.ConfigJupyterBookCmd(rootCmd)
	jupyter.ConfigJLabVimCmd(rootCmd)
	misc.ConfigHomebrewCmd(rootCmd)
	misc.ConfigKeepassXCCmd(rootCmd)
	misc.ConfigKeyboardCmd(rootCmd)
	network.ConfigDownloadGitHubReleaseCmd(rootCmd)
	network.ConfigSSHClientCmd(rootCmd)
	network.ConfigSSHServerCmd(rootCmd)
	shell.ConfigAlacrittyCmd(rootCmd)
	shell.ConfigAtuinCmd(rootCmd)
	shell.ConfigBashItCmd(rootCmd)
	shell.ConfigFishCmd(rootCmd)
	shell.ConfigHyperCmd(rootCmd)
	shell.ConfigNushellCmd(rootCmd)
	shell.ConfigZellijCmd(rootCmd)
	virtualization.ConfigKVMCmd(rootCmd)
	virtualization.ConfigDockerCmd(rootCmd)
	virtualization.ConfigLdcCmd(rootCmd)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
