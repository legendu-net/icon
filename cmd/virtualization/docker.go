package virtualization

import (
	"log"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install and configure Docker container.
func docker(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		if utils.IsLinux() {
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format("{prefix} apt-get update && {prefix} apt-get install {yesStr} docker.io docker-compose", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yesStr} install docker docker-compose", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
			command := utils.Format("{prefix} chown root:docker /var/run/docker.sock", map[string]string{
				"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
			})
			utils.RunCmd(command)
		} else {
			utils.BrewInstallSafe([]string{"docker", "docker-compose", "bash-completion@2"})
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		userToDocker := utils.GetStringFlag(cmd, "user-to-docker")
		if userToDocker != "" {
			if utils.IsLinux() {
				if utils.IsDebianUbuntuSeries() {
					command := utils.Format("{prefix} gpasswd -a {user_to_docker} docker", map[string]string{
						"prefix":         utils.GetCommandPrefix(true, map[string]uint32{}),
						"user_to_docker": userToDocker,
					})
					utils.RunCmd(command)
					log.Printf("Please run the command 'newgrp docker' or logout/login to make the group 'docker' effective!\n")
				}
			} else {
				command := utils.Format("{prefix} dseditgroup -o edit -a {user_to_docker} -t user staff", map[string]string{
					"prefix":         utils.GetCommandPrefix(true, map[string]uint32{}),
					"user_to_docker": userToDocker,
				})
				utils.RunCmd(command)
			}
		}
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		if utils.IsLinux() {
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format("{prefix} apt-get purge {yesStr} docker docker-compose", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yesStr} remove docker docker-compose", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yesStr": utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
		} else {
			utils.RunCmd(
				"brew uninstall docker docker-completion docker-compose docker-compose-completion",
			)
		}
	}
}

var dockerCmd = &cobra.Command{
	Use:     "docker",
	Aliases: []string{},
	Short:   "Install and configure Docker.",
	//Args:  cobra.ExactArgs(1),
	Run: docker,
}

func ConfigDockerCmd(rootCmd *cobra.Command) {
	dockerCmd.Flags().BoolP("install", "i", false, "Install Rust.")
	dockerCmd.Flags().BoolP("config", "c", false, "Configure Rust.")
	dockerCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	dockerCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	dockerCmd.Flags().BoolP("uninstall", "u", false, "Uninstall Rust.")
	dockerCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	user := utils.GetCurrentUser().Username
	dockerCmd.Flags().String("user-to-docker", utils.IfElseString(user == "root", "", user), "Add the specified user to the docker group.")
	rootCmd.AddCommand(dockerCmd)
}
