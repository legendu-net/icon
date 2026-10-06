package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// Chmod changes the mode of the named file to mode.
func Chmod(path, mode string) {
	path = NormalizePath(path)
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
	sourceFile = NormalizePath(sourceFile)
	destinationFile = NormalizePath(destinationFile)
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
// It uses `rip` (rm-improved) when available, otherwise falls back to `rm -rf`.
//
// @param path The path to the file or directory to remove.
func RemoveAll(path string) {
	path = NormalizePath(path)
	// Unlike `rm -rf`, `rip` errors out on a non-existent path, so skip removal
	// when the path is definitively absent. Lstat (not Stat) is used so a broken
	// symlink still counts as present and gets removed; a permission error is not
	// treated as absent, leaving the (possibly sudo'd) removal below to handle it.
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return
	}
	// Removing a path modifies its parent directory, so the parent's permissions matter too.
	prefix := GetCommandPrefix(false, map[string]uint32{
		path:                               unix.W_OK | unix.R_OK,
		filepath.Dir(filepath.Clean(path)): unix.W_OK | unix.X_OK,
	})
	var cmd string
	// Resolve `rip` to its absolute path. `rip` is commonly installed under the
	// user's home (e.g. ~/.cargo/bin), which is not on sudo's secure_path, so the
	// bare name would fail under a sudo'd removal; the absolute path always works.
	if ripPath := LookPath("rip"); ripPath != "" {
		cmd = Format("{prefix} {rip} {path}", map[string]string{
			"prefix": prefix,
			"rip":    ripPath,
			"path":   path,
		})
	} else {
		cmd = Format("{prefix} rm -rf {path}", map[string]string{
			"prefix": prefix,
			"path":   path,
		})
	}
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

// BackupOrRemove backs up the path if backup is true, otherwise removes it.
// Call this before copying or symlinking to prepare the destination.
func BackupOrRemove(path string, backup bool) {
	path = NormalizePath(path)
	if backup {
		Backup(path, "")
	} else {
		RemoveAll(path)
	}
}

// Symlink is a wrapper of os.Symlink with error handling.
//
// @param path The path to the source file/directory.
// @param dstLink The path where the symbolic link will be created.
func Symlink(path, dstLink string) {
	path = NormalizePath(path)
	dstLink = NormalizePath(dstLink)
	MkdirAll(filepath.Dir(dstLink), "")
	prefix := GetCommandPrefix(false, map[string]uint32{
		path:    unix.R_OK,
		dstLink: unix.W_OK | unix.R_OK,
	})
	cmd := Format("{prefix} ln -sv {path} {dstLink}", map[string]string{
		"prefix":  prefix,
		"path":    path,
		"dstLink": dstLink,
	})
	RunCmd(cmd)
}

func SymlinkIntoDir(path, dstDir string) {
	Symlink(path, filepath.Join(dstDir, filepath.Base(path)))
}

// CopyOrSymlink copies src to dst (using CopyFile or CopyDirRegular) when doCopy
// is true, otherwise creates a symlink at dst pointing to src.
func CopyOrSymlink(src, dst string, doCopy bool) {
	src = NormalizePath(src)
	dst = NormalizePath(dst)
	if !ExistsPath(src) {
		log.Fatalf("ERROR - the source path %s does not exist.", src)
	}
	if doCopy {
		if ExistsDir(src) {
			CopyDirRegular(src, dst)
		} else {
			CopyFile(src, dst)
		}
	} else {
		Symlink(src, dst)
	}
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

// DefaultBackupsToKeep is the number of most recent (physical) backups kept per path.
const DefaultBackupsToKeep = 3

// backupSuffix matches the "_<RFC3339 timestamp>" suffix that Backup appends to a path.
var backupSuffix = regexp.MustCompile(
	`_(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:Z|[+-]\d{2}:\d{2}))$`)

// ParseBackupPath splits a backup path created by Backup into the original path
// and the timestamp of the backup.
//
// @param path The path to parse.
//
// @return The original path, the timestamp and whether path is a backup path.
func ParseBackupPath(path string) (string, time.Time, bool) {
	path = filepath.Clean(path)
	match := backupSuffix.FindStringSubmatchIndex(path)
	if match == nil {
		return "", time.Time{}, false
	}
	ts, err := time.Parse(time.RFC3339, path[match[2]:match[3]])
	if err != nil || match[0] == 0 || path[match[0]-1] == filepath.Separator {
		return "", time.Time{}, false
	}
	return path[:match[0]], ts, true
}

// ListBackups lists backups (created by Backup) of a path, newest first.
//
// @param original The path whose backups to list.
func ListBackups(original string) []string {
	original = filepath.Clean(NormalizePath(original))
	entries, err := os.ReadDir(filepath.Dir(original))
	if err != nil {
		return nil
	}
	type backup struct {
		path string
		ts   time.Time
	}
	var backups []backup
	for _, entry := range entries {
		path := filepath.Join(filepath.Dir(original), entry.Name())
		if orig, ts, ok := ParseBackupPath(path); ok && orig == original {
			backups = append(backups, backup{path, ts})
		}
	}
	sort.SliceStable(backups, func(i, j int) bool {
		return backups[i].ts.After(backups[j].ts)
	})
	paths := make([]string, len(backups))
	for i, b := range backups {
		paths[i] = b.path
	}
	return paths
}

// PruneBackups removes old backups (created by Backup) of a path.
// Backups which are symbolic links are always removed as they preserve nothing,
// while the most recent keep physical backups are kept.
//
// @param original The path whose backups to prune.
// @param keep     The number of most recent physical backups to keep.
// @param dryRun   If true, only print backups to be removed without removing them.
//
// @return Backups which are removed (or would be removed if dryRun is true).
func PruneBackups(original string, keep int, dryRun bool) []string {
	var removed []string
	for _, path := range ListBackups(original) {
		info, err := os.Lstat(path)
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink == 0 && keep > 0 {
			keep--
			continue
		}
		if dryRun {
			fmt.Printf("Would remove the backup %s.\n", path)
		} else {
			RemoveAll(path)
			fmt.Printf("The backup %s has been removed.\n", path)
		}
		removed = append(removed, path)
	}
	return removed
}

// Backup renames a path to a backup path.
// If backup is empty, the backup path is the original path suffixed with "_<RFC3339 timestamp>",
// and old backups of the path are pruned to keep the most recent DefaultBackupsToKeep ones.
// A symbolic link is removed instead of being backed up, as it preserves nothing.
//
// @param original The path to back up.
// @param backup   The backup path, or empty to use a timestamped one.
func Backup(original, backup string) {
	original = NormalizePath(original)
	backup = NormalizePath(backup)
	if backup != "" {
		if ExistsPath(original) {
			Rename(original, backup)
		}
		return
	}
	if info, err := os.Lstat(original); err == nil && info.Mode()&os.ModeSymlink != 0 {
		RemoveAll(original)
		fmt.Printf("The symbolic link %s has been removed instead of being backed up.\n", original)
		return
	}
	if ExistsPath(original) {
		backup = filepath.Clean(original) + "_" + time.Now().Format(time.RFC3339)
		Rename(original, backup)
		PruneBackups(original, DefaultBackupsToKeep, false)
	}
}

// AppendToTextFile appends text to a file.
//
// @param path           The path to the file to append to.
// @param text           The text to append to the file.
// @param checkExistence If true, checks if the text already exists in the file before appending.
func AppendToTextFile(path, text string, checkExistence bool) {
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
	prefix := GetCommandPrefix(false, map[string]uint32{
		path: unix.R_OK | unix.W_OK,
	})
	cmd := Format(`echo -e {text} | {prefix} tee -a {path} > /dev/null`, map[string]string{
		"text":   text,
		"prefix": prefix,
		"path":   path,
	})
	RunCmd(cmd)
}
