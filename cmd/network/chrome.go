package network

import (
	"log"
	"path/filepath"
	"runtime"

	"legendu.net/icon/utils"
)

func getExtensionDir() string {
	home := utils.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library/Application Support/Google/Chrome/Default/Extensions")
	case "windows":
		return filepath.Join(home, "AppData/Local/Google/Chrome/User Data/Default/Extensions")
	default:
		return filepath.Join(home, ".config/google-chrome/Default/Extensions")
	}
}

func InstallChromeExtension(id string, name string) {
	dir := getExtensionDir()
	utils.MkdirAll(dir, "700")
	config := filepath.Join(dir, id+".json")
	utils.WriteTextFile(config, `{"external_update_url": "https://clients2.google.com/service/update2/crx"}`, 0o600)
	log.Printf("Installed %s (%s)", config, name)
}
