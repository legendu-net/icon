package utils

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

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
	})
	RunCmd(cmd)
}

// Symlink is a wrapper of os.Symlink with error handling.
//
// @param path The path to the source file/directory.
// @param dstLink The path where the symbolic link will be created.
func Symlink(path string, dstLink string, backup bool, copy bool) {
	path = NormalizePath(path)
	dstLink = NormalizePath(dstLink)
	if backup {
		Backup(dstLink, "")
	} else {
		RemoveAll(dstLink)
	}

	MkdirAll(filepath.Dir(dstLink), "700")
	prefix := GetCommandPrefix(false, map[string]uint32{
		path:    unix.R_OK,
		dstLink: unix.W_OK | unix.R_OK,
	})
	if copy {
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

func SymlinkIntoDir(path string, dstDir string, backup bool, copy bool) {
	Symlink(path, filepath.Join(dstDir, filepath.Base(path)), backup, copy)
}

func Rename(original string, new string) {
	prefix := GetCommandPrefix(false, map[string]uint32{
		original: unix.W_OK | unix.R_OK,
		new:      unix.W_OK | unix.R_OK,
	})
	cmd := Format("{prefix} mv {original} {new}", map[string]string{
		"prefix":   prefix,
		"original": original,
		"new":      new,
	})
	RunCmd(cmd)
	fmt.Printf("The path %s has been renamed to %s.\n", original, new)
}

func Backup(original string, backup string) {
	original = NormalizePath(original)
	backup = NormalizePath(backup)
	if ExistsPath(original) {
		if backup == "" {
			backup = filepath.Clean(original) + "_" + time.Now().Format(time.RFC3339)
		}
		Rename(original, backup)
	}
}
