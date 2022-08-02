package virtualization

import (
	"log"
	"runtime"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install and configure Docker container.
func docker(cmd *cobra.Command, args []string) {
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "linux":
			if utils.IsDebianSeries() {
				command := utils.Format("{prefix} apt-get update && {prefix} apt-get install {yes_s} docker.io docker-compose", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yes_s":  utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} yum install {yes_s} docker docker-compose", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yes_s":  utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
		case "darwin":
			utils.BrewInstallSafe([]string{"docker", "docker-compose", "bash-completion@2"})
		default:
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		userToDocker := utils.GetStringFlag(cmd, "user-to-docker")
		if userToDocker != "" {
			switch runtime.GOOS {
			case "linux":
				if utils.IsDebianSeries() {
					command := utils.Format("{prefix} gpasswd -a {user_to_docker} docker", map[string]string{
						"prefix":       utils.GetCommandPrefix(true, map[string]uint32{}),
						"userToDocker": userToDocker,
					})
					utils.RunCmd(command)
					log.Printf("Please run the command 'newgrp docker' or logout/login to make the group 'docker' effective!\n")
				}
			case "darwin":
				command := utils.Format("{prefix} dseditgroup -o edit -a {userToDocker} -t user staff", map[string]string{
					"prefix":       utils.GetCommandPrefix(true, map[string]uint32{}),
					"userToDocker": userToDocker,
				})
				utils.RunCmd(command)
			default:
			}
		}
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		switch runtime.GOOS {
		case "linux":
			if utils.IsDebianSeries() {
				command := utils.Format("{prefix} apt-get purge {yes_s} docker docker-compose", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yes_s":  utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			} else if utils.IsFedoraSeries() {
				command := utils.Format("{prefix} yum remove {yes_s} docker docker-compose", map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
					"yes_s":  utils.BuildYesFlag(cmd),
				})
				utils.RunCmd(command)
			}
		case "darwin":
			utils.RunCmd(
				"brew uninstall docker docker-completion docker-compose docker-compose-completion",
			)
		default:
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
	DockerCmd.Flags().BoolP("uninstall", "u", false, "Uninstall Rust.")
	user := utils.GetCurrentUser().Username
	DockerCmd.Flags().String("user-to-docker", utils.IfElseString(user == "root", "", user), "Add the specified user to the docker group.")
	// rootCmd.AddCommand(RustCmd)
}
