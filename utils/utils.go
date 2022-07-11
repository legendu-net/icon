package utils

import (
	"bytes"
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
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

func GetCommandPrefix(pathPerms map[string]uint32, runWithSudo string) string {
	runWithSudo = strings.TrimSpace(runWithSudo)
	switch runtime.GOOS {
	case "darwin", "linux":
		if GetCurrentUser().Uid != "0" {
			for path, perm := range pathPerms {
				for !ExistsPath(path) {
					path = filepath.Dir(path)
				}
				if unix.Access(path, perm) != nil {
					if runWithSudo != "" {
						RunCmd("sudo " + runWithSudo)
					}
					return "sudo"
				}
			}
		}
	}
	return ""
}

func ExistsPath(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func Getwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return cwd
}

func UserHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return home
}

func RemoveAll(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		log.Fatal(err)
	}
}

func GetBoolFlag(cmd *cobra.Command, flag string) bool {
	b, err := cmd.Flags().GetBool(flag)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func GetIntFlag(cmd *cobra.Command, flag string) int {
	i, err := cmd.Flags().GetInt(flag)
	if err != nil {
		log.Fatal(err)
	}
	return i
}

func GetStringSliceFlag(cmd *cobra.Command, flag string) []string {
	ss, err := cmd.Flags().GetStringSlice(flag)
	if err != nil {
		log.Fatal(err)
	}
	return ss
}

func ReadDir(dir string) []os.DirEntry {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	return entries
}

func ReadAllText(readCloser io.ReadCloser) string {
	bytes, err := ioutil.ReadAll(readCloser)
	readCloser.Close()
	if err != nil {
		log.Fatal(err)
	}
	return string(bytes)
}

func ReadTextFile(path string) string {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(bytes)
}

func AppendToTextFile(path string, text string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.WriteString(text)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}
}

// Configure shell to add a path into the environment variable PATH.
// @param paths: Absolute paths to add into PATH.
// @param config_file: The path of a shell's configuration file.
func AddPathShell(paths []string, config_file string) {
	text := ReadTextFile(config_file)
	pattern := "\n_PATHS=(\n"
	if strings.Contains(text, pattern) {
		lines := ""
		for _, path := range paths {
			if strings.Contains(path, "*") {
				lines += "    $(ls -d " + path + " 2> /dev/null)\n"
			} else {
				lines += "    \"" + path + "\"\n"
			}
		}
		text = strings.Replace(text, pattern+lines, "", 1)
	} else {
		text = `
# set $PATH
_PATHS=(
	$(ls -d $HOME/*/bin 2> /dev/null)
	$(ls -d $HOME/Library/Python/3.*/bin 2> /dev/null)
)
for ((_i=${#_PATHS[@]}-1; _i>=0; _i--)); do
	_PATH=${_PATHS[$_i]}
	if [[ -d $_PATH && ! "$PATH" =~ (^$_PATH:)|(:$_PATH:)|(:$_PATH$) ]]; then
		export PATH=$_PATH:$PATH
	fi
done
`
		AppendToTextFile(config_file, text)
	}
}

func VirtualMemory() *mem.VirtualMemoryStat {
	memStat, err := mem.VirtualMemory()
	if err != nil {
		log.Fatal(err)
	}
	return memStat
}

func CpuInfo() []cpu.InfoStat {
	cpuInfo, err := cpu.Info()
	if err != nil {
		log.Fatal(err)
	}
	return cpuInfo
}
