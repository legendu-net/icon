package cmd

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/ai"
	"legendu.net/icon/cmd/bigdata"
	"legendu.net/icon/cmd/dev"
	"legendu.net/icon/cmd/filesystem"
	"legendu.net/icon/cmd/ide"
	"legendu.net/icon/cmd/jupyter"
	"legendu.net/icon/cmd/misc"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/cmd/shell"
	"legendu.net/icon/cmd/virtualization"
	"os"
)

var rootCmd = &cobra.Command{
	Use:              "icon",
	Short:            "Install and configure tools.",
	TraverseChildren: true,
	/*
		Run: func(cmd *cobra.Command, args []string) {
		},
	*/
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
	rootCmd.AddCommand(shell.HyperCmd)
	rootCmd.AddCommand(shell.ZoxideCmd)
	rootCmd.AddCommand(dev.GitCmd)
	rootCmd.AddCommand(dev.GolangCmd)
	rootCmd.AddCommand(dev.RustCmd)
	rootCmd.AddCommand(dev.BytehoundCmd)
	rootCmd.AddCommand(dev.PoetryCmd)
	rootCmd.AddCommand(dev.PylintCmd)
	rootCmd.AddCommand(dev.PytypeCmd)
	rootCmd.AddCommand(ide.NeovimCmd)
	rootCmd.AddCommand(ide.HelixCmd)
	rootCmd.AddCommand(ide.SpaceVimCmd)
	rootCmd.AddCommand(ide.FirenvimCmd)
	rootCmd.AddCommand(ide.VscodeCmd)
	rootCmd.AddCommand(ai.PyTorchCmd)
	rootCmd.AddCommand(bigdata.SparkCmd)
	rootCmd.AddCommand(virtualization.DockerCmd)
	rootCmd.AddCommand(misc.KeepassXCCmd)
	rootCmd.AddCommand(filesystem.RipCmd)
}
