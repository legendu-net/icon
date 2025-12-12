package misc

import (
	//"log"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/utils"
)

func readDefaultKeybindings(file string) []string {
	defaultKeyBinding := []string{
		"{",
	}
	if utils.ExistsFile(file) {
		defaultKeyBinding = strings.Split(
			strings.TrimSpace(utils.ReadFileAsString(file)), "\n")
		defaultKeyBinding = defaultKeyBinding[:len(defaultKeyBinding)-1]
	}
	return defaultKeyBinding
}

func readDefaultKeybindingsFromYaml() map[string]string {
	var keyBindings map[string]string
	err := yaml.Unmarshal(
		utils.ReadFile("~/.config/icon-data/keyboard/DefaultKeyBinding.yaml"), &keyBindings)
	if err != nil {
		log.Fatalf("Error unmarshaling data: %v", err)
	}
	return keyBindings
}

func hasAnyPrefix(kb string, keyBindings map[string]string) bool {
	kb = strings.TrimSpace(kb)
	for key := range keyBindings {
		key = "\"" + key + "\""
		if strings.HasPrefix(kb, key) {
			return true
		}
	}
	return false
}

func removeDefaultKeyBindings(defaultKeyBindings []string, keyBindings map[string]string) []string {
	j := 0
	for i, kb := range defaultKeyBindings {
		if !hasAnyPrefix(kb, keyBindings) {
			defaultKeyBindings[j] = defaultKeyBindings[i]
			j++
		}
	}
	return defaultKeyBindings[:j]
}

func addDefaultKeyBindings(defaultKeyBindings []string, keyBindings map[string]string) []string {
	lines := make([]string, len(keyBindings)+1)
	i := 0
	for key, binding := range keyBindings {
		lines[i] = "    \"" + key + "\" = \"" + binding + "\";"
		i += 1
	}
	lines[i] = "}"
	return append(defaultKeyBindings, lines...)
}

func ConfigDefaultKeybindings() {
	dir := "~/Library/KeyBindings"
	utils.MkdirAll(dir, "700")
	file := filepath.Join(dir, "DefaultKeyBinding.dict")
	defaultKeyBindings := readDefaultKeybindings(file)
	keyBindings := readDefaultKeybindingsFromYaml()
	defaultKeyBindings = removeDefaultKeyBindings(defaultKeyBindings, keyBindings)
	defaultKeyBindings = addDefaultKeyBindings(defaultKeyBindings, keyBindings)
	utils.WriteTextFile(file, strings.Join(defaultKeyBindings, "\n"), 0o600)
	fmt.Printf("%s has been updated using keyboard/DefaultKeyBinding.yaml.\n", file)
}

// Configure keyboard.
func keyboard(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		// nothing to install
	}
	if utils.GetBoolFlag(cmd, "config") {
		icon.FetchConfigData(false, "")
		switch runtime.GOOS {
		case "darwin":
			ConfigDefaultKeybindings()
		default:
		}
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		// nothing to uninstall
	}
}

var KeyboardCmd = &cobra.Command{
	Use:     "keyboard",
	Aliases: []string{"kb"},
	Short:   "Configure keyboard related.",
	//Args:  cobra.ExactArgs(1),
	Run: keyboard,
}

func init() {
	KeyboardCmd.Flags().BoolP("install", "i", false, "Install keyboard related tools.")
	KeyboardCmd.Flags().Bool("uninstall", false, "Uninstall keyboard related terminal.")
	KeyboardCmd.Flags().BoolP("config", "c", false, "Configure keyboard related.")
}
