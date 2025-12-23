package bigdata

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v3"
	"legendu.net/icon/cmd/icon"
	"legendu.net/icon/utils"
)

func extractMajorVersion(version string) string {
	index := strings.Index(version, ".")
	if index > 0 {
		return version[:index]
	}
	return version
}

// Get the recommended downloading URL for Spark.
func getSparkDownloadURL(sparkVersion, hadoopVersion string) (string, error) {
	hadoopVersion = extractMajorVersion(hadoopVersion)
	url := "https://www.apache.org/dyn/closer.lua/spark/spark-%s/spark-%s-bin-hadoop%s%s.tgz"
	suffix := ""
	const firstSparkConnectVersion = 4
	if utils.ParseInt(extractMajorVersion(sparkVersion)) >= firstSparkConnectVersion {
		suffix = "-connect"
	}
	url = fmt.Sprintf(url, sparkVersion, sparkVersion, hadoopVersion, suffix)
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to create a HTTP GET request to the URL '%s' with context: %w", url, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("the HTTP GET request to the URL '%s' failed: %w", url, err)
	}
	defer resp.Body.Close()
	if utils.IsErrorHTTPResponse(resp) {
		return "", fmt.Errorf("the HTTP GET request got an error response with the status code %d: %w", resp.StatusCode, err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read the body of the response of the HTTP GET request: %w", err)
	}
	html := string(body)
	begin := strings.Index(html, "<strong>")
	if begin < 0 {
		return "", fmt.Errorf("the HTML does NOT contain the tag <strong>\n%s", html)
	}
	html = html[begin+8:]
	end := strings.Index(html, "</strong>")
	if end < 0 {
		return "", fmt.Errorf("the HTML does NOT contain the tag </strong>\n%s", html)
	}
	return html[:end], nil
}

func extractHdp(url string) string {
	return url[strings.LastIndex(url, "/")+1 : strings.LastIndex(url, ".")]
}

type SparkHadoopVersion struct {
	Spark  string `yaml:"spark"`
	Hadoop string `yaml:"hadoop"`
}

func chooseSparkVersion(versions []SparkHadoopVersion) SparkHadoopVersion {
	for idx, version := range versions {
		fmt.Printf("%d: Spark version - %s, Hadoop version - %s\n", idx, version.Spark, version.Hadoop)
	}
	fmt.Print("Please enter the index corresponding to the Spark/Hadoop versions to install: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading input: %v", err)
	}
	return versions[utils.ParseInt(strings.TrimSpace(input))]
}

func readSparkHadoopVersion(interactive bool) SparkHadoopVersion {
	file := "~/.config/icon-data/spark/version.yaml"
	if !utils.ExistsFile(file) {
		log.Fatalf("Spar/Hadoop versions are not specified or configured (%s).", file)
	}
	var versions []SparkHadoopVersion
	err := yaml.Unmarshal(utils.ReadFile(file), &versions)
	if err != nil {
		log.Fatalf("Error unmarshaling data: %v", err)
	}
	switch len(versions) {
	case 0:
		log.Fatalf("No Spark versions configured in %s", file)
	case 1:
		return versions[0]
	}
	if interactive {
		return chooseSparkVersion(versions)
	}
	return versions[0]
}

// Install and configure Spark.
func spark(cmd *cobra.Command, _ []string) {
	// installation location
	dir := utils.GetStringFlag(cmd, "directory")
	// Spark/Hadoop version
	sparkVersion := utils.GetStringFlag(cmd, "spark-version")
	hadoopVersion := utils.GetStringFlag(cmd, "hadoop-version")
	if (sparkVersion == "") != (hadoopVersion == "") {
		log.Fatal("Either both of spark/hadoop versions or neither of them should be specified!")
	}
	if sparkVersion == "" {
		version := readSparkHadoopVersion(utils.GetBoolFlag(cmd, "interactive"))
		sparkVersion = version.Spark
		hadoopVersion = version.Hadoop
	}
	url, err := getSparkDownloadURL(sparkVersion, hadoopVersion)
	if err != nil {
		log.Fatal(err)
	}
	sparkHdp := extractHdp(url)
	sparkHome := filepath.Join(dir, sparkHdp)
	// install Spark
	prefix := utils.GetCommandPrefix(false, map[string]uint32{
		dir: unix.W_OK | unix.R_OK,
	})
	if utils.GetBoolFlag(cmd, "install") {
		sparkTgz, err := utils.DownloadFile(url, "spark.tgz", true)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Installing Spark into the directory %s ...\n", sparkHome)
		cmd := utils.Format("{prefix} mkdir -p {dir} && {prefix} tar -zxf {sparkTgz} -C {dir} && rm {sparkTgz}", map[string]string{
			"prefix":   prefix,
			"dir":      dir,
			"sparkTgz": sparkTgz,
		})
		utils.RunCmd(cmd)
	}
	if utils.GetBoolFlag(cmd, "config") {
		icon.FetchConfigData(false, "")

		metastoreDB := filepath.Join(sparkHome, "metastoreDb")
		warehouse := filepath.Join(sparkHome, "warehouse")
		cmd := utils.Format(
			`{prefix} mkdir -p {metastoreDb} && 
				{prefix} chmod -R 777 {metastoreDb} &&
				{prefix} mkdir -p {warehouse} &&
				{prefix} chmod -R 777 {warehouse}
				`,
			map[string]string{
				"prefix":      prefix,
				"metastoreDb": metastoreDB,
				"warehouse":   warehouse,
			})
		utils.RunCmd(cmd)
		// spark-defaults.conf
		text := utils.ReadFileAsString("~/.config/icon-data/spark/spark-defaults.conf")
		cmd = utils.Format("echo '{conf}' | {prefix} tee {sparkDefaults} > /dev/null",
			map[string]string{
				"prefix":        prefix,
				"conf":          strings.ReplaceAll(text, "$SPARK_HOME", sparkHome),
				"sparkDefaults": filepath.Join(sparkHome, "conf", "spark-defaults.conf"),
			},
		)
		utils.RunCmd(cmd)
		log.Printf(
			"Spark is configured to use %s as the metastore database and %s as the Hive warehouse.",
			metastoreDB, warehouse,
		)
		// create databases and tables
		/*
		   if schemaDir:
		       createDbs(sparkHome, schemaDir)
		   if not isWin():
		       runCmd(f"{args.prefix} chmod -R 777 {metastoreDb}")
		*/
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
		/*
			cmd = f"{args.prefix} rm -rf {sparkHome}"
			runCmd(cmd)
		*/
	}
}

var sparkCmd = &cobra.Command{
	Use:     "spark",
	Aliases: []string{},
	Short:   "Install and configure Spark.",
	//Args:  cobra.ExactArgs(1),
	Run: spark,
}

func ConfigSparkCmd(rootCmd *cobra.Command) {
	/*
	   subparser.add_argument(
	       "-s",
	       "--schema",
	       "--schema-dir",
	       dest="schema_dir",
	       type=Path,
	       default=None,
	       help="The path to a directory containing schema information." \
	           "The directory contains subdirs whose names are databases to create." \
	           "Each of those subdirs (database) contain SQL files of the format db.table.sql" \
	           "which containing SQL code for creating tables."
	   )
	*/
	sparkCmd.Flags().String("spark-version", "", "The version of Spark version to install.")
	sparkCmd.Flags().String("hadoop-version", "", "The version of Spark version to install.")
	sparkCmd.Flags().Bool("interactive", false, "Choose Spark/Hadoop versions interactively.")
	sparkCmd.Flags().StringP("directory", "d", "/opt", "The directory to install Spark.")
	sparkCmd.Flags().BoolP("install", "i", false, "Install Spark.")
	sparkCmd.Flags().BoolP("uninstall", "u", false, "Uninstall Spark.")
	sparkCmd.Flags().BoolP("config", "c", false, "Configure Spark.")
	sparkCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	sparkCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	rootCmd.AddCommand(sparkCmd)
}
