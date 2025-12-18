package dev

import (
	"context"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"legendu.net/icon/utils"
)

func getGolangVersion() string {
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodGet, "https://github.com/golang/go/tags", http.NoBody)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	html := utils.ReadAllAsText(resp.Body)
	if utils.IsErrorHTTPResponse(resp) {
		log.Fatal("...")
	}
	re := regexp.MustCompile(`tag/go(\d+\.\d+\.\d+)`)
	return re.FindStringSubmatch(html)[1]
}

func installGoLang(prefix string) {
	url := utils.Format("https://go.dev/dl/go{ver}.{os}-{arch}.tar.gz", map[string]string{
		"ver":  getGolangVersion(),
		"os":   runtime.GOOS,
		"arch": utils.HostKernelArch(),
	})
	goTgz, err := utils.DownloadFile(url, "go_*.tar.gz", true)
	if err != nil {
		log.Fatal(err)
	}
	cmd := utils.Format(`{prefix} rm -rf /usr/local/go \
				&& {prefix} tar -C /usr/local/ -xzf {goTgz}`,
		map[string]string{
			"prefix": prefix,
			"goTgz":  goTgz,
		},
	)
	utils.RunCmd(cmd)
}

func installGoLangCiLint(prefix string) {
	script := "https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh"
	cmd := utils.Format(`curl -sSfL {script} | {prefix} sh -s -- -b /usr/local/go/bin`, map[string]string{
		"script": script,
		"prefix": prefix,
	})
	utils.RunCmd(cmd)
}

func installGoPls(prefix string) {
	cmd := utils.Format("{prefix} go install golang.org/x/tools/gopls@latest", map[string]string{
		"prefix": prefix,
	})
	utils.RunCmd(cmd)
}

// Install and configure Golang.
func golang(cmd *cobra.Command, _ []string) {
	prefix := utils.GetCommandPrefix(false, map[string]uint32{
		"/usr/local/go":  unix.W_OK | unix.R_OK,
		"/usr/local":     unix.W_OK | unix.R_OK,
		"/usr/local/bin": unix.W_OK | unix.R_OK,
	})
	if utils.GetBoolFlag(cmd, "install") {
		installGoLang(prefix)
		installGoLangCiLint(prefix)
		installGoPls(prefix)
	}
	if utils.GetBoolFlag(cmd, "config") {
		if utils.IsLinux() {
			usrLocalBin := "/usr/local/bin"
			goBin := "/usr/local/go/bin"
			entries := utils.ReadDir(goBin)
			for _, entry := range entries {
				file := filepath.Join(goBin, entry.Name())
				utils.SymlinkIntoDir(file, usrLocalBin, false, false)
			}
		} else {
		}
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
	}
}

var GolangCmd = &cobra.Command{
	Use:     "golang",
	Aliases: []string{"go"},
	Short:   "Install and configure Golang.",
	//Args:  cobra.ExactArgs(1),
	Run: golang,
}

func init() {
	GolangCmd.Flags().BoolP("install", "i", false, "Install Golang.")
	GolangCmd.Flags().BoolP("uninstall", "u", false, "Uninstall Golang.")
	GolangCmd.Flags().BoolP("config", "c", false, "Configure Golang.")
	GolangCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	GolangCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
}
