package network

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/utils"
)

const sshHome = "~/.ssh"

// Copy configuration files from /home_host/USER/.ssh if it exists.
// @param ssh_home: The home directory (~/.ssh) of SSH client configuration.
func copySshcSettingsFromHost() {
	sshSrc := filepath.Join("/home_host", utils.GetCurrentUser().Name, ".ssh")
	if utils.ExistsDir(sshSrc) {
		// inside a Docker container, use .ssh from host
		utils.RemoveAll(sshHome)
		utils.CopyDir(sshSrc, sshHome)
		adjustPathInConfig()
	}
}

func adjustPathInConfig() {
	path := filepath.Join(sshHome, "config")
	text := utils.ReadFileAsString(path)
	pattern := "IdentityFile=/Users/"
	if strings.Contains(text, pattern) {
		text = strings.ReplaceAll(text, pattern, "IdentityFile=/home/")
		utils.WriteTextFile(path, text, 0o600)
	}
}

// Install and configure SSH client.
func SSHClient(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
	}
	if utils.GetBoolFlag(cmd, "config") {
		icon.FetchConfigData(false, "")
		utils.SymlinkIntoDir("~/.config/icon-data/ssh/client/config", sshHome,
			!utils.GetBoolFlag(cmd, "no-backup"), utils.GetBoolFlag(cmd, "copy"))
		copySshcSettingsFromHost()
		utils.MkdirAll("~/.local/share/ssh", "700")
		utils.Chmod600(sshHome)
		log.Print("The permissions of ~/.ssh and its contents are correctly set.\n")
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
	}
}

var SSHClientCmd = &cobra.Command{
	Use:     "ssh_client",
	Aliases: []string{"sshc"},
	Short:   "Install and configure SSH client.",
	//Args:  cobra.ExactArgs(1),
	Run: SSHClient,
}

func init() {
	SSHClientCmd.Flags().BoolP("install", "i", false, "Install Git.")
	SSHClientCmd.Flags().Bool("uninstall", false, "Uninstall Git.")
	SSHClientCmd.Flags().BoolP("config", "c", false, "Configure Git.")
	SSHClientCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	SSHClientCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
}
