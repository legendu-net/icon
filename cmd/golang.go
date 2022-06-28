package cmd

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

func getGolangVersion() string {
	resp, err := http.Get("https://github.com/golang/go/tags")
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode > 399 {
		log.Fatal("...")
	}
	html := string(body)
	re := regexp.MustCompile(`tag/go(\d+\.\d+\.\d+)`)
	return re.FindStringSubmatch(html)[1]
}

// Install and configure Golang.
func golang(cmd *cobra.Command, args []string) {
	install, err := cmd.Flags().GetBool("install")
	if err != nil {
		log.Fatal(err)
	}
	sudo, err := cmd.Flags().GetBool("sudo")
	if err != nil {
		log.Fatal(err)
	}
	prefix := ""
	if sudo {
		prefix = "sudo"
	}
	if install {
		switch runtime.GOOS {
		case "windows":
		case "darwin":
			// brew_install_safe("go")
		case "linux":
			ver := getGolangVersion()
			url := strings.ReplaceAll("https://go.dev/dl/go{ver}.linux-amd64.tar.gz", "{ver}", ver)
			goTgz := utils.DownloadFile(url, "go_*.tar.gz").Name()
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
	config, err := cmd.Flags().GetBool("config")
	if err != nil {
		log.Fatal(err)
	}
	if config {
		switch runtime.GOOS {
		case "windows":
		case "darwin":
		case "linux":
			usr_local_bin := "/usr/local/bin/"
			files, err := os.ReadDir("/usr/local/go/bin/")
			if err != nil {
				log.Fatal(err)
			}
			for _, file := range files {
				log.Printf(
					"Creating a symbolic link of %s into %s/ ...", file, usr_local_bin,
				)
				cmd := utils.Format("{prefix} ln -svf {file} {usr_local_bin}/", map[string]string{
					"prefix":        prefix,
					"file":          file.Name(),
					"usr_local_bin": usr_local_bin,
				})
				utils.RunCmd(cmd)
			}
		default:
			log.Fatal("The OS ", runtime.GOOS, " is not supported!")
		}
	}
}

var golangCmd = &cobra.Command{
	Use:     "golang",
	Aliases: []string{"go"},
	Short:   "Install and configure Golang.",
	//Args:  cobra.ExactArgs(1),
	Run: golang,
}

func init() {
	golangCmd.Flags().BoolP("install", "i", false, "If specified, install Golang.")
	golangCmd.Flags().BoolP("config", "c", false, "If specified, configure Golang.")
	rootCmd.AddCommand(golangCmd)
}
