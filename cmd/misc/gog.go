package misc

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
)

// installGog installs the prebuilt release for the host OS and architecture.
func installGog(cmd *cobra.Command) {
	tmpdir := utils.CreateTempDir("")
	defer os.RemoveAll(tmpdir)
	archive := filepath.Join(tmpdir, "gog.tar.gz")
	network.DownloadGitHubRelease("openclaw/gogcli", utils.GetStringFlag(cmd, "version"), map[string][]string{
		"common": {"gogcli_", ".tar.gz"},
		"amd64":  {"_amd64"},
		"arm64":  {"_arm64"},
		"linux":  {"_linux_"},
		"darwin": {"_darwin_"},
	}, nil, archive)
	dirBin := utils.NormalizePath(utils.GetStringFlag(cmd, "bin-dir"))
	utils.RunCmd(utils.Format(`tar -xzf "$ICON_GOG_ARCHIVE" -C "$ICON_GOG_TMP_DIR" \
		&& {prefix} mkdir -p "$ICON_GOG_BIN_DIR" \
		&& {prefix} install -m 755 "$ICON_GOG_TMP_DIR/gog" "$ICON_GOG_BIN_DIR/gog"`, map[string]string{
		"prefix": utils.GetCommandPrefix(utils.GetBoolFlag(cmd, "sudo"), map[string]uint32{
			dirBin: unix.W_OK | unix.R_OK,
		}),
	}), "ICON_GOG_ARCHIVE="+archive, "ICON_GOG_TMP_DIR="+tmpdir, "ICON_GOG_BIN_DIR="+dirBin)
}

// copyGogConfig uses native file operations because macOS config paths contain spaces.
// The file is copied rather than symlinked since gog rewrites its settings.
func copyGogConfig(src, dst string, backup bool) {
	const dirMode = 0o700
	const fileMode = 0o600
	data := utils.ReadFile(src)
	if err := os.MkdirAll(filepath.Dir(dst), dirMode); err != nil {
		log.Fatal(err)
	}
	if _, err := os.Lstat(dst); err == nil {
		if backup {
			err = os.Rename(dst, dst+"_"+time.Now().Format(time.RFC3339Nano))
		} else {
			err = os.Remove(dst)
		}
		if err != nil {
			log.Fatal(err)
		}
	} else if !os.IsNotExist(err) {
		log.Fatal(err)
	}
	if err := os.WriteFile(dst, data, fileMode); err != nil {
		log.Fatal(err)
	}
}

// configGog copies only the settings file; credentials and other mutable state stay local.
func configGog(cmd *cobra.Command) {
	icon.FetchConfigData(false, "")
	src := "~/.config/icon-data/gog/config.json"
	if !utils.ExistsFile(src) {
		log.Fatalf("The gog configuration file %s does not exist.", src)
	}
	// Ask gog to resolve GOG_CONFIG_DIR, GOG_HOME, XDG and platform defaults.
	executable := "gog"
	if utils.GetBoolFlag(cmd, "install") {
		var err error
		executable, err = filepath.Abs(filepath.Join(utils.NormalizePath(utils.GetStringFlag(cmd, "bin-dir")), "gog"))
		if err != nil {
			log.Fatal(err)
		}
	}
	output, err := exec.CommandContext(context.Background(), executable, "config", "path", "--json=false").Output()
	if err != nil {
		log.Fatalf("Cannot determine the gog configuration path: %v", err)
	}
	dst := strings.TrimSpace(string(output))
	if !filepath.IsAbs(dst) {
		log.Fatalf("Invalid gog configuration path: %q", dst)
	}
	copyGogConfig(src, dst, utils.ShouldBackup(cmd))
	log.Println("Authorize accounts separately with gog auth; provide GOG_KEYRING_PASSWORD for unattended file-keyring use.")
}

// uninstallGog removes only the executable, using rip when available.
func uninstallGog(cmd *cobra.Command) {
	binary := filepath.Join(utils.NormalizePath(utils.GetStringFlag(cmd, "bin-dir")), "gog")
	info, err := os.Lstat(binary)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		log.Fatal(err)
	}
	remove := utils.LookPath("rip")
	if remove == "" || info.Mode()&os.ModeSymlink != 0 {
		remove = "rm"
	}
	utils.RunCmd(utils.Format(`{prefix} "$ICON_GOG_REMOVE" -- "$ICON_GOG_BINARY"`, map[string]string{
		"prefix": utils.GetCommandPrefix(false, map[string]uint32{
			filepath.Dir(binary): unix.W_OK | unix.X_OK,
		}),
	}), "ICON_GOG_REMOVE="+remove, "ICON_GOG_BINARY="+binary)
}

// Install and configure the Google Workspace CLI.
func gog(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		installGog(cmd)
	}
	if utils.GetBoolFlag(cmd, "config") {
		configGog(cmd)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		uninstallGog(cmd)
	}
}

var gogCmd = &cobra.Command{
	Use:     "gog",
	Aliases: []string{"gogcli"},
	Short:   "Install and configure the Google Workspace CLI.",
	Run:     gog,
}

func ConfigGogCmd(rootCmd *cobra.Command) {
	gogCmd.Flags().BoolP("install", "i", false, "Install gog.")
	gogCmd.Flags().BoolP("config", "c", false, "Copy gog settings from icon-data (account authorization is separate).")
	gogCmd.Flags().BoolP("uninstall", "u", false, "Uninstall the gog executable, preserving settings and credentials.")
	gogCmd.Flags().Bool("sudo", false, "Force using sudo for installation.")
	gogCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	gogCmd.Flags().String("bin-dir", "/usr/local/bin", "The directory for installing the gog executable.")
	gogCmd.Flags().StringP("version", "v", "", "The gog release version constraint (latest by default).")
	rootCmd.AddCommand(gogCmd)
}
