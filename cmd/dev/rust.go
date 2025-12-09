package dev

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
)

func linkRust(cmd *cobra.Command, cargoHome string) {
	linkToDir := utils.GetStringFlag(cmd, "link-to-dir")
	if linkToDir == "" {
		return
	}
	cargoBin := filepath.Join(cargoHome, "bin")
	switch runtime.GOOS {
	case "linux", "darwin":
		prefix := utils.GetCommandPrefix(false, map[string]uint32{
			cargoBin:  unix.R_OK,
			linkToDir: unix.W_OK | unix.R_OK,
		})
		// TODO:
		for _, entry := range utils.ReadDir(cargoBin) {
			bin := filepath.Join(cargoBin, entry.Name())
			utils.Format("{prefix} ln -svf {bin} {linkToDir}/", map[string]string{
				"prefix":    prefix,
				"bin":       bin,
				"linkToDir": linkToDir,
			})
			log.Printf("%s is linked into %s/.", bin, linkToDir)
		}
	default:
	}
}

func installRustNix(rustupHome string, cargoHome string, toolchain string) {
	command := utils.Format(`
		curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | {prefix} bash -s -- --default-toolchain {toolchain} -y \
		&& {cargoHome}/bin/rustup component add rust-src rustfmt clippy \
		&& {cargoHome}/bin/cargo install cargo-cache cargo-edit cargo-criterion`, map[string]string{
		"rustupHome": rustupHome,
		"cargoHome":  cargoHome,
		"toolchain":  toolchain,
		"prefix": utils.GetCommandPrefix(false, map[string]uint32{
			rustupHome: unix.W_OK | unix.R_OK,
			cargoHome:  unix.W_OK | unix.R_OK,
		}),
	})
	utils.RunCmd(command, "RUSTUP_HOME="+rustupHome, "CARGO_HOME="+cargoHome)
	command = utils.Format("{prefix} rm -rf {registry}", map[string]string{
		"registry": filepath.Join(cargoHome, "registry"),
		"prefix": utils.GetCommandPrefix(false, map[string]uint32{
			cargoHome: unix.W_OK | unix.R_OK,
		}),
	})
	utils.RunCmd(command)
	installCargoBinstall()
	installSccache()
}

func installSccache() {
	tmpdir := utils.CreateTempDir("")
	defer os.RemoveAll(tmpdir)
	file := filepath.Join(tmpdir, "sccache.tar.gz")
	network.DownloadGitHubRelease("mozilla/sccache", "", map[string][]string{
		"common": {"tar.gz"},
		"amd64":  {"x86_64"},
		"arm64":  {"aarch64"},
		"linux":  {"unknown", "linux", "musl"},
		"darwin": {"apple", "darwin"},
	}, []string{"pre", "dist", "sha256"}, file)
	command := utils.Format("{prefix} tar --wildcards --strip-components=1 -C /usr/local/bin/ -zxvf {file} */sccache", map[string]string{
		"prefix": utils.GetCommandPrefix(false, map[string]uint32{
			"/usr/local/bin": unix.W_OK | unix.R_OK,
		}),
		"file": file,
	})
	utils.RunCmd(command)
}

func installCargoBinstall() {
	tmpdir := utils.CreateTempDir("")
	defer os.RemoveAll(tmpdir)
	file := filepath.Join(tmpdir, "cargo-binstall.tgz")
	network.DownloadGitHubRelease("cargo-bins/cargo-binstall", "", map[string][]string{
		"common": {"tgz"},
		"amd64":  {"x86_64"},
		"arm64":  {"aarch64"},
		"linux":  {"unknown", "linux", "gnu"},
		"darwin": {"apple", "darwin"},
	}, []string{"pre", "full"}, file)
	command := utils.Format("{prefix} tar -C /usr/local/bin/ -zxvf {file}", map[string]string{
		"prefix": utils.GetCommandPrefix(false, map[string]uint32{
			"/usr/local/bin": unix.W_OK | unix.R_OK,
		}),
		"file": file,
	})
	utils.RunCmd(command)
}

// Install and configure Rust.
func rust(cmd *cobra.Command, _ []string) {
	rustupHome := utils.GetStringFlag(cmd, "rustup-home")
	if rustupHome == "" {
		rustupHome = filepath.Join(utils.UserHomeDir(), ".rustup")
	}
	cargoHome := utils.GetStringFlag(cmd, "cargo-home")
	if cargoHome == "" {
		cargoHome = filepath.Join(utils.UserHomeDir(), ".cargo")
	}
	toolchain := utils.GetStringFlag(cmd, "toolchain")
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "darwin":
			utils.BrewInstallSafe([]string{"pkg-config", "openssl"})
			installRustNix(rustupHome, cargoHome, toolchain)
		case "linux":
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format(`{prefix} apt-get update \
						&& {prefix} apt-get install -y gcc cmake libssl-dev pkg-config`, map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
				})
				utils.RunCmd(command)
			}
			installRustNix(rustupHome, cargoHome, toolchain)
		default:
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		linkRust(cmd, cargoHome)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		command := utils.Format(`RUSTUP_HOME={rustupHome} CARGO_HOME={cargoHome} PATH={cargoHome}/bin:$PATH \
				{prefix} rustup self uninstall`, map[string]string{
			"rustupHome": rustupHome,
			"cargoHome":  cargoHome,
			"prefix": utils.GetCommandPrefix(false, map[string]uint32{
				rustupHome: unix.W_OK | unix.R_OK,
				cargoHome:  unix.W_OK | unix.R_OK,
			}),
		})
		utils.RunCmd(command)
	}
}

var RustCmd = &cobra.Command{
	Use:     "rust",
	Aliases: []string{"rustup", "cargo"},
	Short:   "Install and configure Rust.",
	//Args:  cobra.ExactArgs(1),
	Run: rust,
}

func init() {
	RustCmd.Flags().BoolP("install", "i", false, "Install Rust.")
	RustCmd.Flags().BoolP("config", "c", false, "Configure Rust.")
	RustCmd.Flags().BoolP("uninstall", "u", false, "Uninstall Rust.")
	RustCmd.Flags().String("link-to-dir", "", "The directory to link commands (cargo and rustc) to.")
	RustCmd.Flags().String("rustup-home", "", "Value for the RUSTUP_HOME environment.")
	RustCmd.Flags().String("cargo-home", "", "Value for the CARGO_HOME environment.")
	RustCmd.Flags().String("toolchain", "stable", "The Rust toolchain (stable by default) to install.")
	RustCmd.Flags().BoolP("path", "p", false, "Configure the PATH environment variable.")
	// rootCmd.AddCommand(RustCmd)
}
