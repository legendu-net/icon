package utils

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
)

func RunCmd(cmd string) {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.Command("pwsh", "-Command", cmd)
	case "linux", "darwin":
		command = exec.Command("bash", "-c", cmd)
	default:
		log.Fatal("The OS ", runtime.GOOS, " is not supported!")
	}
	var out bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &out
	command.Stderr = &stderr
	err := command.Run()
	if err != nil {
		log.Fatal(fmt.Sprint(err)+": "+stderr.String()+" when running the command:\n", cmd)
	}
}

func Format(cmd string, hmap map[string]string) string {
	for key, val := range hmap {
		cmd = strings.ReplaceAll(cmd, "{"+key+"}", val)
	}
	return cmd
}

// Download file from the given URL.
func DownloadFile(url string, name string) *os.File {
	log.Printf("Downloading %s from: %s\n", name, url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	// create a temp file to receive the download
	out, err := os.CreateTemp(os.TempDir(), name)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(out, resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s has been downloaded to %s", name, out.Name())
	return out
}

func Max(x int, y int) int {
	if x >= y {
		return x
	}
	return y
}

func GetCurrentUser() *user.User {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return currentUser
}