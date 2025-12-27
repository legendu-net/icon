package dev

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install and configure Deno.
func deno(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		cmd := `curl -fsSL https://deno.land/install.sh | sh -s - -y --no-modify-path`
		utils.RunCmd(cmd)
	}
	if utils.GetBoolFlag(cmd, "config") {
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
	}
}

var denoCmd = &cobra.Command{
	Use:     "deno",
	Aliases: []string{},
	Short:   "Install and configure Deno.",
	//Args:  cobra.ExactArgs(1),
	Run: deno,
}

func ConfigDenoCmd(rootCmd *cobra.Command) {
	denoCmd.Flags().BoolP("install", "i", false, "Install Deno.")
	denoCmd.Flags().BoolP("config", "c", false, "Configure Deno.")
	denoCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	denoCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	denoCmd.Flags().BoolP("uninstall", "u", false, "Uninstall Deno.")
	rootCmd.AddCommand(denoCmd)
}
