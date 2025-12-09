package utils

import (
	"fmt"
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

// GetFileMode retrieves the file mode (permissions) of a given file.
//
// This function uses os.Stat to get file information and returns the file mode.
// If an error occurs during the file stat, the function will terminate with a fatal log.
//
// @param file The path to the file.
// @return The file mode (fs.FileMode).
func GetFileMode(file string) fs.FileMode {
	fileInfo, err := os.Stat(file)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return fileInfo.Mode()
}

// copyFile copies a file from the source path to the destination path.
//
// This function reads the content of the source file, writes it to the destination file,
// and then sets the destination file's permissions to match those of the source file.
// If any error occurs during these operations, the function will terminate with a fatal log.
//
// @param sourceFile      The path to the source file.
// @param destinationFile The path to the destination file where the source file will be copied.
func CopyFile(sourceFile string, destinationFile string) {
	sourceFile = NormalizePath(sourceFile)
	destinationFile = NormalizePath(destinationFile)
	MkdirAll(filepath.Dir(destinationFile), 0o700)
	input, err := os.ReadFile(sourceFile)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	WriteFile(destinationFile, input, 0o600)
	Chmod(destinationFile, GetFileMode(sourceFile))
	log.Printf("%s is copied to %s.\n", sourceFile, destinationFile)
}

// CopyFileToDir copies a file from a source path to a destination directory.
//
// It constructs the destination file path by joining the destination directory
// with the base name of the source file. Then it calls the copyFile function to perform the actual copy.
//
// @param sourceFile      The path to the source file.
// @param destinationDir The path to the destination directory where the source file will be copied.
func CopyFileToDir(sourceFile string, destinationDir string) {
	CopyFile(sourceFile, filepath.Join(destinationDir, filepath.Base(sourceFile)))
}

// CopyDir recursively copies a source directory to a destination directory.
//
// The destination directory is created with the same permission as the source directory
// if it does not already exist. This function behaves similar to the Linux command `cp -r`.
//
// @param sourceDir      The path to the source directory.
// @param destinationDir The path to the destination directory where the source directory
//
//	and its contents will be copied.
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
				CopyFile(sourceFile, filepath.Join(destinationDir, entry.Name()))
			}
		}
	}
}

// NormalizePath normalizes a given path string.
//
// This function handles paths that start with "~" which represents the user's home directory.
// If the path starts with "~", it replaces it with the user's home directory obtained
// from UserHomeDir(). Otherwise, it returns the path as is.
//
// @param path The path string to normalize.
// @return The normalized path string.
func NormalizePath(path string) string {
	if path == "~" {
		return UserHomeDir()
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(UserHomeDir(), path[2:])
	}
	return path
}

// Chmod changes the mode of the named file to mode.
//
// If an error occurs during the chmod operation, the function will terminate with a fatal log.
//
// @param path The path to the file.
func Chmod(path string, mode fs.FileMode) {
	err := os.Chmod(path, mode)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
}

// Chmod600 recursively changes the file mode of a file or directory to 0600 (rw-------).
//
// If the given path is a directory, this function will recursively apply the 0700
// permissions (rwx------) to the directory and 0600 to all files within it. If the
// path is a file, it will simply apply 0600 permissions to that file.
//
// @param path The path to the file or directory.
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

// RunCmd executes a command (using pwsh or bash based on the OS) in the terminal.
//
// @param cmd The command to execute as a string.
// @param env Optional environment variables to set for the command execution.
//
// @example RunCmd("ls -l", "MY_VAR=myvalue")
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

// Format replaces placeholders in a string with values from a map.
//
// This function takes a command string `cmd` and a map `hmap` as input. It
// iterates through the map, replacing each occurrence of a placeholder in
// `cmd` with the corresponding value from the map. Placeholders are denoted
// by the format "{key}" where "key" is a key in the `hmap`.
//
// @param cmd The command string with placeholders.
// @param hmap A map where keys are placeholder names and values are their replacements.
//
// @return The formatted command string with placeholders replaced by their values.
//
// @example Format("Hello, {name}!", map[string]string{"name": "World"}) // Returns "Hello, World!"
func Format(cmd string, hmap map[string]string) string {
	for key, val := range hmap {
		cmd = strings.ReplaceAll(cmd, "{"+key+"}", val)
	}
	return cmd
}

// HttpGetAsBytes performs an HTTP GET request to the specified URL and returns the response body as a byte slice.
//
// This function sends an HTTP GET request to the given URL. It handles potential errors during the request,
// retries with exponential backoff, and deals with rate limiting if the server indicates it. If the
// request is successful, it reads the response body and returns it as a byte slice. If any error occurs,
// the function terminates with a fatal log message.
//
// @param url The URL to send the HTTP GET request to.
// @param retry The number of times to retry the request if it fails or encounters a rate limit.
// @param initial_waiting_seconds The initial number of seconds to wait before retrying the request.
//
// @return The response body as a byte slice.
//
// @example
//
//	body := HttpGetAsBytes("https://api.example.com/data", 3, 1)
//	// Process the body data
//
// @remarks
//
//	The function implements the following retry logic:
//	  - If an error occurs, it waits for `initial_waiting_seconds` and retries up to `retry` times.
//	  - If a rate limit is encountered (x-ratelimit-remaining == "0"), it waits until the rate limit resets
//	    (x-ratelimit-reset) plus an additional 10 seconds before retrying.
//	  - If a non-success status code is received, it prints out the `x-ratelimit-limit`, `x-ratelimit-remaining`
//	    and `x-ratelimit-reset` for debugging purposes.
//	  - If it failed to read the body, it retries with exponential backoff.
//	The function will terminate with log.Fatal if:
//	  - The HTTP GET request got an error and retry count is exhausted.
//	  - The HTTP GET request got an error response (status code > 399) and retry count is exhausted.
//	  - Error reading the response body and retry count is exhausted.
//
//	The retry delay increases exponentially with each retry attempt (initial_waiting_seconds * 2 for each retry).
func HttpGetAsBytes(url string, retry int8, initial_waiting_seconds int32) []byte {
	resp, err := http.Get(url)
	if err != nil {
		if retry > 0 {
			time.Sleep(time.Duration(initial_waiting_seconds) * time.Second)
			return HttpGetAsBytes(url, retry-1, initial_waiting_seconds*2)
		}
		log.Fatal("The HTTP GET request on the URL ", url, " got the following error:\n", err)
	}
	if resp.StatusCode > 399 {
		if resp.Header.Get("x-ratelimit-remaining") == "0" {
			time.Sleep(time.Until(time.Unix(ParseInt(resp.Header.Get("x-ratelimit-reset"))+10, 0)))
			return HttpGetAsBytes(url, retry, initial_waiting_seconds)
		}
		if retry > 0 {
			time.Sleep(time.Duration(initial_waiting_seconds) * time.Second)
			return HttpGetAsBytes(url, retry-1, initial_waiting_seconds*2)
		}
		log.Fatal(
			"The HTTP GET request on the URL ", url, " got an error response with the status code ",
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
		if retry > 0 {
			time.Sleep(time.Duration(initial_waiting_seconds) * time.Second)
			return HttpGetAsBytes(url, retry-1, initial_waiting_seconds*2)
		}
		log.Fatal("Reading the response body of the http GET request on the url ", url, " got the following error:\n", err)
	}
	return body
}

// HttpGetAsString performs an HTTP GET request to the specified URL and returns the response body as a string.
//
// This function sends an HTTP GET request to the given URL, leveraging the HttpGetAsBytes function
// to handle the HTTP request, retries, and rate limiting. It then converts the response body
// from a byte slice to a string. If any error occurs during the request or body conversion,
// the function will terminate with a fatal log message.
//
// @param url The URL to send the HTTP GET request to.
// @param retry The number of times to retry the request if it fails or encounters a rate limit.
// @param initial_waiting_seconds The initial number of seconds to wait before retrying the request.
//
// @return The response body as a string.
//
// @example
//
//	body := HttpGetAsString("https://api.example.com/data", 3, 1)
//	// Process the body string
//
// @remarks
//
//	This function is a wrapper around HttpGetAsBytes and provides a string-based interface to the HTTP response body.
//	Please refer to the documentation of HttpGetAsBytes for more detailed information about the retry and rate limiting
//	logic employed.
func HttpGetAsString(url string, retry int8, initial_waiting_seconds int32) string {
	return string(HttpGetAsBytes(url, retry, initial_waiting_seconds))
}

// CreateTempDir creates a new temporary directory.
//
// This function creates a new temporary directory in the default directory for temporary files.
// The directory name is generated with the given pattern.
// If the pattern includes a "*", the "*" will be replaced with random characters.
// If an error occurs during the directory creation, the function will terminate with a fatal log.
//
// @param pattern The pattern for generating the directory name.
//
// @return The path to the created temporary directory.
//
// @example
// dir := CreateTempDir("my-temp-dir-*")
func CreateTempDir(pattern string) string {
	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return dir
}

// Download file from the given URL.
//
// This function downloads a file from a specified URL and saves it to the local
// filesystem. It can optionally save the file to a temporary directory, which is
// useful for files that don't need to be persisted long-term. If the download is
// successful, the function returns the local path where the file was saved.
// If any error occurs during the download or saving process, the function will
// terminate with a fatal log message.
//
// @param url The URL of the file to download.
// @param name The desired name of the file when saved locally.
// @param useTempDir If true, the file will be saved to a temporary directory.
//
// @return The local path where the downloaded file is saved.
//
// @example
//
//	path := DownloadFile("http://example.com/file.zip", "file.zip", true)
//	// Process the file at 'path'
//	// If useTempDir is true, the file will be in a temporary directory.
//
//	path := DownloadFile("http://example.com/file.zip", "file.zip", false)
//	// Process the file at 'path'
//	// If useTempDir is false, the file will be in current directory.
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

// Max returns the larger of two integers.
//
// This function compares two integers, x and y, and returns the larger of the two.
// If both integers are equal, it returns x.
//
// @param x The first integer.
// @param y The second integer.
//
// @return The larger of the two integers.
//
// @example
//
//	result := Max(5, 10) // Returns 10
//	result := Max(10, 5) // Returns 10
func Max(x int, y int) int {
	if x >= y {
		return x
	}
	return y
}

// GetCurrentUser retrieves the current user's information.
//
// This function uses the `user.Current()` method to get the current user's information.
// It returns a pointer to a `user.User` struct that contains various user-related details,
// such as the user ID, username, home directory, etc.
// If there is an error while retrieving the current user's information,
// the function will terminate with a fatal log message.
//
// @return A pointer to a `user.User` struct representing the current user.
//
// @example
//
//	currentUser := GetCurrentUser()
//	fmt.Println("Username:", currentUser.Username)
func GetCurrentUser() *user.User {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return currentUser
}

// Executes a command with `sudo` if the `sudo` command is available.
//
// This function attempts to execute a given command with `sudo`.
// First, it checks if the `sudo` command exists in the system's PATH.
// If `sudo` is not found, it returns an empty string.
// If `sudo` is found and `runWithSudo` is not an empty string,
// it runs the specified command with `sudo` using the RunCmd function.
//
// @param runWithSudo The command to run with sudo.
//
// @return "sudo" if the sudo command was executed, otherwise an empty string.
//
// @example
//
//	sudo("apt-get update")
//	// Executes "sudo apt-get update" if sudo is available.
//
//	sudo("")
//	// Returns "" without executing any command.
//
//	if sudo("apt-get install tree") == "sudo" {
//		// Do something after sudo command is executed successfully.
//	}
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

// GetCommandPrefix determines the appropriate command prefix for running commands.
//
// This function determines the command prefix (e.g., "sudo") needed for
// running commands based on the operating system, the current user's UID, and
// file permissions. If `forceSudo` is true, it always returns "sudo" on Linux
// and macOS if the current user is not root. Otherwise, it checks if the user
// has the necessary permissions to access specific files/directories specified
// in `pathPerms`. If the user lacks permissions, it returns "sudo".
//
// @param forceSudo A boolean indicating whether to force the use of sudo.
// @param pathPerms A map where keys are file paths and values are permission
//
//	bits (e.g., unix.W_OK for write permission).
//
// @return The appropriate command prefix ("sudo" or an empty string).
//
// @example
//
//	// Force the use of sudo for all commands.
//	prefix := GetCommandPrefix(true, nil)
//	// prefix will be "sudo" on Linux/macOS if the user is not root.
//
//	// Check if the user has write access to a specific directory.
//	pathPerms := map[string]uint32{
//		"/etc/myconfig": unix.W_OK,
//	}
//	prefix := GetCommandPrefix(false, pathPerms)
//	// prefix will be "sudo" on Linux/macOS if the user is not root and does
//	// not have write access to /etc/myconfig.
//
//	// Check if the user has write access to a specific directory and its parent.
//	pathPerms := map[string]uint32{
//		"/etc/myconfig/file.txt": unix.W_OK,
//	}
func GetCommandPrefix(forceSudo bool, pathPerms map[string]uint32) string {
	switch runtime.GOOS {
	case "darwin", "linux":
		if GetCurrentUser().Uid != "0" {
			if forceSudo {
				return sudo("true")
			}
			for path, perm := range pathPerms {
				path = NormalizePath(path)
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

// ExistsPath checks if a file or directory exists at the specified path.
//
// @param path The path to the file or directory.
//
// @return true if the file or directory exists, false otherwise.
//
// @example
//
//	if ExistsPath("/tmp/myfile.txt") { ... }
func ExistsPath(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// ExistsDir checks if a directory exists at the specified path.
//
// This function checks if a directory exists at the specified path using
// os.Stat. It returns true if a directory exists at the path, and false
// otherwise, including if there is an error or if a file, but not a
// directory, exists at the path.
//
// @param path The path to check for a directory.
//
// @return true if a directory exists at the path, false otherwise.
//
// @example
// if ExistsDir("/tmp/mydir") { ... }
func ExistsDir(path string) bool {
	stat, err := os.Stat(NormalizePath(path))
	if os.IsNotExist(err) {
		return false
	}
	return stat.IsDir()
}

// ExistsFile checks if a file exists at the specified path.
//
// This function uses `os.Stat` to get file information.
// It returns true if a file exists at the given path, and false otherwise.
// It returns false if a directory exists at the path.
//
// @param path The path to the file.
//
// @return true if a file exists at the path, false otherwise.
//
// @example
// if ExistsFile("/tmp/myfile.txt") { ... }
func ExistsFile(path string) bool {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !stat.IsDir()
}

// Getwd returns the current working directory.
//
// This function retrieves the current working directory (CWD) of the process.
// If the retrieval of the CWD fails, the function terminates with a fatal log.
//
// @return The current working directory as a string.
//
// @example
//
//	cwd := Getwd()
//	if cwd == "/home/user/project" {
//	  // ...
//	}
func Getwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return cwd
}

// UserHomeDir returns the current user's home directory.
//
// This function retrieves the home directory path for the current user.
// It internally uses `os.UserHomeDir()` to get the home directory path.
// If an error occurs during the retrieval of the home directory, the function
// terminates with a fatal log.
//
// @return The path to the current user's home directory as a string.
//
// @example
//
//	homeDir := UserHomeDir()
//	fmt.Println("Current user's home directory:", homeDir)
func UserHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return home
}

// RemoveAll removes the specified path and any children it contains.
//
// This function removes a file or directory at the given path, including all
// its subdirectories and files. If the path does not exist or an error occurs
// during removal, the function terminates with a fatal log message.
//
// @param path The path to the file or directory to remove.
//
// @example
//
//	RemoveAll("/tmp/mydir") // Removes /tmp/mydir and all its contents.
//	RemoveAll("/tmp/myfile.txt") // Removes /tmp/myfile.txt.
func RemoveAll(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
}

// GetBoolFlag retrieves the boolean value of a flag from a Cobra command.
//
// This function retrieves the value of a boolean flag from a Cobra command object.
// It uses the GetBool method of the command's flag set.
// If an error occurs while retrieving the flag's value, the function terminates with a fatal log message.
//
// @param cmd  A pointer to a Cobra command object.
// @param flag The name of the flag to retrieve.
//
// @return The boolean value of the flag.
//
// @example
// b := GetBoolFlag(myCmd, "myFlag")
func GetBoolFlag(cmd *cobra.Command, flag string) bool {
	b, err := cmd.Flags().GetBool(flag)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return b
}

// GetIntFlag retrieves the integer value of a flag from a Cobra command.
//
// This function retrieves the value of an integer flag from a Cobra command object.
// It uses the GetInt method of the command's flag set.
// If an error occurs while retrieving the flag's value, the function terminates with a fatal log message.
//
// @param cmd  A pointer to a Cobra command object.
// @param flag The name of the flag to retrieve.
//
// @return The integer value of the flag.
//
// @example
//
//	i := GetIntFlag(myCmd, "myFlag")
//	// i contains the value of the --myFlag flag
func GetIntFlag(cmd *cobra.Command, flag string) int {
	i, err := cmd.Flags().GetInt(flag)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return i
}

// GetStringFlag retrieves the string value of a flag from a Cobra command.
//
// This function retrieves the value of a string flag from a Cobra command object.
// It uses the GetString method of the command's flag set.
// If an error occurs while retrieving the flag's value, the function terminates with a fatal log message.
//
// @param cmd  A pointer to a Cobra command object.
// @param flag The name of the flag to retrieve.
//
// @return The string value of the flag.
//
// @example
//
//	s := GetStringFlag(myCmd, "myFlag")
//	// s contains the value of the --myFlag flag
//

func GetStringFlag(cmd *cobra.Command, flag string) string {
	s, err := cmd.Flags().GetString(flag)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return s
}

// GetStringSliceFlag retrieves the string slice value of a flag from a Cobra command.
//
// This function retrieves the value of a string slice flag from a Cobra command object.
// It uses the GetStringSlice method of the command's flag set.
// If an error occurs while retrieving the flag's value, the function terminates with a fatal log message.
//
// @param cmd  A pointer to a Cobra command object.
// @param flag The name of the flag to retrieve.
//
// @return The string slice value of the flag.
//
// @example
//
//	ss := GetStringSliceFlag(myCmd, "myFlag")
//	// ss contains the value of the --myFlag flag
//

func GetStringSliceFlag(cmd *cobra.Command, flag string) []string {
	ss, err := cmd.Flags().GetStringSlice(flag)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return ss
}

// ReadDir reads the named directory and returns
// a list of directory entries sorted by filename.
// If an error occurs during the directory reading, the function
// will terminate with a fatal log.
//
// @param dir The name of the directory to read.
//
// @return A slice of DirEntry representing the directory's contents.
//
// @example
//
//	entries := ReadDir("/tmp/mydir")
//	// Iterate through entries in /tmp/mydir.
func ReadDir(dir string) []os.DirEntry {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return entries
}

// ReadAllAsText reads all data from an io.ReadCloser and returns it as a string.
//
// This function reads all the data from the given io.ReadCloser, such as an HTTP
// response body or a file. It then closes the io.ReadCloser to release any associated
// resources. Finally, it converts the read data (which is initially a byte slice)
// into a string and returns it. If an error occurs during the read or close
// operation, the function will terminate with a fatal log message.
//
// @param readCloser The io.ReadCloser from which to read the data.
//
// @return The read data as a string.
//
// @example
//
//	responseBody := ReadAllAsText(resp.Body)
//	// Process the body string.
func ReadAllAsText(readCloser io.ReadCloser) string {
	bytes, err := io.ReadAll(readCloser)
	readCloser.Close()
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return string(bytes)
}

// ReadFile reads the entire content of a file and returns it as a byte slice.
//
// @param path The path to the file to be read.
//
// @return The content of the file as a slice of bytes.
//
// @example
//
//	content := ReadFile("/tmp/myfile.txt")
func ReadFile(path string) []byte {
	bytes, err := os.ReadFile(NormalizePath(path))
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return bytes
}

// ReadFileAsString reads the entire content of a file and returns it as a string.
//
// This function reads the file specified by 'path' and returns its entire
// content as a string. It utilizes the 'ReadFile' function internally to read
// the file content as bytes and then converts it to a string.
// If an error occurs while reading the file, the function will terminate with
// a fatal log message.
//
// @param path The path to the file to be read.
//
// @return The content of the file as a string.
//
// @example content := ReadFileAsString("/tmp/myfile.txt")
func ReadFileAsString(path string) string {
	return string(ReadFile(path))
}

// WriteFile writes a byte slice to a file with the specified permissions.
//
// This function writes the given byte slice `data` to a file specified by `fileName`.
// It sets the file's permissions to the specified `perm` value. If an error occurs
// during the file write operation, such as the file not being writable or permission
// issues, the function terminates with a fatal log message.
//
// @param fileName The name of the file to write to.
// @param data     The byte slice to write to the file.
// @param perm     The file mode (permissions) to set for the file.
//
// @example
// WriteFile("/tmp/myfile.txt", []byte("Hello, world!"), 0644)
func WriteFile(fileName string, data []byte, perm fs.FileMode) {
	err := os.WriteFile(fileName, data, perm)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
}

// WriteTextFile writes a string to a file with the specified permissions.
//
// This function writes the given string `text` to a file specified by `path`.
// It internally utilizes the `WriteFile` function to write the string to the
// file as a byte slice. The file's permissions are set to the specified `perm`
// value. If an error occurs during the file write operation, such as the file
// not being writable or permission issues, the function terminates with a fatal
// log message.
//
// @param path The path to the file to write to.
// @param text The string to write to the file.
// @param perm The file mode (permissions) to set for the file.
// @example WriteTextFile("/tmp/myfile.txt", "Hello, world!", 0644)
func WriteTextFile(path string, text string, perm fs.FileMode) {
	WriteFile(path, []byte(text), perm)
}

// ReplacePattern updates a text file by replacing patterns with specified substitutions.
//
// This function reads a text file specified by `path`, replaces all occurrences of
// the `pattern` with `repl`, and then writes the modified content back to the
// same file. The file's permissions remain unchanged throughout this operation.
// It leverages `ReadFileAsString`, `strings.ReplaceAll`, `WriteTextFile`, and
// `GetFileMode` internally to perform the read, modify, and write operations.
// If any error occurs during these steps, such as file reading or writing errors,
// the function will terminate with a fatal log message.
//
// @param path    The path to the text file to modify.
// @param pattern The string pattern to search for and replace within the file.
// @param repl    The replacement string to substitute the pattern with.
//
// @example
//
//	ReplacePattern("/tmp/myfile.txt", "old_text", "new_text")
//	// Replaces all occurrences of "old_text" with "new_text" in /tmp/myfile.txt.
func ReplacePattern(path string, pattern string, repl string) {
	text := ReadFileAsString(path)
	text = strings.ReplaceAll(text, pattern, repl)
	WriteTextFile(path, text, GetFileMode(path))
}

// AppendToTextFile appends text to a file.
//
// This function appends the given text to a file specified by `path`.
// It can optionally check for the existence of the text in the file before appending.
// If `checkExistence` is true, the function will check if the `text` already exists
// in the file. If it does, the function will return without appending. If `checkExistence`
// is false, the function will directly append the text to the file without checking.
// If the file does not exist, it will be created. If an error occurs during the file
// operation (open, write, close), the function will terminate with a fatal log message.
//
// @param path           The path to the file to append to.
// @param text           The text to append to the file.
// @param checkExistence If true, checks if the text already exists in the file before appending.
//
// @example
//
//	AppendToTextFile("/tmp/myfile.txt", "This is some text to append.", false)
//	// Appends "This is some text to append." to /tmp/myfile.txt.
//	AppendToTextFile("/tmp/myfile.txt", "This is some text to append.", true)
//	// Checks if "This is some text to append." already exists in the file, and only
//	// appends it if it doesn't.
//
// @remarks
//
//		If `checkExistence` is true, the `text` is trimmed using `strings.TrimSpace`
//		before checking for its existence in the file. This ensures that leading and
//		trailing spaces do not prevent a match.
//
//		If the `text` does not exists and it is required to append it,
//	 the file is opened with `os.O_APPEND|os.O_CREATE|os.O_WRONLY` flags and with
//	 a permission mode of `0o644` (rw-r--r--).
//
//	 if the text already exists, it does nothing.
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

// GetBashConfigFile returns the path to the Bash configuration file based on the operating system.
//
// This function returns the path to the Bash configuration file for the current user.
// On Linux, it returns the path to ".bashrc" in the user's home directory.
// On other operating systems, it returns the path to ".bash_profile" in the user's home directory.
//
// @return The path to the Bash configuration file as a string.
//
// @example
//
//	bashConfigFile := GetBashConfigFile()
//	// On Linux, bashConfigFile will be something like "/home/user/.bashrc"
//	// On macOS, it will be something like "/Users/user/.bash_profile"
func GetBashConfigFile() string {
	home := UserHomeDir()
	file := ".bash_profile"
	if runtime.GOOS == "linux" {
		file = ".bashrc"
	}
	return filepath.Join(home, file)
}

// ConfigBash configures the Bash shell environment.
//
// This function performs the following configurations to the Bash shell:
//   - configure the shell's PATH environment variable smartly
//   - set the environment variables VISUAL and EDITOR to nvim with a fallback to vim.
func ConfigBash() {
	bashConfigFile := GetBashConfigFile()
	ConfigShellPath(bashConfigFile)
	AppendToTextFile(
		bashConfigFile,
		`
if which nvim > /dev/null; then
	export VISUAL=nvim
	export EDITOR=nvim
else
	export VISUAL=vim
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
	if GetLinuxDistId() == "idx" {
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
	$(ls -d /opt/*/bin 2> /dev/null)
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

// VirtualMemory retrieves information about the system's virtual memory.
//
// This function retrieves details about the system's virtual memory using the
// `mem.VirtualMemory()` function from the `github.com/shirou/gopsutil/mem`
// package. It returns a pointer to a `mem.VirtualMemoryStat` struct, which
// contains various memory-related statistics like total memory, available
// memory, used memory, free memory, etc.
// If there is an error while retrieving the virtual memory information,
// the function will terminate with a fatal log message.
//
// @return A pointer to a `mem.VirtualMemoryStat` struct representing the system's virtual memory information.
//
// @example
// memInfo := VirtualMemory()
func VirtualMemory() *mem.VirtualMemoryStat {
	memStat, err := mem.VirtualMemory()
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return memStat
}

// CpuInfo retrieves information about the system's CPU.
//
// This function retrieves detailed information about each logical CPU core on the system.
// It utilizes the `cpu.Info()` function from the `github.com/shirou/gopsutil/cpu`
// package to gather this data. The information returned includes details such as the
// model name, number of cores, clock speed, and other CPU-specific attributes.
// If there is an error while retrieving the CPU information, the function will
// terminate with a fatal log message.
//
// @return A slice of `cpu.InfoStat` structs, each representing information about a logical CPU core.
//
// @example
// cpuInfo := CpuInfo()
func CpuInfo() []cpu.InfoStat {
	cpuInfo, err := cpu.Info()
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return cpuInfo
}

// BuildYesFlag constructs a string flag for commands that require confirmation.
//
// This function checks if a Cobra command has the "yes" flag set to true. If it is,
// it returns the string "-y"; otherwise, it returns an empty string. This is commonly
// used to automatically answer "yes" to prompts during command execution.
//
// @param cmd A pointer to a Cobra command object.
//
// @return "-y" if the "yes" flag is set, otherwise an empty string.
//
// @example
//
//	yesFlag := BuildYesFlag(myCmd)
//	// If the --yes flag is present and set to true, yesFlag will be "-y".
func BuildYesFlag(cmd *cobra.Command) string {
	return IfElseString(GetBoolFlag(cmd, "yes"), "-y", "")
}

// BuildPipUninstall constructs a command to uninstall a Python package using pip.
//
// This function creates a command string to uninstall a Python package using pip.
// It takes a Cobra command as input and retrieves the path to the Python executable
// from the "python" flag. It then formats a command that uses pip to uninstall
// the specified package.
//
// @param cmd A pointer to a Cobra command object.
//
// @return A string representing the pip uninstall command.
//
// @example
//
//	uninstallCmd := BuildPipUninstall(myCmd)
//	// If --python is set to "/usr/bin/python3", uninstallCmd will be "/usr/bin/python3 -m pip uninstall"
func BuildPipUninstall(cmd *cobra.Command) string {
	python := GetStringFlag(cmd, "python")
	return Format("{python} -m pip uninstall", map[string]string{
		"python": python,
	})
}

// BuildPipInstall constructs a command to install a Python package using pip.
//
// This function generates a command string to install a Python package using pip.
// It takes a Cobra command as input and retrieves the path to the Python executable
// from the "python" flag. It also handles the optional "--user" flag and the
// "extra-pip-options" flag to allow for user-level installations and additional
// pip options. If the specified Python executable is not found, it returns an
// empty string.
//
// @param cmd A pointer to a Cobra command object.
//
// @return A string representing the pip install command.
//
// @example
//
//	installCmd := BuildPipInstall(myCmd)
//	// If --python is set to "/usr/bin/python3", --user is present,
//	// and --extra-pip-options is set to "index-url=https://example.com/simple,trusted-host=example.com"
//	// installCmd will be:
//	// "PIP_BREAK_SYSTEM_PACKAGES=1 /usr/bin/python3 -m pip install --user --index-url=https://example.com/simple --trusted-host=example.com"
//
// @remarks
//
//		The `PIP_BREAK_SYSTEM_PACKAGES=1` environment variable is set to allow for
//		the installation of packages that might conflict with system packages.
//		This is considered an unsafe practice, be cautious when using this function.
//
//	 The extra-pip-options must be separated by comma.
//
//	 The python executable should be added to the PATH.
//
//	 If the python executable is not found, it returns "".
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

// ExistsCommand checks if a command exists in the system's PATH.
//
// This function uses `exec.LookPath` to check if a given command exists and
// is executable in the system's PATH. It returns true if the command is found,
// and false otherwise. This is useful for checking if external tools or
// programs are available before attempting to execute them.
//
// @param cmd The name of the command to check for.
//
// @return true if the command exists in the PATH, false otherwise.
//
// @example
//
//	if ExistsCommand("ls") {
//		fmt.Println("ls command exists")
//	}
func ExistsCommand(cmd string) bool {
	cmd = NormalizePath(cmd)
	_, err := exec.LookPath(cmd)
	return err == nil
}

// MkdirAll creates a directory and all necessary parent directories.
//
// This function is similar to `os.MkdirAll` but it terminates the program
// with a fatal log message if an error occurs. It ensures that the specified
// directory, along with any necessary parent directories, are created with the
// given permissions.
//
// @param path The path of the directory to create.
// @param perm The file mode (permissions) to set for the newly created directories.
//
// @example
//
//	MkdirAll("/tmp/mydir/subdir", 0o755)
func MkdirAll(path string, perm os.FileMode) {
	err := os.MkdirAll(NormalizePath(path), perm)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
}

// AddPythonFlags adds common Python-related flags to a Cobra command.
//
// This function adds the following flags to the given Cobra command:
//   - `--python`: Specifies the path to the Python3 executable. The default value is "python3".
//   - `--user`: Indicates whether to install Python packages to the user's local directory.
//     By default, this flag is false.
//   - `--extra-pip-options`: Allows specifying extra options (separated by commas) to pass to pip.
//     By default, this flag is empty.
//
// These flags are commonly used when working with Python packages and installations.
//
// @param cmd A pointer to the Cobra command to which the flags will be added.
//
// @example
//
//	AddPythonFlags(myCmd) // Adds the Python-related flags to the 'myCmd' command.
func AddPythonFlags(cmd *cobra.Command) {
	cmd.Flags().String("python", "python3", "Path to the python3 command.")
	cmd.Flags().Bool("user", false, "Install Python packages to user's local directory.")
	cmd.Flags().StringSlice("extra-pip-options", []string{}, "Extra options (separated by comma) to pass to pip.")
}

// IsLinux checks if the current operating system is Linux.
//
// This function determines whether the operating system of the current runtime environment
// is Linux. It uses the `runtime.GOOS` constant to identify the operating system.
// It returns true if the operating system is Linux, and false otherwise.
//
// @return true if the current OS is Linux, false otherwise.
//
// @example
//
//	if IsLinux() {
//	  fmt.Println("Running on Linux")
//	} else {
//	  fmt.Println("Not running on Linux")
//	}
func IsLinux() bool {
	switch runtime.GOOS {
	case "linux":
		return true
	default:
		return false
	}
}

// GetLinuxDistId retrieves the distribution ID of the current Linux system.
//
// This function retrieves the ID of the Linux distribution by reading the `/etc/os-release`
// file, which is a standard way to get the OS release information on Linux systems.
// It uses the `distro.OSRelease()` function to parse the `/etc/os-release` file and
// then looks up the "ID" field to get the distribution ID (e.g., "ubuntu", "debian", "fedora").
// If the "ID" field is not found in the `/etc/os-release` file, it returns an empty string.
//
// @return The distribution ID of the current Linux system, or an empty string if not found.
//
// @example
//
//	distId := GetLinuxDistId()
//	if distId == "ubuntu" {
//	  fmt.Println("Running on Ubuntu")
//	}
func GetLinuxDistId() string {
	m := distro.OSRelease()
	distId, found := m["ID"]
	if found {
		return distId
	}
	return ""
}

// IsUbuntu checks if the current Linux distribution is Ubuntu.
//
// This function determines whether the operating system is Ubuntu by checking
// the distribution ID using `GetLinuxDistId()`. It returns true if the ID is
// "ubuntu" and false otherwise.
//
// @return true if the current OS is Ubuntu, false otherwise.
//
// @example
//
//	if IsUbuntu() {
//	  fmt.Println("Running on Ubuntu")
//	} else {
//	  fmt.Println("Not running on Ubuntu")
//	}
func IsUbuntu() bool {
	return GetLinuxDistId() == "ubuntu"
}

// IsDebian checks if the current Linux distribution is Debian.
//
// This function determines whether the operating system is Debian by checking
// the distribution ID using `GetLinuxDistId()`. It returns true if the ID is
// "debian" and false otherwise.
//
// @return true if the current OS is Debian, false otherwise.
//
// @example
//
//	if IsDebian() {
//	  fmt.Println("Running on Debian")
//	} else {
//	  fmt.Println("Not running on Debian")
//	}
func IsDebian() bool {
	return GetLinuxDistId() == "debian"
}

// IsDebianSeries checks if the current Linux distribution belongs to the Debian series.
//
// This function determines whether the current Linux distribution is part of the
// Debian series, which includes Debian, Antix, and LMDE (Linux Mint Debian Edition).
// It uses the `GetLinuxDistId()` function to get the distribution ID and then
// checks if it matches any of the IDs in the Debian series.
//
// @return true if the current OS is part of the Debian series, false otherwise.
//
// @example
//
//	if IsDebianSeries() {
//	  fmt.Println("Running a Debian-based distribution")
//	} else {
//	  fmt.Println("Not running a Debian-based distribution")
//	}
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

// IsDebianUbuntuSeries checks if the current Linux distribution belongs to the Debian or Ubuntu series.
//
// This function determines whether the current Linux distribution is part of either the
// Debian or Ubuntu series, which includes Debian, Antix, LMDE (Linux Mint Debian Edition),
// Ubuntu, Linux Mint, and Pop!_OS.
// It uses the `GetLinuxDistId()` function to get the distribution ID and then checks if
// it matches any of the IDs in either the Debian or Ubuntu series.
//
// @return true if the current OS is part of the Debian or Ubuntu series, false otherwise.
//
// @example
//
//	if IsDebianUbuntuSeries() {
//	  fmt.Println("Running a Debian or Ubuntu-based distribution")
//	} else {
//	  fmt.Println("Not running a Debian or Ubuntu-based distribution")
//	}
//
// @remarks:
// LinuxMint and Pop!_OS are based on Ubuntu.
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

// IsUbuntuSeries checks if the current Linux distribution belongs to the Ubuntu series.
//
// This function determines whether the current Linux distribution is part of the
// Ubuntu series, which includes Ubuntu, Linux Mint, and Pop!_OS. It uses the
// `GetLinuxDistId()` function to get the distribution ID and then checks if
// it matches any of the IDs in the Ubuntu series.
//
// @return true if the current OS is part of the Ubuntu series, false otherwise.
//
// @example
//
//	if IsUbuntuSeries() {
//	  fmt.Println("Running an Ubuntu-based distribution")
//	} else {
//	  fmt.Println("Not running an Ubuntu-based distribution")
//	}
//
// @remarks:
// LinuxMint and Pop!_OS are based on Ubuntu.

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

// IsFedoraSeries checks if the current Linux distribution belongs to the Fedora series.
//
// This function determines whether the current Linux distribution is part of the
// Fedora series, which includes Fedora, CentOS, and RHEL (Red Hat Enterprise Linux).
// It uses the `GetLinuxDistId()` function to get the distribution ID and then checks
// if it matches any of the IDs in the Fedora series.
//
// @return true if the current OS is part of the Fedora series, false otherwise.
//
// @example
//
//	if IsFedoraSeries() {
//	  fmt.Println("Running a Fedora-based distribution")
//	} else {
//	  fmt.Println("Not running a Fedora-based distribution")
//	}
//
// @remarks:
// CentOS and RHEL are based on Fedora.

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

// IfElseString returns one of two strings based on a boolean condition.
//
// This function provides a simple way to choose between two string values based
// on the value of a boolean variable. If the boolean `b` is true, it returns
// the string `t`; otherwise, it returns the string `f`. This is similar to a
// ternary operator in other languages.
//
// @param b The boolean condition to evaluate.
// @param t The string to return if `b` is true.
// @param f The string to return if `b` is false.
//
// @return `t` if `b` is true, `f` otherwise.
//
// @example
//
//	result := IfElseString(x > 5, "greater", "less or equal")
//	// If x is greater than 5, result will be "greater"; otherwise, it will be "less or equal".
//

func IfElseString(b bool, t string, f string) string {
	if b {
		return t
	}
	return f
}

// Using Homebrew to install packages
//
// This function installs a list of packages using Homebrew. It attempts to
// install each package and, if the installation fails (e.g., because the
// package is already installed), it attempts to link the package with the
// `--overwrite` and `--force` options to ensure it is properly configured.
//
// The function iterates through each package in the provided `pkgs` slice.
// For each package, it formats a Homebrew command string that first attempts
// to install the package with the `--force` option, and if that fails, attempts
// to link the package with `--overwrite` and `--force`.
//
// If any errors occur during the execution of the Homebrew commands,
// the function will terminate with a fatal log message.
//
// @param pkgs A slice of strings representing the packages to install.
//
// @example BrewInstallSafe([]string{"git", "zsh"})
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

// IsSocket checks if a file is a socket.
//
// This function determines whether the file located at the given path is a socket.
// It uses the os.Stat function to get file information and checks if the file's mode type
// is a socket by comparing it with fs.ModeSocket. If there is an error during the
// file stat operation, the function will terminate with a fatal log message.
//
// @param path The path to the file to check.
//
// @return true if the file is a socket, false otherwise.
//
// @example
//
//	if IsSocket("/var/run/docker.sock") { ... }
//
// Check if a file is a socket.
func IsSocket(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return fileInfo.Mode().Type() == fs.ModeSocket
}

// Symlink is a wrapper of os.Symlink with error handling.
//
// @param path The path to the source file/directory.
// @param dstLink The path where the symbolic link will be created.
//
// @example
//
//	Symlink("/path/to/source/file.txt", "/path/to/link/file.txt")
func Symlink(path string, dstLink string, backup bool) {
	path = NormalizePath(path)
	dstLink = NormalizePath(dstLink)
	if backup {
		Backup(dstLink, "")
	} else {
		RemoveAll(dstLink)
	}

	MkdirAll(filepath.Dir(dstLink), 0o700)
	err := os.Symlink(path, dstLink)
	if err != nil {
		log.Fatalf("Failed to link the file %s to %s!\n", path, dstLink)
	}
}

func SymlinkIntoDir(path string, dstDir string, backup bool) {
	Symlink(path, filepath.Join(dstDir, filepath.Base(path)), backup)
}

// Update map1 using map2.
//
// This function updates the contents of `map1` with the contents of `map2`.
// It iterates through the keys of `map2` and checks if each key exists in `map1`.
//
// If a key from `map2` does not exist in `map1`, it's added with its corresponding value.
//
// If a key from `map2` exists in `map1`, the behavior depends on the type of the
// associated values:
//   - If both values are of type `orderedmap.OrderedMap[string, any]`, then this
//     function recursively calls itself to merge the two inner maps.
//   - If the value in `map2` is a `orderedmap.OrderedMap[string, any]` but the
//     value in `map1` is not, the value in `map1` is replaced with the one from `map2`.
//   - If the value in `map2` is not a `orderedmap.OrderedMap[string, any]`, the value
//     in `map1` is replaced with the value from `map2`.
//
// This function is designed to deeply merge two `orderedmap.OrderedMap[string, any]`
// objects, with `map2` taking precedence when there are conflicts.
//
// @param map1 The target map to be updated.
// @param map2 The source map from which to take updates.
//
// @example
//
//	map1 := orderedmap.NewOrderedMap[string, any]()
//	map1.Set("a", 1)
//	map1.Set("b", 2)
//	map2 := orderedmap.NewOrderedMap[string, any]()
//	map2.Set("b", 3)
//	map2.Set("c", 4)
//	UpdateMap(map1, map2)
//	// Now map1 is {a: 1, b: 3, c: 4}
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

func HostInfo() *host.InfoStat {
	info, err := host.Info()
	if err != nil {
		log.Fatal(err)
	}
	return info
}

func HostKernelArch() string {
	switch HostInfo().KernelArch {
	case "x86_64", "amd64":
		return "amd64"
	case "arm64", "aarch64":
		return "arm64"
	default:
		return "_other"
	}
}

// BuildKernelOSKeywords constructs a list of keywords based on kernel architecture and operating system.
//
// @param keywords A map where keys are keyword categories and values are lists of keywords.
//
// @return A slice of strings representing the combined list of keywords.
//
// @example
//
//	keywords := map[string][]string{
//		"common":             {"keyword1", "keyword2"},
//		"amd64":             {"amd64_keyword"},
//		"arm64":              {"arm64_keyword"},
//		"darwin":             {"darwin_keyword"},
//		"linux":              {"linux_keyword"},
//		"DebianUbuntuSeries": {"debian_ubuntu_keyword"},
//		"FedoraSeries":       {"fedora_keyword"},
//		"OtherLinux":         {"other_linux_keyword"},
//	}
//	result := BuildKernelOSKeywords(keywords)
//	// result might contain a combination of the above keywords based on the OS and architecture

func BuildKernelOSKeywords(keywords map[string][]string) []string {
	kwds := keywords["common"]
	k, found := keywords[HostKernelArch()]
	if found {
		kwds = append(kwds, k...)
	}
	k, found = keywords[runtime.GOOS]
	if found {
		kwds = append(kwds, k...)
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
	return kwds
}

// ParseInt converts a string to an int64.
//
// This function converts the input string `str` to an int64. It assumes that
// the string represents a base-10 integer. If the string cannot be parsed as
// an int64, the function terminates with a fatal log message.
//
// @param str The string to convert to an int64.
//
// @return The int64 representation of the string.
//
// @example
//
//	num := ParseInt("12345")
//	// num is now the int64 12345.
//	num := ParseInt("invalid")
//	// This will result in a fatal log message.
//
// @remarks
// This function terminates with a log.Fatal if the string conversion fails.
func ParseInt(str string) int64 {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		log.Fatalf("Error converting string to int64: %v\n", err)
	}
	return i
}

func Rename(original string, new string) {
	err := os.Rename(original, new)
	if err != nil {
		log.Fatal(err)
	}
}

func Backup(original string, backup string) {
	original = NormalizePath(original)
	backup = NormalizePath(backup)
	if ExistsDir(original) {
		if backup == "" {
			backup = filepath.Clean(original) + "_" + time.Now().Format(time.RFC3339)
		}
		Rename(original, backup)
		fmt.Printf("%s has been backed up to %s.\n", original, backup)
	}
}
