package cmd

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

func getDockerImagePort(imageName string) int {
	if strings.HasPrefix(imageName, "dclong/") {
		imageName = imageName[7:]
		ports := map[string]int{
			"jupyterlab": 8888,
			"jupyterhub": 8000,
			"gitpod":     8000,
			"vscode":     8080,
		}
		for prefix, port := range ports {
			if strings.HasPrefix(imageName, prefix) {
				return port
			}
		}
	}
	return 0
}

func appendDockerImagePort(command *[]string, cmd *cobra.Command, imageName string) {
	port := getDockerImagePort(imageName)
	if port > 0 {
		portHost := utils.GetIntFlag(cmd, "port")
		if portHost == 0 {
			portHost = port
		}
		*command = append(*command, fmt.Sprintf("--publish=%d:%d", portHost, port))
	}
	extraPortMappings := utils.GetStringSliceFlag(cmd, "extra-port-mappings")
	if len(extraPortMappings) > 0 {
		for _, m := range extraPortMappings {
			*command = append(*command, "--publish="+m)
		}
	}
}

func getDockerImageCommand(imageName string) string {
	dockerCommands := map[string]string{
		"dclong/vscode-server": "/scripts/sys/init.sh --switch-user",
		"dclong/jupyterlab":    "/scripts/sys/init.sh --switch-user",
	}
	for prefix, dockerCommand := range dockerCommands {
		if strings.HasPrefix(imageName, prefix) {
			return dockerCommand
		}
	}
	if strings.HasPrefix(imageName, "dclong/") {
		return "/scripts/sys/init.sh"
	}
	return ""
}

func appendDockerImageCommand(command *[]string, args *[]string) {
	*command = append(*command, *args...)
	if len(*args) == 1 {
		*command = append(*command, getDockerImageCommand((*args)[0]))
	}
}

func getDockerImageHostname(imageName string) string {
	start := strings.Index(imageName, "/") + 1
	end := strings.Index(imageName, ":")
	if end < 0 {
		end = len(imageName)
	}
	return imageName[start:end]
}

// Launch a Docker container.
func ldc(cmd *cobra.Command, args []string) {
	currentUser := utils.GetCurrentUser()
	userName := utils.GetStringFlag(cmd, "user")
	if userName == "" {
		userName = currentUser.Username
	}
	password := utils.GetStringFlag(cmd, "password")
	if password == "" {
		password = userName
	}
	userId := currentUser.Uid
	groupId := currentUser.Gid
	command := []string{
		"docker",
		"run",
		"-it",
		"--init",
		"--privileged",
		"--cap-add",
		"SYS_ADMIN",
		"--platform",
		"linux/amd64",
		"--log-opt",
		"max-size=50m",
		"-e",
		"DOCKER_USER=" + userName,
		"-e",
		"DOCKER_USER_ID=" + userId,
		"-e",
		"DOCKER_PASSWORD=" + password,
		"-e",
		"DOCKER_GROUP_ID=" + groupId,
		"-e",
		"DOCKER_ADMIN_USER=" + userName,
		"--hostname",
		getDockerImageHostname(args[0]),
	}
	cwd := utils.Getwd()
	command = append(command, "-v", cwd+":/workdir")
	home := utils.UserHomeDir()
	if utils.GetBoolFlag(cmd, "mount-home") {
		command = append(command, "-v", filepath.Dir(home)+":/home_host")
	}
	detach := utils.GetBoolFlag(cmd, "detach")
	if detach {
		command[2] = "-d"
	}
	if runtime.GOOS == "linux" {
		memStat := utils.VirtualMemory()
		memory := int(0.8 * float64(memStat.Total))
		command = append(command, "--memory="+strconv.Itoa(memory)+"b")
		cpuInfo := utils.CpuInfo()
		cpus := utils.Max(len(cpuInfo)-1, 1)
		command = append(command, "--cpus="+strconv.Itoa(cpus))
	}
	appendDockerImagePort(&command, cmd, args[0])
	appendDockerImageCommand(&command, &args)
	command_s := strings.Join(command, " ")
	log.Printf("Launching Docker container using the following command:\n\n%s\n\n", command_s)
	if !utils.GetBoolFlag(cmd, "dry-run") {
		utils.RunCmd(command_s)
	}
}

var ldcCmd = &cobra.Command{
	Use:     "ldc [flags] image_name[:tag] [image_command]",
	Aliases: []string{},
	Short:   "Launch a container of a Docker image.",
	Args:    cobra.MinimumNArgs(1),
	Run:     ldc,
}

func init() {
	ldcCmd.Flags().BoolP("detach", "d", false, "Run container in background and print container ID.")
	ldcCmd.Flags().IntP("port", "p", 0, "The port on the Docker host to forward to the port inside the Docker container.")
	ldcCmd.Flags().StringP("user", "u", "", "The user to create in the Docker container.")
	ldcCmd.Flags().StringP("password", "P", "", "The default password for the user (to create in the Docker container).")
	ldcCmd.Flags().StringSlice("extra-port-mappings", []string{}, "Extra port mappings.")
	ldcCmd.Flags().BoolP("mount-home", "m", false, "Mount /home on the host as /home_host in the Docker container.")
	ldcCmd.Flags().Bool("dry-run", false, "Print out the docker command without running it.")
	rootCmd.AddCommand(ldcCmd)
}
