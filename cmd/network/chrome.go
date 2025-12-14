package network

import (
	"log"
	"path/filepath"
	"runtime"

	"legendu.net/icon/utils"
)

func getExtensionDir() string {
	switch runtime.GOOS {
	case "darwin":
		return "~/Library/Application Support/Google/Chrome/Default/Extensions"
	default:
		return "~/.config/google-chrome/Default/Extensions"
	}
}

func InstallChromeExtension(id string, name string) {
	dir := getExtensionDir()
	utils.MkdirAll(dir, "700")
	config := filepath.Join(dir, id+".json")
	utils.WriteTextFile(config, `{"external_update_url": "https://clients2.google.com/service/update2/crx"}`, 0o600)
	log.Printf("Installed %s (%s)", config, name)
}
