package cmd

import (
	"os"

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
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(network.DownloadGitHubReleaseCmd)
	rootCmd.AddCommand(network.SshClientCmd)
	rootCmd.AddCommand(network.SshServerCmd)
	rootCmd.AddCommand(jupyter.IpythonCmd)
	rootCmd.AddCommand(jupyter.JLabVimCmd)
	rootCmd.AddCommand(jupyter.JupyterBookCmd)
	rootCmd.AddCommand(jupyter.GanymedeCmd)
	rootCmd.AddCommand(shell.AlacrittyCmd)
	rootCmd.AddCommand(shell.AtuinCmd)
	rootCmd.AddCommand(shell.BashItCmd)
	rootCmd.AddCommand(shell.FishCmd)
	rootCmd.AddCommand(shell.HyperCmd)
	rootCmd.AddCommand(shell.NushellCmd)
	rootCmd.AddCommand(shell.ZellijCmd)
	rootCmd.AddCommand(dev.GitCmd)
	rootCmd.AddCommand(dev.GolangCmd)
	rootCmd.AddCommand(dev.RustCmd)
	rootCmd.AddCommand(dev.BytehoundCmd)
	rootCmd.AddCommand(dev.PytypeCmd)
	rootCmd.AddCommand(ide.NeovimCmd)
	rootCmd.AddCommand(ide.HelixCmd)
	rootCmd.AddCommand(ide.FirenvimCmd)
	rootCmd.AddCommand(ide.VscodeCmd)
	rootCmd.AddCommand(ai.PyTorchCmd)
	rootCmd.AddCommand(bigdata.SparkCmd)
	rootCmd.AddCommand(virtualization.DockerCmd)
	rootCmd.AddCommand(virtualization.LdcCmd)
	rootCmd.AddCommand(misc.KeepassXCCmd)
	rootCmd.AddCommand(misc.KeyboardCmd)
	rootCmd.AddCommand(filesystem.RipCmd)
	rootCmd.AddCommand(icon.CompletionCmd)
	rootCmd.AddCommand(icon.DataCmd)
	rootCmd.AddCommand(icon.UpdateCmd)
	rootCmd.AddCommand(icon.VersionCmd)
}
