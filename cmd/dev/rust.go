package dev

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
)

const Linux = "linux"
const Darwin = "darwin"

func linkRust(cmd *cobra.Command, cargoHome string) {
	linkToDir := utils.GetStringFlag(cmd, "link-to-dir")
	if linkToDir == "" {
		return
	}
	cargoBin := filepath.Join(cargoHome, "bin")
	for _, entry := range utils.ReadDir(cargoBin) {
		utils.SymlinkIntoDir(filepath.Join(cargoBin, entry.Name()), linkToDir, false, false)
	}
}

func installRustNix(rustupHome, cargoHome, toolchain string) {
	command := utils.Format(`
		curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | \
			{prefix} bash -s -- --default-toolchain {toolchain} -y \
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
		Linux:    {"unknown", Linux, "musl"},
		Darwin:   {"apple", Darwin},
	}, []string{"pre", "dist", "sha256"}, file)
	command := utils.Format(`{prefix} tar --wildcards --strip-components=1 \
			-C /usr/local/bin/ -zxvf {file} */sccache`, map[string]string{
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
		Linux:    {"unknown", Linux, "gnu"},
		Darwin:   {"apple", Darwin},
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
		if utils.IsLinux() {
			if utils.IsDebianUbuntuSeries() {
				command := utils.Format(`{prefix} apt-get update \
						&& {prefix} apt-get install -y gcc cmake libssl-dev pkg-config`, map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
				})
				utils.RunCmd(command)
			}
			installRustNix(rustupHome, cargoHome, toolchain)
		} else {
			utils.BrewInstallSafe([]string{"pkg-config", "openssl"})
			installRustNix(rustupHome, cargoHome, toolchain)
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

var rustCmd = &cobra.Command{
	Use:     "rust",
	Aliases: []string{"rustup", "cargo"},
	Short:   "Install and configure Rust.",
	//Args:  cobra.ExactArgs(1),
	Run: rust,
}

func ConfigRustCmd(rootCmd *cobra.Command) {
	rustCmd.Flags().BoolP("install", "i", false, "Install Rust.")
	rustCmd.Flags().BoolP("config", "c", false, "Configure Rust.")
	rustCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	rustCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	rustCmd.Flags().BoolP("uninstall", "u", false, "Uninstall Rust.")
	rustCmd.Flags().String("link-to-dir", "", "The directory to link commands (cargo and rustc) to.")
	rustCmd.Flags().String("rustup-home", "", "Value for the RUSTUP_HOME environment.")
	rustCmd.Flags().String("cargo-home", "", "Value for the CARGO_HOME environment.")
	rustCmd.Flags().String("toolchain", "stable", "The Rust toolchain (stable by default) to install.")
	rustCmd.Flags().BoolP("path", "p", false, "Configure the PATH environment variable.")
	rootCmd.AddCommand(rustCmd)
}
