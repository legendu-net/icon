package dev

import (
	"log"

	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
)

// Install and configure Rust.
func bytehound(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		if utils.IsLinux() {
			network.DownloadGitHubRelease("koute/bytehound", "", map[string][]string{
				"common": {"bytehound", "tgz"},
				"amd64":  {"x86_64"},
				"linux":  {"linux", "gnu"},
			}, []string{}, "/tmp/bytehound.tar.gz")
			command := utils.Format(`mkdir -p ~/.local/bin && tar -zxvf /tmp/bytehound.tar.gz -C ~/.local/bin \
				&& mkdir -p ~/.local/lib && mv ~/.local/bin/libbytehound.so ~/.local/lib`, map[string]string{})
			utils.RunCmd(command)
			log.Println("libbytehound.so has been installed to ~/.local/lib.")
			log.Println("bytehound and bytehound-gather has been installed to ~/.local/bin.")
		} else {
			log.Fatalln("Bytehound is only supported on linux!")
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
	}
}

var BytehoundCmd = &cobra.Command{
	Use:     "bytehound",
	Aliases: []string{"bh", "byteh", "bhound"},
	Short:   "Install and configure Bytehound.",
	//Args:  cobra.ExactArgs(1),
	Run: bytehound,
}

func init() {
	BytehoundCmd.Flags().BoolP("install", "i", false, "Install Bytehound.")
	BytehoundCmd.Flags().BoolP("config", "c", false, "Configure Bytehound.")
	BytehoundCmd.Flags().BoolP("uninstall", "u", false, "Uninstall Bytehound.")
	BytehoundCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	BytehoundCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
}
