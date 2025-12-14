package utils

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// GetFileMode retrieves the file mode (permissions) of a given file.
//
// This function uses os.Stat to get file information and returns the file mode.
// If an error occurs during the file stat, the function will terminate with a fatal log.
//
// @param file The path to the file.
// @return The file mode (fs.FileMode).
func getFileMode(file string) fs.FileMode {
	fileInfo, err := os.Stat(file)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return fileInfo.Mode()
}

func dir(path string) string {
	return filepath.Dir(NormalizePath(path))
}

// CopyFileToDir copies a file from a source path to a destination directory.
//
// It constructs the destination file path by joining the destination directory
// with the base name of the source file. Then it calls the copyFile function to perform the actual copy.
//
// @param sourceFile      The path to the source file.
// @param destinationDir The path to the destination directory where the source file will be copied.
func CopyFileToDir(sourceFile string, destinationDir string) {
	sourceFile = NormalizePath(sourceFile)
	destinationDir = NormalizePath(destinationDir)
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
	MkdirAll(destinationDir, "")
	for _, entry := range ReadDir(sourceDir) {
		if entry.IsDir() {
			srcDir := filepath.Join(sourceDir, entry.Name())
			dstDir := filepath.Join(destinationDir, entry.Name())
			CopyDir(srcDir, dstDir)
		} else {
			sourceFile := filepath.Join(sourceDir, entry.Name())
			CopyFile(sourceFile, filepath.Join(destinationDir, entry.Name()))
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
	WriteTextFile(path, text, getFileMode(path))
}

// ExistsPath checks if a file or directory exists at the specified path.
//
// @param path The path to the file or directory.
//
// @return true if the file or directory exists, false otherwise.
func ExistsPath(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// ExistsDir checks if a directory exists at the specified path.
//
// @param path The path to check for a directory.
//
// @return true if a directory exists at the path, false otherwise.
func ExistsDir(path string) bool {
	stat, err := os.Stat(NormalizePath(path))
	if os.IsNotExist(err) {
		return false
	}
	return stat.IsDir()
}

// ExistsFile checks if a file exists at the specified path.
//
// @param path The path to the file.
//
// @return true if a file exists at the path, false otherwise.
func ExistsFile(path string) bool {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !stat.IsDir()
}

// Getwd returns the current working directory.
//
// @return The current working directory as a string.
func Getwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return cwd
}

// UserHomeDir returns the current user's home directory.
//
// @return The path to the current user's home directory as a string.
func UserHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return home
}

// ReadDir reads the named directory and returns
// a list of directory entries sorted by filename.
// If an error occurs during the directory reading, the function
// will terminate with a fatal log.
//
// @param dir The name of the directory to read.
//
// @return A slice of DirEntry representing the directory's contents.
func ReadDir(dir string) []os.DirEntry {
	dir = NormalizePath(dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return entries
}

// ReadAllAsText reads all data from an io.ReadCloser and returns it as a string.
//
// @param readCloser The io.ReadCloser from which to read the data.
//
// @return The read data as a string.
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
func ReadFile(path string) []byte {
	bytes, err := os.ReadFile(NormalizePath(path))
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return bytes
}

// ReadFileAsString reads the entire content of a file and returns it as a string.
//
// @param path The path to the file to be read.
func ReadFileAsString(path string) string {
	return string(ReadFile(path))
}

// WriteFile writes a byte slice to a file with the specified permissions.
//
// @param fileName The name of the file to write to.
// @param data     The byte slice to write to the file.
// @param perm     The file mode (permissions) to set for the file.
func WriteFile(fileName string, data []byte, perm fs.FileMode) {
	fileName = NormalizePath(fileName)
	err := os.WriteFile(fileName, data, perm)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
}

// WriteTextFile writes a string to a file with the specified permissions.
//
// @param path The path to the file to write to.
// @param text The string to write to the file.
// @param perm The file mode (permissions) to set for the file.
func WriteTextFile(path string, text string, perm fs.FileMode) {
	WriteFile(path, []byte(text), perm)
}

// AppendToTextFile appends text to a file.
//
// @param path           The path to the file to append to.
// @param text           The text to append to the file.
// @param checkExistence If true, checks if the text already exists in the file before appending.
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
