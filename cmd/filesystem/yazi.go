package filesystem

import (
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/dev"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/utils"
)

// yaziDep is an auxiliary tool leveraged by Yazi for previewing, searching and
// handling archives. Commands are the executables indicating that the
// dependency is already available, in which case it is not installed again.
// Apt, Dnf and Brew are the names of the package providing the dependency on
// the corresponding package manager.
type yaziDep struct {
	Commands []string
	Apt      string
	Dnf      string
	Brew     string
}

// yaziDeps are the optional dependencies recommended by Yazi. unzip and git are
// not dependencies of Yazi itself but are required to extract its release
// archive and to install its plugins respectively. fd is installed separately
// by InstallFd, which also makes fdfind available as fd.
var yaziDeps = []yaziDep{
	{Commands: []string{"git"}, Apt: "git", Dnf: "git", Brew: "git"},
	{Commands: []string{"ffmpeg"}, Apt: "ffmpeg", Dnf: "ffmpeg-free", Brew: "ffmpeg"},
	{Commands: []string{"7z", "7zz"}, Apt: "7zip", Dnf: "7zip", Brew: "sevenzip"},
	{Commands: []string{"jq"}, Apt: "jq", Dnf: "jq", Brew: "jq"},
	{Commands: []string{"pdftoppm"}, Apt: "poppler-utils", Dnf: "poppler-utils", Brew: "poppler"},
	{Commands: []string{"rg"}, Apt: "ripgrep", Dnf: "ripgrep", Brew: "ripgrep"},
	{Commands: []string{"fzf"}, Apt: "fzf", Dnf: "fzf", Brew: "fzf"},
	{Commands: []string{"zoxide"}, Apt: "zoxide", Dnf: "zoxide", Brew: "zoxide"},
	{Commands: []string{"magick", "convert"}, Apt: "imagemagick", Dnf: "ImageMagick", Brew: "imagemagick"},
	{Commands: []string{"chafa"}, Apt: "chafa", Dnf: "chafa", Brew: "chafa"},
	{Commands: []string{"file"}, Apt: "file", Dnf: "file", Brew: "file-formula"},
	{Commands: []string{"unzip"}, Apt: "unzip", Dnf: "unzip", Brew: "unzip"},
}

// missingYaziDeps returns the packages (named by pkgName, which picks the
// package name of the package manager in use) of the Yazi dependencies which
// are not available on the current machine yet.
func missingYaziDeps(pkgName func(yaziDep) string) []string {
	pkgs := []string{}
	for _, dep := range yaziDeps {
		if slices.ContainsFunc(dep.Commands, utils.ExistsCommand) {
			continue
		}
		if pkg := pkgName(dep); pkg != "" {
			pkgs = append(pkgs, pkg)
		}
	}
	return pkgs
}

// installMissingYaziDeps installs the missing Yazi dependencies using the
// package manager command template, which is filled in with the command prefix,
// the yes flag and the packages to install.
func installMissingYaziDeps(cmd *cobra.Command, pkgName func(yaziDep) string, template string) {
	pkgs := missingYaziDeps(pkgName)
	if len(pkgs) == 0 {
		return
	}
	command := utils.Format(template, map[string]string{
		"prefix": utils.GetCommandPrefix(
			true,
			map[string]uint32{},
		),
		"yesStr": utils.BuildYesFlag(cmd),
		"pkgs":   strings.Join(pkgs, " "),
	})
	utils.RunCmd(command)
}

// installYaziDeps installs the dependencies of Yazi which are missing on the
// current machine using the native package manager.
func installYaziDeps(cmd *cobra.Command) {
	brewName := func(dep yaziDep) string { return dep.Brew }
	if !utils.IsLinux() {
		utils.BrewInstallSafe(missingYaziDeps(brewName))
		return
	}
	switch {
	case utils.IsAtomicLinux():
		// Packages cannot be installed into an image-based (rpm-ostree) Linux
		// distribution without layering them and rebooting,
		// so Homebrew is used instead.
		if utils.ExistsCommand("brew") {
			utils.BrewInstallSafe(missingYaziDeps(brewName))
		} else {
			log.Print("WARNING: Homebrew is not available, so dependencies of Yazi are not installed.")
		}
	case utils.IsDebianUbuntuSeries():
		installMissingYaziDeps(cmd, func(dep yaziDep) string { return dep.Apt },
			`{prefix} apt-get {yesStr} update \
					&& {prefix} apt-get {yesStr} install {pkgs}`)
	case utils.IsFedoraSeries():
		installMissingYaziDeps(cmd, func(dep yaziDep) string { return dep.Dnf },
			"{prefix} dnf {yesStr} install {pkgs}")
	default:
		log.Print("WARNING: the Linux distribution is not supported, so dependencies of Yazi are not installed.")
	}
}

// installYazi downloads the prebuilt Yazi binaries (yazi and ya) from its
// GitHub releases and installs them. When global is true the binaries are
// installed into /usr/local/bin; otherwise they are installed into
// ~/.local/bin (no privilege escalation). Yazi is not reliably packaged in the
// Debian/Ubuntu and Fedora series, so the official prebuilt binaries are used.
func installYazi(global bool) {
	if !utils.ExistsCommand("unzip") {
		log.Fatal("The command unzip is required to extract the release archive of Yazi but is not available.")
	}
	tmpdir := utils.CreateTempDir("")
	defer os.RemoveAll(tmpdir)
	file := filepath.Join(tmpdir, "yazi.zip")
	network.DownloadGitHubRelease("sxyazi/yazi", "", map[string][]string{
		"common": {".zip"},
		"amd64":  {"x86_64"},
		"arm64":  {"aarch64"},
		"linux":  {"linux", "gnu"},
		"darwin": {"apple", "darwin"},
	}, []string{}, file)
	prefix := ""
	binDir := "~/.local/bin"
	if global {
		prefix = utils.GetCommandPrefix(
			true,
			map[string]uint32{},
		)
		binDir = "/usr/local/bin"
	}
	command := utils.Format(`{prefix} mkdir -p {binDir} \
		&& {prefix} unzip -o -j {file} '*/yazi' '*/ya' -d {binDir} \
		&& {prefix} chmod +x {binDir}/yazi {binDir}/ya`, map[string]string{
		"prefix": prefix,
		"binDir": binDir,
		"file":   file,
	})
	utils.RunCmd(command)
}

// yaziPlugin is a plugin of Yazi. Pkg is the name of the package as accepted by
// `ya pkg add`, which is either owner/repo or owner/repo:sub-package for a
// repository hosting multiple plugins. Requires are the commands the plugin
// needs at runtime. Install installs those commands (all of them, not one by
// one) and is nil for a plugin whose requirements icon cannot install, in which
// case a missing requirement is only warned about.
type yaziPlugin struct {
	Pkg      string
	Requires []string
	Install  func(global bool)
}

// yaziPlugins are the Yazi plugins to install.
var yaziPlugins = []yaziPlugin{
	{Pkg: "Adda0/jjui", Requires: []string{"jj", "jjui"}, Install: dev.InstallJjTools},
}

// existsPluginRequire reports whether a command required by a Yazi plugin is
// available. The candidate installation directories are checked in addition to
// PATH because ~/.local/bin is not necessarily on the current shell's PATH, in
// which case a command installed there would be reinstalled on every run.
func existsPluginRequire(command string) bool {
	if utils.ExistsCommand(command) {
		return true
	}
	for _, binDir := range []string{"~/.local/bin", "/usr/local/bin"} {
		if utils.ExistsFile(binDir + "/" + command) {
			return true
		}
	}
	return false
}

// installYaziPluginRequires installs the commands which the plugin needs at
// runtime but which are missing on the current machine.
func installYaziPluginRequires(plugin yaziPlugin, global bool) {
	missing := []string{}
	for _, require := range plugin.Requires {
		if !existsPluginRequire(require) {
			missing = append(missing, require)
		}
	}
	if len(missing) == 0 {
		return
	}
	noun := utils.IfElseString(len(missing) == 1, "command", "commands")
	if plugin.Install == nil {
		log.Printf("WARNING: the Yazi plugin %s is missing the required %s %s.",
			plugin.Pkg, noun, strings.Join(missing, ", "))
		return
	}
	log.Printf("Installing the %s %s required by the Yazi plugin %s ...",
		noun, strings.Join(missing, ", "), plugin.Pkg)
	plugin.Install(global)
}

// resolveYa returns the command used to invoke ya (the package manager of
// Yazi). It prefers the ya found on PATH; if ya is not on PATH (e.g. just
// installed into ~/.local/bin which is not yet in the current shell's PATH),
// it falls back to the absolute path in the candidate installation directories.
func resolveYa() string {
	if yaPath := utils.LookPath("ya"); yaPath != "" {
		return yaPath
	}
	for _, binDir := range []string{"~/.local/bin", "/usr/local/bin"} {
		yaPath := binDir + "/ya"
		if utils.ExistsFile(yaPath) {
			return utils.NormalizePath(yaPath)
		}
	}
	return "ya"
}

// yaziConfigHome returns the configuration directory of Yazi, which is
// overridden by the environment variables YAZI_CONFIG_HOME and
// XDG_CONFIG_HOME (in this order of precedence).
func yaziConfigHome() string {
	if dir := os.Getenv("YAZI_CONFIG_HOME"); dir != "" {
		return dir
	}
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "yazi")
	}
	return "~/.config/yazi"
}

// yaziPluginDir returns the directory into which `ya pkg add` installs the
// package pkg, which is named after the sub-package if there is one and after
// the repository otherwise.
func yaziPluginDir(pkg string) string {
	name := filepath.Base(pkg)
	if _, sub, found := strings.Cut(pkg, ":"); found {
		name = sub
	}
	return filepath.Join(yaziConfigHome(), "plugins", name+".yazi")
}

// installYaziPlugins installs the Yazi plugins which are not installed yet,
// together with the commands they need at runtime.
// `ya pkg add` fails on an already installed plugin, so the existing ones are
// skipped, which also keeps manual customizations in package.toml intact.
func installYaziPlugins(global bool) {
	ya := resolveYa()
	// Deploy the packages which are declared in package.toml but missing from
	// the plugins directory. This is a no-op (and not an error) otherwise.
	command := utils.Format("{ya} pkg install", map[string]string{"ya": ya})
	utils.RunCmd(command)
	for _, plugin := range yaziPlugins {
		dir := yaziPluginDir(plugin.Pkg)
		if utils.ExistsDir(dir) {
			log.Printf("The Yazi plugin %s is already installed into %s.", plugin.Pkg, dir)
		} else {
			command := utils.Format("{ya} pkg add {pkg}", map[string]string{
				"ya":  ya,
				"pkg": plugin.Pkg,
			})
			utils.RunCmd(command)
		}
		installYaziPluginRequires(plugin, global)
	}
}

// configYazi symlinks (or copies) the configuration of Yazi tracked in
// icon-data into the configuration directory of Yazi. The entries of
// ~/.config/icon-data/yazi are handled one by one (instead of the directory as
// a whole) so that what `ya pkg` manages, namely the plugins directory and the
// package.toml lock file, stays local to the machine.
func configYazi(cmd *cobra.Command) {
	icon.FetchConfigData(false, "")
	src := "~/.config/icon-data/yazi"
	if !utils.ExistsDir(src) {
		log.Fatalf("The Yazi configuration directory %s does not exist.", src)
	}
	configHome := yaziConfigHome()
	for _, entry := range utils.ReadDir(src) {
		dst := filepath.Join(configHome, entry.Name())
		utils.BackupOrRemove(dst, utils.ShouldBackup(cmd))
		utils.CopyOrSymlink(filepath.Join(src, entry.Name()), dst, utils.GetBoolFlag(cmd, "copy"))
	}
}

// uninstallYazi removes the Yazi binaries installed by installYazi. It searches
// the candidate installation directories (~/.local/bin and /usr/local/bin) and
// removes the binaries wherever they are found. RemoveAll skips non-existent
// paths and derives privilege escalation from the target path's write
// permissions, so the system location is handled with sudo only when necessary.
// The dependencies of Yazi are general-purpose tools which are likely shared
// with other applications, so they are kept.
func uninstallYazi() {
	for _, binDir := range []string{"~/.local/bin", "/usr/local/bin"} {
		for _, bin := range []string{"yazi", "ya"} {
			utils.RemoveAll(binDir + "/" + bin)
		}
	}
}

// Install and configure Yazi.
func yazi(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		installYaziDeps(cmd)
		InstallFd(utils.GetBoolFlag(cmd, "yes"))
		if utils.IsLinux() {
			installYazi(utils.GetBoolFlag(cmd, "global"))
		} else {
			utils.BrewInstallSafe([]string{"yazi"})
		}
		installYaziPlugins(utils.GetBoolFlag(cmd, "global"))
	}
	if utils.GetBoolFlag(cmd, "config") {
		configYazi(cmd)
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		if utils.IsLinux() {
			uninstallYazi()
		} else {
			utils.RunCmd("brew uninstall yazi")
		}
	}
}

var yaziCmd = &cobra.Command{
	Use:     "yazi",
	Aliases: []string{},
	Short:   "Install and configure Yazi.",
	//Args:  cobra.ExactArgs(1),
	Run: yazi,
}

func ConfigYaziCmd(rootCmd *cobra.Command) {
	yaziCmd.Flags().BoolP("install", "i", false, "Install Yazi.")
	yaziCmd.Flags().BoolP("uninstall", "u", false, "Uninstall Yazi.")
	yaziCmd.Flags().BoolP("config", "c", false, "Configure Yazi.")
	yaziCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	yaziCmd.Flags().Bool("global", false, "Install Yazi and the commands required by its plugins into /usr/local/bin instead of ~/.local/bin.")
	yaziCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	yaziCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	rootCmd.AddCommand(yaziCmd)
}
