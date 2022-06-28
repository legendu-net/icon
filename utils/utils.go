package utils

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
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
	err := command.Run()
	if err != nil {
		log.Fatal(err, " when running the command: ", cmd)
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
	log.Printf("Spark has been downloaded to %s", out.Name())
	return out
}
