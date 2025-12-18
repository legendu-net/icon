package network

import (
	"log"
	"path/filepath"
	"runtime"

	"legendu.net/icon/utils"
)

func getExtensionDir() string {
	if runtime.GOOS == "darwin" {
		return "~/Library/Application Support/Google/Chrome/Default/Extensions"
	}
	return "~/.config/google-chrome/Default/Extensions"
}

func InstallChromeExtension(id, name string) {
	dir := getExtensionDir()
	utils.MkdirAll(dir, "700")
	config := filepath.Join(dir, id+".json")
	//nolint:mnd // readable
	utils.WriteTextFile(config, `{"external_update_url": "https://clients2.google.com/service/update2/crx"}`, 0o600)
	log.Printf("Installed %s (%s)", config, name)
}
