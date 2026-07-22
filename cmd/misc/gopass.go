package misc

import (
	"log"
	"strconv"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/utils"
)

// gopassConfig holds the gopass-specific settings used to set up the gopass
// store. The user identity (name and email) is single-sourced from
// ~/.config/icon-data/user.yaml via utils.ReadUserConfig.
// AgeAgentTimeout is the number of seconds the age agent caches an unlocked
// identity. It is optional: any non-positive value (including the 0 of an
// absent field) falls back to defaultAgeAgentTimeout.
type gopassConfig struct {
	GitURL          string `yaml:"gitUrl"`
	AgeAgentTimeout int    `yaml:"ageAgentTimeout"`
}

// defaultAgeAgentTimeout is the age.agent-timeout (in seconds) applied when
// ageAgentTimeout is not configured in ~/.config/icon-data/gopass/config.yaml.
const defaultAgeAgentTimeout = 7200

// runPackageCmd dispatches a package-manager command (install/uninstall) for
// the current OS, filling in the sudo prefix and yes-flag for the given
// apt-get/dnf templates, or running the brew command directly on macOS.
func runPackageCmd(cmd *cobra.Command, debian, fedora, brew string) {
	if utils.IsLinux() {
		if utils.IsDebianUbuntuSeries() {
			utils.RunCmd(utils.Format(debian, map[string]string{
				"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
				"yesStr": utils.BuildYesFlag(cmd),
			}))
		} else if utils.IsFedoraSeries() {
			utils.RunCmd(utils.Format(fedora, map[string]string{
				"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
				"yesStr": utils.BuildYesFlag(cmd),
			}))
		}
	} else {
		utils.RunCmd(brew)
	}
}

// Install and configure gopass.
func gopass(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		runPackageCmd(cmd,
			`{prefix} apt-get {yesStr} update \
					&& {prefix} apt-get {yesStr} install gopass age`,
			"{prefix} dnf {yesStr} install gopass age",
			"brew install gopass age",
		)
	}
	if utils.GetBoolFlag(cmd, "config") {
		icon.FetchConfigData(false, "")
		configFile := "~/.config/icon-data/gopass/config.yaml"
		if !utils.ExistsFile(configFile) {
			log.Fatalf("The gopass configuration file %s does not exist.", configFile)
		}
		var cfg gopassConfig
		if err := yaml.Unmarshal(utils.ReadFile(configFile), &cfg); err != nil {
			log.Fatalf("Error parsing %s: %v", configFile, err)
		}
		if cfg.GitURL == "" {
			log.Fatalf("gitUrl is not configured in %s.", configFile)
		}
		user := utils.ReadUserConfig()
		store := "~/.local/share/gopass/stores/root"
		utils.BackupOrRemove(store, utils.ShouldBackup(cmd))
		utils.RunCmd(utils.Format(
			`gopass setup --crypto age --storage gitfs \
				--remote "{gitUrl}" \
				--name "{userName}" \
				--email "{userEmail}"`,
			map[string]string{
				"gitUrl":    cfg.GitURL,
				"userName":  user.UserName,
				"userEmail": user.UserEmail,
			},
		))
		ageAgentTimeout := cfg.AgeAgentTimeout
		if ageAgentTimeout <= 0 {
			ageAgentTimeout = defaultAgeAgentTimeout
		}
		utils.RunCmd("gopass config age.agent-enabled true")
		utils.RunCmd(utils.Format(
			"gopass config age.agent-timeout {ageAgentTimeout}",
			map[string]string{"ageAgentTimeout": strconv.Itoa(ageAgentTimeout)},
		))
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		runPackageCmd(cmd,
			"{prefix} apt-get {yesStr} purge gopass age",
			"{prefix} dnf {yesStr} remove gopass age",
			"brew uninstall gopass age",
		)
	}
}

var gopassCmd = &cobra.Command{
	Use:     "gopass",
	Aliases: []string{},
	Short:   "Install and configure gopass.",
	Run:     gopass,
}

func ConfigGopassCmd(rootCmd *cobra.Command) {
	gopassCmd.Flags().BoolP("install", "i", false, "Install gopass.")
	gopassCmd.Flags().Bool("uninstall", false, "Uninstall gopass.")
	gopassCmd.Flags().BoolP("config", "c", false, "Configure gopass.")
	gopassCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	gopassCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	gopassCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	rootCmd.AddCommand(gopassCmd)
}
