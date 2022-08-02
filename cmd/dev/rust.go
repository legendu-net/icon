package dev

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
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

// Install and configure Rust.
func rust(cmd *cobra.Command, args []string) {
	rustupHome := utils.GetStringFlag(cmd, "rustup-home")
	if rustupHome == "" {
		rustupHome = filepath.Join(utils.UserHomeDir(), ".rustup")
	}
	cargoHome := utils.GetStringFlag(cmd, "cargo-home")
	if cargoHome == "" {
		cargoHome = filepath.Join(utils.UserHomeDir(), ".cargo")
	}
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "linux", "darwin":
			if utils.IsDebianSeries() {
				command := utils.Format(`{prefix} apt-get update \
						&& {prefix} apt-get install -y gcc cmake libssl-dev pkg-config`, map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
				})
				utils.RunCmd(command)
			}
			command := utils.Format(`
				curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | {prefix} bash -s -- -y \
                && {cargoHome}/bin/rustup component add rust-src rustfmt clippy \
                && {cargoHome}/bin/cargo install sccache cargo-cache cargo-edit`, map[string]string{
				"rustupHome": rustupHome,
				"cargoHome":  cargoHome,
				"prefix": utils.GetCommandPrefix(false, map[string]uint32{
					rustupHome: unix.W_OK | unix.R_OK,
					cargoHome:  unix.W_OK | unix.R_OK,
				}),
			})
			utils.RunCmd(command, "RUSTUP_HOME="+rustupHome, "CARGO_HOME="+cargoHome)
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
	RustCmd.Flags().BoolP("path", "p", false, "Configure the PATH environment variable.")
	// rootCmd.AddCommand(RustCmd)
}
