package icon

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// maxCleanDepth is the max depth (relative to a scanned dir) of backups to look for,
// e.g., ~/.local/share/gopass/stores/root_<timestamp> is at depth 5 relative to ~.
const maxCleanDepth = 5

// findBackedUpPaths finds paths under dir which have backups created by utils.Backup.
// Only paths which still exist are returned, as utils.Backup is always followed by
// putting something back to the original path. This avoids touching timestamped files
// which are not created by icon.
//
// @param dir  The directory to scan.
// @param seen Original paths already found (e.g., in other dirs), which are skipped.
//
// @return Original paths (not in seen) which have backups.
func findBackedUpPaths(dir string, seen map[string]bool) []string {
	dir = filepath.Clean(utils.NormalizePath(dir))
	cacheDir := utils.NormalizePath("~/.cache")
	var originals []string
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if path == dir {
				return err
			}
			// Skip unreadable paths instead of aborting the whole scan.
			if entry != nil && entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if path == dir {
			return nil
		}
		if original, _, ok := utils.ParseBackupPath(path); ok {
			if _, err := os.Lstat(original); err == nil && !seen[original] {
				seen[original] = true
				originals = append(originals, original)
			}
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			rel, _ := filepath.Rel(dir, path)
			if path == cacheDir || strings.Count(rel, string(filepath.Separator))+1 >= maxCleanDepth {
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return originals
}

// Clean up old backups of configuration files created by icon.
func clean(cmd *cobra.Command, _ []string) {
	keep := utils.GetIntFlag(cmd, "keep")
	if keep < 0 {
		log.Fatalf("ERROR - the number of backups to keep must be non-negative, got %d.", keep)
	}
	dryRun := utils.GetBoolFlag(cmd, "dry-run")
	seen := map[string]bool{}
	count := 0
	for _, dir := range utils.GetStringSliceFlag(cmd, "dir") {
		for _, original := range findBackedUpPaths(dir, seen) {
			count += len(utils.PruneBackups(original, keep, dryRun))
		}
	}
	if dryRun {
		fmt.Printf("%d backups would be removed.\n", count)
	} else {
		fmt.Printf("%d backups have been removed.\n", count)
	}
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up old backups of configuration files created by icon.",
	Long: `Clean up old backups of configuration files created by icon.

Any path named <path>_<RFC3339 timestamp> whose <path> still exists is treated as a backup.
Backups which are symbolic links are always removed,
while the most recent --keep physical backups of each path are kept.
Run with --dry-run first to review backups to be removed.`,
	Run: clean,
}

func ConfigCleanCmd(rootCmd *cobra.Command) {
	cleanCmd.Flags().IntP("keep", "n", utils.DefaultBackupsToKeep,
		"The number of most recent (physical) backups to keep per path.")
	cleanCmd.Flags().Bool("dry-run", false, "Print backups to be removed without removing them.")
	cleanCmd.Flags().StringSliceP("dir", "d", []string{"~"}, "Directories to scan for backups.")
	rootCmd.AddCommand(cleanCmd)
}
