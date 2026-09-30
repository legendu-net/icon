package utils

import (
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Package is a package providing a command which is shared by multiple tools.
// Commands are the executables indicating that the package is already
// available, in which case it is not installed again. Apt, Dnf and Brew are the
// names of the package on the corresponding package manager. If Link is true
// and the first command is missing but another one of Commands exists (e.g.,
// fdfind, the name of fd in the Debian/Ubuntu series), the latter is symlinked
// into ~/.local/bin as the first command.
type Package struct {
	Commands []string
	Apt      string
	Dnf      string
	Brew     string
	Link     bool
}

// Packages shared by multiple tools.
var (
	PkgBat     = Package{Commands: []string{"bat", "batcat"}, Apt: "bat", Dnf: "bat", Brew: "bat", Link: true}
	PkgFd      = Package{Commands: []string{"fd", "fdfind"}, Apt: "fd-find", Dnf: "fd-find", Brew: "fd", Link: true}
	PkgRipgrep = Package{Commands: []string{"rg"}, Apt: "ripgrep", Dnf: "ripgrep", Brew: "ripgrep"}
	PkgUnzip   = Package{Commands: []string{"unzip"}, Apt: "unzip", Dnf: "unzip", Brew: "unzip"}
)

// missingPackages returns the names (picked by pkgName, which picks the name of
// the package on the package manager in use) of the packages which are not
// available on the current machine yet.
func missingPackages(pkgs []Package, pkgName func(Package) string) []string {
	names := []string{}
	for _, pkg := range pkgs {
		if slices.ContainsFunc(pkg.Commands, ExistsCommand) {
			continue
		}
		if name := pkgName(pkg); name != "" {
			names = append(names, name)
		}
	}
	return names
}

// installMissingPackages installs the missing packages using the package
// manager command template, which is filled in with the command prefix, the
// yes flag and the packages to install.
func installMissingPackages(pkgs []Package, yes bool, pkgName func(Package) string, template string) {
	names := missingPackages(pkgs, pkgName)
	if len(names) == 0 {
		return
	}
	command := Format(template, map[string]string{
		"prefix": GetCommandPrefix(true, map[string]uint32{}),
		"yesStr": IfElseString(yes, "-y", ""),
		"pkgs":   strings.Join(names, " "),
	})
	RunCmd(command)
}

// InstallPackages installs the packages which are missing on the current
// machine using the native package manager, all in one go, so that the
// packages already available (e.g., installed for another tool) are skipped
// without invoking the package manager at all. Afterwards, the commands of the
// packages with Link set are made available under their first names.
func InstallPackages(yes bool, pkgs ...Package) {
	brewName := func(pkg Package) string { return pkg.Brew }
	switch {
	case !IsLinux():
		BrewInstallSafe(missingPackages(pkgs, brewName))
	case IsAtomicLinux():
		// Packages cannot be installed into an image-based (rpm-ostree) Linux
		// distribution without layering them and rebooting,
		// so Homebrew is used instead.
		if ExistsCommand("brew") {
			BrewInstallSafe(missingPackages(pkgs, brewName))
		} else if names := missingPackages(pkgs, brewName); len(names) > 0 {
			log.Printf("WARNING: Homebrew is not available, so %s are not installed.", strings.Join(names, ", "))
		}
	case IsDebianUbuntuSeries():
		installMissingPackages(pkgs, yes, func(pkg Package) string { return pkg.Apt },
			`{prefix} apt-get {yesStr} update \
				&& {prefix} apt-get {yesStr} install {pkgs}`)
	case IsFedoraSeries():
		installMissingPackages(pkgs, yes, func(pkg Package) string { return pkg.Dnf },
			"{prefix} dnf {yesStr} install {pkgs}")
	default:
		log.Print("WARNING: the Linux distribution is not supported, so packages are not installed.")
	}
	for i := range pkgs {
		if pkgs[i].Link {
			linkPackageCommand(&pkgs[i])
		}
	}
}

// packageLink returns the path of the symbolic link which makes an alternative
// name of the command of the package available under its first name.
func packageLink(pkg *Package) string {
	return "~/.local/bin/" + pkg.Commands[0]
}

// linkPackageCommand symlinks the first available alternative command of the
// package into ~/.local/bin under the first name of Commands if the latter is
// not available.
func linkPackageCommand(pkg *Package) {
	command := pkg.Commands[0]
	if ExistsCommand(command) {
		return
	}
	var alt string
	for _, c := range pkg.Commands[1:] {
		if alt = LookPath(c); alt != "" {
			break
		}
	}
	if alt == "" {
		return
	}
	link := packageLink(pkg)
	// os.Lstat (instead of ExistsPath) so that a dangling symbolic link is detected.
	if _, err := os.Lstat(NormalizePath(link)); err == nil {
		log.Printf("WARNING: %s already exists but is not found as the command %s.", link, command)
		return
	}
	Symlink(alt, link)
	if !ExistsCommand(command) {
		log.Printf("WARNING: %s is not on PATH, so %s is not found.", filepath.Dir(link), command)
	}
}

// UninstallPackages uninstalls the packages using the native package manager
// and removes the symbolic links created by InstallPackages afterwards.
func UninstallPackages(yes bool, pkgs ...Package) {
	names := func(pkgName func(Package) string) string {
		ns := []string{}
		for _, pkg := range pkgs {
			ns = append(ns, pkgName(pkg))
		}
		return strings.Join(ns, " ")
	}
	var command string
	switch {
	case !IsLinux():
		command = "brew uninstall " + names(func(pkg Package) string { return pkg.Brew })
	case IsAtomicLinux():
		if !ExistsCommand("brew") {
			log.Print("WARNING: Homebrew is not available, so packages are not uninstalled.")
			return
		}
		command = "brew uninstall " + names(func(pkg Package) string { return pkg.Brew })
	case IsDebianUbuntuSeries():
		command = "{prefix} apt-get {yesStr} purge " + names(func(pkg Package) string { return pkg.Apt })
	case IsFedoraSeries():
		command = "{prefix} dnf {yesStr} remove " + names(func(pkg Package) string { return pkg.Dnf })
	default:
		log.Print("WARNING: the Linux distribution is not supported, so packages are not uninstalled.")
		return
	}
	RunCmd(Format(command, map[string]string{
		"prefix": GetCommandPrefix(true, map[string]uint32{}),
		"yesStr": IfElseString(yes, "-y", ""),
	}))
	for i := range pkgs {
		if !pkgs[i].Link {
			continue
		}
		link := packageLink(&pkgs[i])
		target, err := os.Readlink(NormalizePath(link))
		if err == nil && slices.Contains(pkgs[i].Commands[1:], filepath.Base(target)) {
			RemoveAll(link)
		}
	}
}
