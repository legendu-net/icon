package filesystem

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// fdLink is the symbolic link which makes fdfind (the name of fd in the
// Debian/Ubuntu series) available as fd.
const fdLink = "~/.local/bin/fd"

// runFdPackageManager manages the fd package using the package manager of the
// current platform. The apt and dnf command templates are filled in with the
// command prefix and the yes flag. Homebrew is used on macOS and on
// image-based (rpm-ostree) Linux distributions, into which packages cannot be
// installed without layering them and rebooting.
func runFdPackageManager(yes bool, brew func(), aptTemplate, dnfTemplate string) {
	var template string
	switch {
	case !utils.IsLinux():
		brew()
		return
	case utils.IsAtomicLinux():
		if utils.ExistsCommand("brew") {
			brew()
		} else {
			log.Print("WARNING: Homebrew is not available, so fd is not managed.")
		}
		return
	case utils.IsDebianUbuntuSeries():
		template = aptTemplate
	case utils.IsFedoraSeries():
		template = dnfTemplate
	default:
		log.Print("WARNING: the Linux distribution is not supported, so fd is not managed.")
		return
	}
	command := utils.Format(template, map[string]string{
		"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
		"yesStr": utils.IfElseString(yes, "-y", ""),
	})
	utils.RunCmd(command)
}

// linkFd makes fdfind available as fd (which tools such as Television invoke)
// by symlinking it into ~/.local/bin if fd is not available but fdfind is.
func linkFd() {
	if utils.ExistsCommand("fd") {
		return
	}
	fdfind := utils.LookPath("fdfind")
	if fdfind == "" {
		return
	}
	// os.Lstat (instead of utils.ExistsPath) so that a dangling symbolic link is detected.
	if _, err := os.Lstat(utils.NormalizePath(fdLink)); err == nil {
		log.Printf("WARNING: %s already exists but is not found as the command fd.", fdLink)
		return
	}
	utils.Symlink(fdfind, fdLink)
	if !utils.ExistsCommand("fd") {
		log.Printf("WARNING: %s is not on PATH, so fd is not found.", filepath.Dir(fdLink))
	}
}

// InstallFd installs fd if neither fd nor fdfind is available, and makes
// fdfind available as fd if necessary. It is exported because other tools
// (Television) rely on fd, so that they install it exactly the way
// `icon fd --install` does.
func InstallFd(yes bool) {
	if !utils.ExistsCommand("fd") && !utils.ExistsCommand("fdfind") {
		runFdPackageManager(
			yes,
			func() { utils.BrewInstallSafe([]string{"fd"}) },
			`{prefix} apt-get {yesStr} update \
				&& {prefix} apt-get {yesStr} install fd-find`,
			"{prefix} dnf {yesStr} install fd-find",
		)
	}
	linkFd()
}

// uninstallFd removes the symbolic link created by linkFd and the fd package.
func uninstallFd(yes bool) {
	if target, err := os.Readlink(utils.NormalizePath(fdLink)); err == nil && filepath.Base(target) == "fdfind" {
		utils.RemoveAll(fdLink)
	}
	runFdPackageManager(
		yes,
		func() { utils.RunCmd("brew uninstall fd") },
		"{prefix} apt-get {yesStr} purge fd-find",
		"{prefix} dnf {yesStr} remove fd-find",
	)
}

// Install and configure fd.
func fd(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		InstallFd(utils.GetBoolFlag(cmd, "yes"))
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		uninstallFd(utils.GetBoolFlag(cmd, "yes"))
	}
}

var fdCmd = &cobra.Command{
	Use:     "fd",
	Aliases: []string{"fdfind"},
	Short:   "Install fd.",
	//Args:  cobra.ExactArgs(1),
	Run: fd,
}

func ConfigFdCmd(rootCmd *cobra.Command) {
	fdCmd.Flags().BoolP("install", "i", false, "Install fd.")
	fdCmd.Flags().BoolP("uninstall", "u", false, "Uninstall fd.")
	fdCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	rootCmd.AddCommand(fdCmd)
}
