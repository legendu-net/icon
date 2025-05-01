package utils

import (
	"embed"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"periph.io/x/host/v3/distro"
)

//go:embed data
var data embed.FS

func ReadEmbedFile(path string) []byte {
	bytes, err := data.ReadFile(path)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return bytes
}

func ReadEmbedFileAsString(name string) string {
	return string(ReadEmbedFile(name))
}

func GetFileMode(file string) fs.FileMode {
	fileInfo, err := os.Stat(file)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return fileInfo.Mode()
}

func copyFile(sourceFile string, destinationFile string) {
	input, err := os.ReadFile(sourceFile)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	WriteFile(destinationFile, input, 0o600)
	Chmod(destinationFile, GetFileMode(sourceFile))
	log.Printf("%s is copied to %s.\n", sourceFile, destinationFile)
}

func copyFileToDir(sourceFile string, destinationDir string) {
	destinationFile := filepath.Join(destinationDir, filepath.Base(sourceFile))
	copyFile(sourceFile, destinationFile)
}

func CopyDir(sourceDir string, destinationDir string) {
	if !ExistsDir(destinationDir) {
		MkdirAll(destinationDir, GetFileMode(sourceDir))
	}
	for _, entry := range ReadDir(sourceDir) {
		if entry.IsDir() {
			srcDir := filepath.Join(sourceDir, entry.Name())
			dstDir := filepath.Join(destinationDir, entry.Name())
			CopyDir(srcDir, dstDir)
		} else {
			sourceFile := filepath.Join(sourceDir, entry.Name())
			if !IsSocket(sourceFile) {
				copyFile(sourceFile, filepath.Join(destinationDir, entry.Name()))
			}
		}
	}
}

// NormalizePath expands the leading '~' in a path to the user's home directory.
func NormalizePath(path string) string {
	if strings.HasPrefix(path, "~") {
		return filepath.Join(UserHomeDir(), path[1:])
	}
	return path
}

func Chmod(path string, mode fs.FileMode) {
	err := os.Chmod(path, mode)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
}

func Chmod600(path string) {
	if ExistsDir(path) {
		Chmod(path, 0o700)
		for _, entry := range ReadDir(path) {
			Chmod600(filepath.Join(path, entry.Name()))
		}
	} else {
		Chmod(path, 0o600)
	}
}

func CopyEmbedFile(sourceFile string, destinationFile string, mode os.FileMode, info bool) {
	bytes := ReadEmbedFile(sourceFile)
	dir := filepath.Dir(destinationFile)
	if !ExistsPath(dir) {
		err := os.MkdirAll(dir, 0o700)
		if err != nil {
			log.Fatal("ERROR - ", err)
		}
	}
	WriteFile(destinationFile, bytes, 0o600)
	Chmod(destinationFile, mode)
	if info {
		log.Printf("%s is copied to %s.\n", sourceFile, destinationFile)
	}
}

func CopyEmbedFileToDir(sourceFile string, destinationDir string, mode os.FileMode, info bool) {
	destinationFile := filepath.Join(destinationDir, filepath.Base(sourceFile))
	CopyEmbedFile(sourceFile, destinationFile, mode, info)
}

func RunCmd(cmd string, env ...string) {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.Command("pwsh", "-Command", cmd)
	case "linux", "darwin":
		command = exec.Command("bash", "-c", cmd)
	default:
		log.Fatal("ERROR - The OS ", runtime.GOOS, " is not supported!")
	}
	command.Env = append(os.Environ(), env...)
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

func HttpGetAsBytes(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode > 399 {
		log.Fatal(
			"HTTP request got an error response with the status code ",
			resp.StatusCode,
			"\n",
			"x-ratelimit-limit: ",
			resp.Header.Get("x-ratelimit-limit"),
			"\n",
			"x-ratelimit-remaining: ",
			resp.Header.Get("x-ratelimit-remaining"),
			"\n",
			"x-ratelimit-reset: ",
			time.Unix(ParseInt(resp.Header.Get("x-ratelimit-reset")), 0).Local(),
			"\n",
		)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return body
}

func HttpGetAsString(url string) string {
	return string(HttpGetAsBytes(url))
}

func CreateTempDir(pattern string) string {
	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return dir
}

// Download file from the given URL.
func DownloadFile(url string, name string, useTempDir bool) string {
	var out *os.File
	var err error
	if useTempDir {
		name = filepath.Join(CreateTempDir(""), name)
	}
	out, err = os.Create(name)
	if err != nil {
		log.Fatal("ERROR - ", err, ": ", name)
	}
	defer out.Close()
	log.Printf("Downloading %s to %s\n", url, out.Name())
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return out.Name()
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
	_, err := exec.LookPath("sudo")
	if err != nil {
		return ""
	}
	runWithSudo = strings.TrimSpace(runWithSudo)
	if runWithSudo != "" {
		RunCmd("sudo " + runWithSudo)
	}
	return "sudo"
}

func GetCommandPrefix(forceSudo bool, pathPerms map[string]uint32) string {
	switch runtime.GOOS {
	case "darwin", "linux":
		if GetCurrentUser().Uid != "0" {
			if forceSudo {
				return sudo("true")
			}
			for path, perm := range pathPerms {
				for !ExistsPath(path) {
					path = filepath.Dir(path)
				}
				if unix.Access(path, perm) != nil {
					return sudo("true")
				}
			}
		}
	}
	return ""
}

func ExistsPath(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
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
	bytes, err := io.ReadAll(readCloser)
	readCloser.Close()
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return string(bytes)
}

func ReadFile(path string) []byte {
	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return bytes
}

func ReadFileAsString(path string) string {
	return string(ReadFile(path))
}

func WriteFile(fileName string, data []byte, perm fs.FileMode) {
	err := os.WriteFile(fileName, data, perm)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
}

func WriteTextFile(path string, text string, perm fs.FileMode) {
	WriteFile(path, []byte(text), perm)
}

/*
*
Update a text file by replacing patterns with specified substitutions.
*/
func ReplacePattern(path string, pattern string, repl string) {
	text := ReadFileAsString(path)
	text = strings.ReplaceAll(text, pattern, repl)
	WriteTextFile(path, text, GetFileMode(path))
}

func AppendToTextFile(path string, text string, checkExistence bool) {
	if checkExistence {
		fileContent := ""
		if ExistsFile(path) {
			fileContent = ReadFileAsString(path)
		}
		if !strings.Contains(fileContent, strings.TrimSpace(text)) {
			AppendToTextFile(path, text, false)
		}
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	defer f.Close()
	_, err = f.WriteString(text)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
}

func GetBashConfigFile() string {
	home := UserHomeDir()
	file := ".bash_profile"
	if runtime.GOOS == "linux" {
		file = ".bashrc"
	}
	return filepath.Join(home, file)
}

func ConfigBash() {
	bashConfigFile := GetBashConfigFile()
	ConfigShellPath(bashConfigFile)
	AppendToTextFile(
		bashConfigFile,
		`
if which nvim > /dev/null; then
	export EDITOR=nvim
else
	export EDITOR=vim
fi
`,
		true,
	)
	if runtime.GOOS == "linux" {
		sourceIn := `
# source in ~/.bashrc
if [[ -f $HOME/.bashrc ]]; then
	. $HOME/.bashrc
fi
`
		AppendToTextFile(
			filepath.Join(UserHomeDir(), ".bash_profile"),
			sourceIn,
			true,
		)
	}
}

// Configure shell to add a path into the environment variable PATH.
// @param paths: Absolute paths to add into PATH.
// @param config_file: The path of a shell's configuration file.
func ConfigShellPath(config_file string) {
	if GetHostPlatform() == "idx" {
		return
	}
	text := ReadFileAsString(config_file)
	if !strings.Contains(text, ". /scripts/path.sh") && !strings.Contains(text, "\n_PATHS=(\n") {
		text = `
# set $PATH
_PATHS=(
	$(ls -d $HOME/*/bin 2> /dev/null)
	$(ls -d $HOME/.*/bin 2> /dev/null)
	$(ls -d $HOME/Library/Python/3.*/bin 2> /dev/null)
	$(ls -d /usr/local/*/bin 2> /dev/null)
)
for ((_i=${#_PATHS[@]}-1; _i>=0; _i--)); do
	_PATH=${_PATHS[$_i]}
	if [[ -d $_PATH && ! "$PATH" =~ (^$_PATH:)|(:$_PATH:)|(:$_PATH$) ]]; then
		export PATH=$_PATH:$PATH
	fi
done
`
		AppendToTextFile(config_file, text, true)
	}
	log.Printf("%s is configured to insert common bin paths into $PATH.", config_file)
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

func BuildPipUninstall(cmd *cobra.Command) string {
	python := GetStringFlag(cmd, "python")
	return Format("{python} -m pip uninstall", map[string]string{
		"python": python,
	})
}

func BuildPipInstall(cmd *cobra.Command) string {
	python := GetStringFlag(cmd, "python")
	_, err := exec.LookPath(python)
	if err != nil {
		return ""
	}
	user := ""
	if GetBoolFlag(cmd, "user") {
		user = "--user"
	}
	extraPipOptions := GetStringSliceFlag(cmd, "extra-pip-options")
	options := ""
	for _, option := range extraPipOptions {
		options += "--" + option
	}
	return Format("PIP_BREAK_SYSTEM_PACKAGES=1 {python} -m pip install {user} {options}", map[string]string{
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

func IsLinux() bool {
	switch runtime.GOOS {
	case "linux":
		return true
	default:
		return false
	}
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
	}
	distId := GetLinuxDistId()
	for _, id := range ids {
		if distId == id {
			return true
		}
	}
	return false
}

func IsDebianUbuntuSeries() bool {
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

// Using Homebrew to install packages
// without throwing exceptions if a package to install already exists.
// @param pkgs: A list of packages to install using Homebrew.
func BrewInstallSafe(pkgs []string) {
	for _, pkg := range pkgs {
		command := Format("brew install --force {pkg} || brew link --overwrite --force {pkg}", map[string]string{
			"pkg": pkg,
		})
		RunCmd(command)
	}
}

// Check if a file is a socket.
func IsSocket(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return fileInfo.Mode().Type() == fs.ModeSocket
}

func LinkFile(srcFile string, dstLink string) {
	err := os.Symlink(srcFile, dstLink)
	if err != nil {
		log.Fatalf("Failed to link the file %s to %s!\n", srcFile, dstLink)
	}
}

// Update map1 using map2.
func UpdateMap(map1 orderedmap.OrderedMap[string, any], map2 orderedmap.OrderedMap[string, any]) {
	for _, key2 := range map2.Keys() {
		val2, _ := map2.Get(key2)
		val1, map1HasKey2 := map1.Get(key2)
		if !map1HasKey2 {
			map1.Set(key2, val2)
			continue
		}
		switch val2.(type) {
		case orderedmap.OrderedMap[string, any]:
			switch val1.(type) {
			case orderedmap.OrderedMap[string, any]:
				UpdateMap(val1.(orderedmap.OrderedMap[string, any]), val2.(orderedmap.OrderedMap[string, any]))
			default:
				map1.Set(key2, val2)
			}
		default:
			map1.Set(key2, val2)
		}
	}
}

func BuildKernelOSKeywords(keywords map[string][]string) []string {
	kwds := keywords["common"]
	info, err := host.Info()
	if err != nil {
		log.Fatal(err)
	}
	switch info.KernelArch {
	case "x86_64":
		x86_64, found := keywords["x86_64"]
		if found {
			kwds = append(kwds, x86_64...)
		}
	case "arm64", "aarch64":
		arm64, found := keywords["arm64"]
		if found {
			kwds = append(kwds, arm64...)
		}
	default:
	}
	switch runtime.GOOS {
	case "darwin":
		darwin, found := keywords["darwin"]
		if found {
			kwds = append(kwds, darwin...)
		}
	case "linux":
		linux, found := keywords["linux"]
		if found {
			kwds = append(kwds, linux...)
		}
		if IsDebianUbuntuSeries() {
			debianUbuntuSeries, found := keywords["DebianUbuntuSeries"]
			if found {
				kwds = append(kwds, debianUbuntuSeries...)
			}
		} else if IsFedoraSeries() {
			fedoraSeries, found := keywords["FedoraSeries"]
			if found {
				kwds = append(kwds, fedoraSeries...)
			}
		} else {
			otherLinux, found := keywords["OtherLinux"]
			if found {
				kwds = append(kwds, otherLinux...)
			}
		}
	default:
	}
	return kwds
}

func GetHostPlatform() string {
	h, err := host.Info()
	if err != nil {
		log.Fatal(err)
	}
	return h.Platform
}

func ParseInt(str string) int64 {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		log.Fatal("Error converting string to int64: %v\n", err)
	}
	return i
}
