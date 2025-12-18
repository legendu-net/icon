package utils

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// Chmod changes the mode of the named file to mode.
func Chmod(path, mode string) {
	prefix := GetCommandPrefix(false, map[string]uint32{
		path: unix.W_OK | unix.R_OK,
	})
	cmd := Format("{prefix} chmod -R {mode} {path}", map[string]string{
		"prefix": prefix,
		"mode":   mode,
		"path":   path,
	})
	RunCmd(cmd)
}

// Chmod600 recursively changes file modes of files under a directory to 600.
//
// @param path The path to the file or directory.
func Chmod600(path string) {
	if ExistsDir(path) {
		Chmod(path, "700")
		for _, entry := range ReadDir(path) {
			Chmod600(filepath.Join(path, entry.Name()))
		}
	} else {
		Chmod(path, "600")
	}
}

// copyFile copies a file from the source path to the destination path.
//
// @param sourceFile      The path to the source file.
// @param destinationFile The path to the destination file where the source file will be copied.
func CopyFile(sourceFile, destinationFile string) {
	MkdirAll(dir(destinationFile), "")

	prefix := GetCommandPrefix(false, map[string]uint32{
		sourceFile:      unix.R_OK,
		destinationFile: unix.R_OK | unix.W_OK,
	})
	cmd := Format("{prefix} cp {sourceFile} {destinationFile}", map[string]string{
		"prefix":          prefix,
		"sourceFile":      sourceFile,
		"destinationFile": destinationFile,
	})
	RunCmd(cmd)
	log.Printf("%s is copied to %s.\n", sourceFile, destinationFile)
}

// RemoveAll removes the specified path and any children it contains.
//
// @param path The path to the file or directory to remove.
func RemoveAll(path string) {
	prefix := GetCommandPrefix(false, map[string]uint32{
		path: unix.W_OK | unix.R_OK,
	})
	cmd := Format("{prefix} rm -rf {path}", map[string]string{
		"prefix": prefix,
		"path":   path,
	})
	RunCmd(cmd)
}

// MkdirAll creates a directory and all necessary parent directories.
//
// @param path The path of the directory to create.
// @param perm The file mode (permissions) to set for the newly created directories.
func MkdirAll(path, perm string) {
	perm = strings.TrimSpace(perm)
	path = NormalizePath(path)
	prefix := GetCommandPrefix(false, map[string]uint32{
		path: unix.R_OK | unix.W_OK | unix.X_OK,
	})
	cmd := "{prefix} mkdir -p {path}"
	if perm != "" {
		cmd += " && {prefix} chmod -R {perm} {path}"
	}
	cmd = Format(cmd, map[string]string{
		"prefix": prefix,
		"path":   path,
		"perm":   perm,
	})
	RunCmd(cmd)
}

// Symlink is a wrapper of os.Symlink with error handling.
//
// @param path The path to the source file/directory.
// @param dstLink The path where the symbolic link will be created.
func Symlink(path, dstLink string, backup, copyPath bool) {
	path = NormalizePath(path)
	dstLink = NormalizePath(dstLink)
	if backup {
		Backup(dstLink, "")
	} else {
		RemoveAll(dstLink)
	}

	MkdirAll(filepath.Dir(dstLink), "")
	prefix := GetCommandPrefix(false, map[string]uint32{
		path:    unix.R_OK,
		dstLink: unix.W_OK | unix.R_OK,
	})
	if copyPath {
		cmd := Format("{prefix} cp -ir {path} {dstLink}", map[string]string{
			"prefix":  prefix,
			"path":    path,
			"dstLink": dstLink,
		})
		RunCmd(cmd)
	} else {
		cmd := Format("{prefix} ln -sv {path} {dstLink}", map[string]string{
			"prefix":  prefix,
			"path":    path,
			"dstLink": dstLink,
		})
		RunCmd(cmd)
	}
}

func SymlinkIntoDir(path, dstDir string, backup, copyPath bool) {
	Symlink(path, filepath.Join(dstDir, filepath.Base(path)), backup, copyPath)
}

func Rename(originalPath, newPath string) {
	prefix := GetCommandPrefix(false, map[string]uint32{
		originalPath: unix.W_OK | unix.R_OK,
		newPath:      unix.W_OK | unix.R_OK,
	})
	cmd := Format("{prefix} mv {originalPath} {newPath}", map[string]string{
		"prefix":       prefix,
		"originalPath": originalPath,
		"newPath":      newPath,
	})
	RunCmd(cmd)
	fmt.Printf("The path %s has been renamed to %s.\n", originalPath, newPath)
}

func Backup(original, backup string) {
	original = NormalizePath(original)
	backup = NormalizePath(backup)
	if ExistsPath(original) {
		if backup == "" {
			backup = filepath.Clean(original) + "_" + time.Now().Format(time.RFC3339)
		}
		Rename(original, backup)
	}
}
