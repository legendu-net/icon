package cmd

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
	"log"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func getDockerImagePort(imageName string) int {
	if strings.HasPrefix(imageName, "dclong/") {
		imageName = imageName[7:]
		if strings.HasPrefix(imageName, "jupyterlab") {
			return 8888
		}
		if strings.HasPrefix(imageName, "jupyterhub") {
			return 8000
		}
		if strings.HasPrefix(imageName, "vscode") {
			return 8080
		}
	}
	return 0
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
	userName := currentUser.Username
	userId := currentUser.Uid
	groupId := currentUser.Gid
	command := []string{
		"docker",
		"run",
		"-it",
		"--init",
		"--platform",
		"linux/amd64",
		"--log-opt",
		"max-size=50m",
		"-e",
		"DOCKER_USER=" + userName,
		"-e",
		"DOCKER_USER_ID=" + userId,
		"-e",
		"DOCKER_PASSWORD=" + userName,
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
	command = append(command, "-v", filepath.Dir(home)+":/home_host")
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
	port := getDockerImagePort(args[0])
	if port > 0 {
		portHost := utils.GetIntFlag(cmd, "port")
		if portHost == 0 {
			portHost = port
		}
		command = append(command, "--publish="+strconv.Itoa(portHost)+":"+strconv.Itoa(port))
	}
	extraPortMappings := utils.GetStringSliceFlag(cmd, "extra-port-mappings")
	if len(extraPortMappings) > 0 {
		for _, m := range extraPortMappings {
			command = append(command, "--publish="+m)
		}
	}
	command = append(command, args...)
	if len(args) == 1 && strings.HasPrefix(args[0], "dclong/") {
		command = append(command, "/scripts/sys/init.sh")
	}
	command_s := strings.Join(command, " ")
	log.Printf("Launching Docker container using the following command:\n\n%s\n\n", command_s)
	utils.RunCmd(command_s)
}

var ldcCmd = &cobra.Command{
	Use:     "ldc [flags] image_name[:tag] [image_command]",
	Aliases: []string{},
	Short:   "Launch a container of a Docker image.",
	Args:    cobra.MinimumNArgs(1),
	Run:     ldc,
}

func init() {
	ldcCmd.Flags().BoolP("detach", "d", false, "If specified, run container in background and print container ID.")
	ldcCmd.Flags().IntP("port", "p", 0, "The port on the Docker host to forward to the port inside the Docker container.")
	ldcCmd.Flags().StringSlice("extra-port-mappings", []string{}, "Extra port mappings.")
	rootCmd.AddCommand(ldcCmd)
}
