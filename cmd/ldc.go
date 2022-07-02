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

func getDockerImagePort(imageName string) int {
	if strings.HasPrefix(imageName, "dclong/") {
        imageName = imageName[7:]
        if strings.HasPrefix(imageName, "jupyterlab") {
            return 8888
		}
        if strings.HasPrefix(imageName, "jupyterhub") {
            return 8000
		}
        if strings.HasPrefix(imageName, "vscode") {
            return 8080
		}
	}
    return 0
}


func getDockerImageHostname(imageName string) string {
    start := strings.Index(imageName, "/") + 1
    end := strings.Index(imageName, ":")
    if end < 0 {
        end = len(imageName)
	}
    return imageName[start:end]
}

// Launch a Docker container.
func ldc(cmd *cobra.Command, args []string) {
	CURRENT_USER, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
    USER := currentUser.Username
    USER_ID := currentUser.Uid
    GROUP_ID := currentUser.Gid
    cmd := []string {
        "docker",
        "run",
		"-it",
        "--init",
        "--log-opt",
        "max-size=50m",
        "-e",
        "DOCKER_USER=" + USER,
        "-e",
        "DOCKER_USER_ID=" + USER_ID,
        "-e",
        "DOCKER_PASSWORD=" + USER,
        "-e",
        "DOCKER_GROUP_ID=" + GROUP_ID,
        "-e",
        "DOCKER_ADMIN_USER=" + USER,
        "-v",
        os.Getwd() + ":/workdir",
        "-v",
        os.UserHomeDir().parent + ":/home_host",
        "--hostname",
        _get_hostname(args.image_name[0]),
    }
	if detach {
		cmd[2] = "-d" 
	}
    if sys.platform == "linux":
        memory = os.sysconf("SC_PAGE_SIZE") * os.sysconf("SC_PHYS_PAGES")
        memory = int(memory * 0.8)
        cmd.append(f"--memory={memory}b")
        cpus = max(os.cpu_count() - 1, 1)
        cmd.append(f"--cpus={cpus}")
    port = _get_port(args.image_name[0])
    if port:
        cmd.append(f"--publish={args.port if args.port else port}:{port}")
    if args.extra_port_mappings:
        cmd.extend("-p " + mapping for mapping in args.extra_port_mappings)
    cmd.extend(args.image_name)
    if len(args.image_name) == 1 and args.image_name[0].startswith("dclong/"):
        cmd.append("/scripts/sys/init.sh")
    logger.debug(
        "Launching Docker container using the following command:\n{}", " ".join(cmd)
    )
    sp.run(cmd, check=True)
}

var ldcCmd = &cobra.Command{
	Use:     "ldc",
	Aliases: []string{},
	Short:   "Launch a container of a Docker image.",
	//Args:  cobra.ExactArgs(1),
	Run: ldc,
}

func init() {
	/*
    parser.add_argument(
        "image_name",
        nargs="+",
        help="The name (including tag) of the Docker image to launch."
    )
    parser.add_argument(
        "--extra-publish",
        "--extra-port-mappings",
        dest="extra_port_mappings",
        nargs="*",
        default=(),
        help="Extra port mappings."
    )
	*/
	ldcCmd.Flags().BoolP("detach", "d", false, "If specified, run container in background and print container ID.")
	ldcCmd.Flags().IntP("port", "p", 0, "The port on the Docker host to forward to the port inside the Docker container.")
	rootCmd.AddCommand(ldcCmd)
}
