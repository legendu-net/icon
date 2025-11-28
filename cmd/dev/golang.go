package dev

import (
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
	resp, err := http.Get("https://github.com/golang/go/tags")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	html := utils.ReadAllAsText(resp.Body)
	if resp.StatusCode > 399 {
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
	goTgz := utils.DownloadFile(url, "go_*.tar.gz", true)
	cmd := utils.Format(`{prefix} rm -rf /usr/local/go \
				&& {prefix} tar -C /usr/local/ -xzf {goTgz}\
				&& {prefix} rm -rf /usr/local/go/pkg/*/cmd`,
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
func golang(cmd *cobra.Command, args []string) {
	prefix := utils.GetCommandPrefix(false, map[string]uint32{
		"/usr/local/go":  unix.W_OK | unix.R_OK,
		"/usr/local":     unix.W_OK | unix.R_OK,
		"/usr/local/bin": unix.W_OK | unix.R_OK,
	})
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "darwin", "linux":
			installGoLang(prefix)
			installGoLangCiLint(prefix)
			installGoPls(prefix)
		default:
			log.Fatal("The OS ", runtime.GOOS, " is not supported!")
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		switch runtime.GOOS {
		case "windows":
		case "darwin":
		case "linux":
			usr_local_bin := "/usr/local/bin"
			go_bin := "/usr/local/go/bin"
			entries := utils.ReadDir(go_bin)
			for _, entry := range entries {
				file := filepath.Join(go_bin, entry.Name())
				log.Printf(
					"Creating a symbolic link of %s into %s/ ...", file, usr_local_bin,
				)
				cmd := utils.Format("{prefix} ln -svf {file} {usr_local_bin}/", map[string]string{
					"prefix":        prefix,
					"file":          file,
					"usr_local_bin": usr_local_bin,
				})
				utils.RunCmd(cmd)
			}
		default:
			log.Fatal("The OS ", runtime.GOOS, " is not supported!")
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
	// rootCmd.AddCommand(golangCmd)
}
