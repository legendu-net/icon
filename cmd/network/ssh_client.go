package network

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

var sshHome = filepath.Join(utils.UserHomeDir(), ".ssh")

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
func SshClient(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
	}
	if utils.GetBoolFlag(cmd, "config") {
		utils.SymlinkIntoDir("~/.config/icon-data/ssh/client/config", sshHome, true)
		copySshcSettingsFromHost()
		utils.MkdirAll(filepath.Join(sshHome, "control"), 0o700)
		/*
					switch runtime.GOOS {
					case "linux", "darwin":
			            command = utils.Format("{prefix} chown -R {USER}:`id -g {USER}` {HOME}/.ssh", map[string]string{
						})
			            utils.RunCmd(command)
					default:
					}
		*/
		utils.Chmod600(sshHome)
		log.Print("The permissions of ~/.ssh and its contents are correctly set.\n")
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
	}
}

var SshClientCmd = &cobra.Command{
	Use:     "ssh_client",
	Aliases: []string{"sshc"},
	Short:   "Install and configure SSH client.",
	//Args:  cobra.ExactArgs(1),
	Run: SshClient,
}

func init() {
	SshClientCmd.Flags().BoolP("install", "i", false, "Install Git.")
	SshClientCmd.Flags().Bool("uninstall", false, "Uninstall Git.")
	SshClientCmd.Flags().BoolP("config", "c", false, "Configure Git.")
	// rootCmd.AddCommand(sshClientCmd)
}
