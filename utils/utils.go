package utils

import (
	"os/exec"
	"runtime"
	"log"
	"strings"
)

func RunCmd(cmd string) {
	var command *exec.Cmd
    switch runtime.GOOS {
		case "windows":
			command = exec.Command("pwsh", "-Command", cmd)
		default:
			command = exec.Command("bash", "-c", cmd)
	}
	err := command.Run()
	if err != nil {
		log.Fatal(err, " when running the command: ", cmd)
	}
}

func Format(cmd string, hmap map[string]string) string {
	for key, val := range hmap {
		cmd = strings.ReplaceAll(cmd, "{" + key + "}", val)
	}
	return cmd
}
