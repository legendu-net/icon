package utils

import (
	"embed"
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
	"periph.io/x/host/v3/distro"
	"runtime"
	"strings"
)

//go:embed data
var Data embed.FS

func ReadEmbedFile(name string) []byte {
	bytes, err := Data.ReadFile(name)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return bytes
}

func ReadEmbedFileAsString(name string) string {
	return string(ReadEmbedFile(name))
}

/*
func CopyFile(sourceFile string, destinationFile string) {
	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	err = ioutil.WriteFile(destinationFile, input, 0600)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	log.Printf("%s is copied to %s.\n", sourceFile, destinationFile)
}
*/

func CopyEmbedFile(sourceFile string, destinationFile string) {
	bytes := ReadEmbedFile(sourceFile)
	dir := filepath.Dir(destinationFile)
	if !ExistsPath(dir) {
		err := os.MkdirAll(dir, 0700)
		if err != nil {
			log.Fatal("ERROR - ", err)
		}
	}
	err := ioutil.WriteFile(destinationFile, bytes, 0600)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	fileInfo, err := os.Stat(sourceFile)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	err = os.Chmod(destinationFile, fileInfo.Mode())
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	log.Printf("%s is copied to %s.\n", sourceFile, destinationFile)
}

func CopyEmbedFileToDir(sourceFile string, destinationDir string) {
	destinationFile := filepath.Join(destinationDir, filepath.Base(sourceFile))
	CopyEmbedFile(sourceFile, destinationFile)
}

func RunCmd(cmd string) {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.Command("pwsh", "-Command", cmd)
	case "linux", "darwin":
		command = exec.Command("bash", "-c", cmd)
	default:
		log.Fatal("ERROR - The OS ", runtime.GOOS, " is not supported!")
	}
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		log.Fatal("ERROR - ", err, ": when running the command:\n", cmd)
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
		log.Fatal("ERROR - ", err)
	}
	// create a temp file to receive the download
	out, err := os.CreateTemp(os.TempDir(), name)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	_, err = io.Copy(out, resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal("ERROR - ", err)
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
		log.Fatal("ERROR - ", err)
	}
	return currentUser
}

func sudo(runWithSudo string) string {
	runWithSudo = strings.TrimSpace(runWithSudo)
	if runWithSudo != "" {
		RunCmd("sudo " + runWithSudo)
	}
	return "sudo"
}

func GetCommandPrefix(forceSudo bool, pathPerms map[string]uint32, runWithSudo string) string {
	switch runtime.GOOS {
	case "darwin", "linux":
		if GetCurrentUser().Uid != "0" {
			if forceSudo {
				return sudo(runWithSudo)
			}
			for path, perm := range pathPerms {
				for !ExistsPath(path) {
					path = filepath.Dir(path)
				}
				if unix.Access(path, perm) != nil {
					return sudo(runWithSudo)
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

func ExistsDir(path string) bool {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return stat.IsDir()
}

func ExistsFile(path string) bool {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !stat.IsDir()
}

func Getwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return cwd
}

func UserHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return home
}

func RemoveAll(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
}

func GetBoolFlag(cmd *cobra.Command, flag string) bool {
	b, err := cmd.Flags().GetBool(flag)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return b
}

func GetIntFlag(cmd *cobra.Command, flag string) int {
	i, err := cmd.Flags().GetInt(flag)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return i
}

func GetStringFlag(cmd *cobra.Command, flag string) string {
	s, err := cmd.Flags().GetString(flag)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return s
}

func GetStringSliceFlag(cmd *cobra.Command, flag string) []string {
	ss, err := cmd.Flags().GetStringSlice(flag)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return ss
}

func ReadDir(dir string) []os.DirEntry {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return entries
}

func ReadAllAsText(readCloser io.ReadCloser) string {
	bytes, err := ioutil.ReadAll(readCloser)
	readCloser.Close()
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return string(bytes)
}

func ReadTextFile(path string) string {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return string(bytes)
}

func WriteTextFile(path string, text string) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	_, err = f.WriteString(text)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
}

func AppendToTextFile(path string, text string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	_, err = f.WriteString(text)
	f.Close()
	if err != nil {
		log.Fatal("ERROR - ", err)
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
		log.Fatal("ERROR - ", err)
	}
	return memStat
}

func CpuInfo() []cpu.InfoStat {
	cpuInfo, err := cpu.Info()
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return cpuInfo
}

func BuildYesFlag(cmd *cobra.Command) string {
	return IfElseString(GetBoolFlag(cmd, "yes"), "-y", "")
}

func BuildPipInstall(cmd *cobra.Command) string {
	python := GetStringFlag(cmd, "python")
	user := ""
	if GetBoolFlag(cmd, "user") {
		user = "--user"
	}
	extraPipOptions := GetStringSliceFlag(cmd, "extra-pip-options")
	options := ""
	for _, option := range extraPipOptions {
		options += "--" + option
	}
	return Format("{python} -m pip install {user} {options}", map[string]string{
		"python":  python,
		"user":    user,
		"options": options,
	})
}

func ExistsCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func MkdirAll(path string, perm os.FileMode) {
	err := os.MkdirAll(path, perm)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
}

func AddPythonFlags(cmd *cobra.Command) {
	cmd.Flags().String("python", "python3", "Path to the python3 command.")
	cmd.Flags().Bool("user", false, "Install Python packages to user's local directory.")
	cmd.Flags().StringSlice("extra-pip-options", []string{}, "Extra options (separated by comma) to pass to pip.")
}

func GetLinuxDistId() string {
	m := distro.OSRelease()
	distId, found := m["ID"]
	if found {
		return distId
	} else {
		return ""
	}
}

func IsUbuntu() bool {
	return GetLinuxDistId() == "ubuntu"
}

func IsDebian() bool {
	return GetLinuxDistId() == "debian"
}

func IsDebianSeries() bool {
	ids := []string{
		"debian",
		"antix",
		"lmde",
		"ubuntu", "linuxmint", "pop",
	}
	distId := GetLinuxDistId()
	for _, id := range ids {
		if distId == id {
			return true
		}
	}
	return false
}

func IsUbuntuSeries() bool {
	ids := []string{
		"ubuntu", "linuxmint", "pop",
	}
	distId := GetLinuxDistId()
	for _, id := range ids {
		if distId == id {
			return true
		}
	}
	return false
}

func IsFedoraSeries() bool {
	ids := []string{
		"fedora", "centos", "rhel",
	}
	distId := GetLinuxDistId()
	for _, id := range ids {
		if distId == id {
			return true
		}
	}
	return false
}

func IfElseString(b bool, t string, f string) string {
	if b {
		return t
	} else {
		return f
	}
}
