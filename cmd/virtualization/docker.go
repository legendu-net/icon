package virtualization

import (
	"log"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install and configure Docker container.
func docker(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		if utils.IsLinux() {
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format("{prefix} apt-get update && {prefix} apt-get install {yesStr} docker.io docker-compose", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yesStr":  utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yesStr} install docker docker-compose", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yesStr":  utils.BuildYesFlag(cmd),
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
					"yesStr":  utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} dnf {yesStr} remove docker docker-compose", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yesStr":  utils.BuildYesFlag(cmd),
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

var DockerCmd = &cobra.Command{
	Use:     "docker",
	Aliases: []string{},
	Short:   "Install and configure Docker.",
	//Args:  cobra.ExactArgs(1),
	Run: docker,
}

func init() {
	DockerCmd.Flags().BoolP("install", "i", false, "Install Rust.")
	DockerCmd.Flags().BoolP("config", "c", false, "Configure Rust.")
	DockerCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	DockerCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	DockerCmd.Flags().BoolP("uninstall", "u", false, "Uninstall Rust.")
	DockerCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	user := utils.GetCurrentUser().Username
	DockerCmd.Flags().String("user-to-docker", utils.IfElseString(user == "root", "", user), "Add the specified user to the docker group.")
}
