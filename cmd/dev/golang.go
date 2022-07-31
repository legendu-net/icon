package dev

import (
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"legendu.net/icon/utils"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
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

// Install and configure Golang.
func golang(cmd *cobra.Command, args []string) {
	prefix := utils.GetCommandPrefix(false, map[string]uint32{
		"/usr/local/go":  unix.W_OK | unix.R_OK,
		"/usr/local":     unix.W_OK | unix.R_OK,
		"/usr/local/bin": unix.W_OK | unix.R_OK,
	})
	if utils.GetBoolFlag(cmd, "install") {
		switch runtime.GOOS {
		case "windows":
		case "darwin":
			// brew_install_safe("go")
		case "linux":
			ver := getGolangVersion()
			url := strings.ReplaceAll("https://go.dev/dl/go{ver}.linux-amd64.tar.gz", "{ver}", ver)
			goTgz := utils.DownloadFile(url, "go_*.tar.gz", true)
			cmd := utils.Format(`{prefix} rm -rf /usr/local/go \
						&& {prefix} tar -C /usr/local/ -xzf {goTgz}`,
				map[string]string{
					"prefix": prefix,
					"goTgz":  goTgz,
				},
			)
			utils.RunCmd(cmd)
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
	GolangCmd.Flags().BoolP("config", "c", false, "Configure Golang.")
	// rootCmd.AddCommand(golangCmd)
}
