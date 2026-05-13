package network

import (
	"log"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/utils"
)

var sshHome = utils.NormalizePath("~/.ssh")

// Copy configuration files from /home_host/USER/.ssh if it exists.
// @param ssh_home: The home directory (~/.ssh) of SSH client configuration.
func copySshcSettingsFromHost() {
	//nolint:gocritic // filepathJoin: "/" is intentional to start an absolute path
	sshSrc := filepath.Join("/", "home_host", utils.GetCurrentUser().Name, ".ssh")
	if utils.ExistsDir(sshSrc) {
		// inside a Docker container, copy .ssh from host
		utils.CopyDirRegular(sshSrc, sshHome)
		adjustPathInConfig()
	}
}

func adjustPathInConfig() {
	path := filepath.Join(sshHome, "config")
	if !utils.ExistsFile(path) {
		return
	}
	text := utils.ReadFileAsString(path)
	replacements := [][2]string{
		{"IdentityFile=", "IdentityFile "},
		{"IdentityFile /Users/[^/]+/", "IdentityFile ~/"},
		{"IdentityFile /home/[^/]+/", "IdentityFile ~/"},
	}
	original := text
	for _, r := range replacements {
		text = regexp.MustCompile(r[0]).ReplaceAllString(text, r[1])
	}
	if text != original {
		//nolint:mnd // readable
		utils.WriteTextFile(path, text, 0o600)
	}
}

// Install and configure SSH client.
func SSHClient(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
	}
	if utils.GetBoolFlag(cmd, "config") {
		utils.BackupOrRemove(sshHome, utils.ShouldBackup(cmd))
		copySshcSettingsFromHost()
		icon.FetchConfigData(false, "")
		dst := filepath.Join(sshHome, "config")
		if !utils.ExistsFile(dst) {
			utils.CopyFile("~/.config/icon-data/ssh/client/config", dst)
		}
		utils.MkdirAll("~/.local/share/ssh", "700")
		utils.Chmod600(sshHome)
		log.Print("The permissions of ~/.ssh and its contents are correctly set.\n")
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
	}
}

var sshClientCmd = &cobra.Command{
	Use:     "ssh_client",
	Aliases: []string{"sshc"},
	Short:   "Install and configure SSH client.",
	//Args:  cobra.ExactArgs(1),
	Run: SSHClient,
}

func ConfigSSHClientCmd(rootCmd *cobra.Command) {
	sshClientCmd.Flags().BoolP("install", "i", false, "Install Git.")
	sshClientCmd.Flags().Bool("uninstall", false, "Uninstall Git.")
	sshClientCmd.Flags().BoolP("config", "c", false, "Configure Git.")
	sshClientCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	rootCmd.AddCommand(sshClientCmd)
}
