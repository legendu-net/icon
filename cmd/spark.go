package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"embed"
	"legendu.net/icon/utils"
	"github.com/spf13/cobra"
)

//go:embed data/spark/spark-defaults.conf
var spark_defaults embed.FS

// Get the latest version of Spark.
func get_version() string {
    log.Printf("Parsing the latest version of Spark...")
	resp, err := http.Get("https://spark.apache.org/downloads.html")
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	html := string(body)
    re := regexp.MustCompile(`Latest Release \(Spark (\d.\d.\d)\)`)
	for _, line := range strings.Split(html, "\n") {
		match := re.FindString(line)
		if match != "" {
			return match
		}
	}
	return ""
}

// Get the recommended downloading URL for Spark. 
func get_download_url(spark_version string, hadoop_version string) string {
	// TODO: substitute Spark/Hadoop versions
	resp, err := http.Get("https://www.apache.org/dyn/closer.lua/spark/spark-3.3.0/spark-3.3.0-bin-hadoop3.tgz")
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
	html = html[strings.Index(html, "<strong>")+8:]
	return html[:strings.Index(html, "</strong>")]
}

// Download Spark.
func download(url string) *os.File {
	// You can get the download URL from: https://archive.apache.org/dist/spark/spark-3.3.0/
	// url = "https://archive.apache.org/dist/spark/spark-3.3.0/spark-3.3.0-bin-hadoop3.tgz"
	log.Printf("Downloading Spark from: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	// create a temp file to receive the download
	out, err := os.CreateTemp(os.TempDir(), "spark_*.tgz")
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(out, resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Spark has been downloaded to %s", out.Name())
	return out
}

// Install and configure Spark.
func spark(cmd *cobra.Command, args []string) {
	// get Spark version
	spark_version, err := cmd.Flags().GetString("spark-version")
	if err != nil {
		log.Fatal(err)
	}
    if spark_version == "" {
        spark_version = get_version()
	}
	// get Hadoop version
	hadoop_version := "3"
	/*
	hadoop_version, err := cmd.Flags().GetString("hadoop-version")
	if err != nil {
		log.Fatal(err)
	}
    if hadoop_version == "" {
        hadoop_version = get_version()
	}
	*/
    // installation location
	dir, err := cmd.Flags().GetString("directory")
	if err != nil {
		log.Fatal(err)
	}
    spark_hdp := fmt.Sprintf("spark-%s-bin-hadoop%s", spark_version, hadoop_version)
    spark_home := filepath.Join(dir, spark_hdp)
	// install Spark
	install, err := cmd.Flags().GetBool("install")
	if err != nil {
		log.Fatal(err)
	}
	prefix := "sudo"
    if install {
		spark_tgz := download("https://archive.apache.org/dist/spark/spark-3.3.0/spark-3.3.0-bin-hadoop3.tgz")
		log.Printf("Installing Spark into the directory %s ...\n", spark_home)
		switch runtime.GOOS {
			case "windows":
			default:
				cmd := utils.Format("{prefix} mkdir -p {dir} && {prefix} tar -zxf {spark_tgz} -C {dir} && rm {spark_tgz}", map[string]string {
					"prefix": prefix, 
					"dir": dir, 
					"spark_tgz": spark_tgz.Name(), 
				})
				utils.RunCmd(cmd)
		}
	}
	config, err := cmd.Flags().GetBool("config")
	if err != nil {
		log.Fatal(err)
	}
    if config {
        metastore_db := filepath.Join(spark_home, "metastore_db")
        warehouse := filepath.Join(spark_home, "warehouse")
		switch runtime.GOOS {
			case "windows":
				os.MkdirAll(metastore_db, 0750)
				os.MkdirAll(warehouse, 0750)
			default:
				cmd := utils.Format(
					`{prefix} mkdir -p {metastore_db} && 
					{prefix} chmod -R 777 {metastore_db} &&
					{prefix} mkdir -p {warehouse} &&
					{prefix} chmod -R 777 {warehouse}
					`,
					map[string]string {
						"prefix": prefix,
						"metastore_db": metastore_db,
						"warehouse": warehouse,
				})
				utils.RunCmd(cmd)
				// spark-defaults.conf
				bytes, err := spark_defaults.ReadFile("data/spark/spark-defaults.conf")
				if err != nil {
					log.Fatal(err)
				}
				cmd = utils.Format("echo '{conf}' | {prefix} tee {spark_defaults} > /dev/null",
					map[string]string {
						"prefix": prefix,
						"conf": strings.ReplaceAll(string(bytes), "$SPARK_HOME", spark_home),
						"spark_defaults": filepath.Join(spark_home, "conf/spark-defaults.conf"),
					},
				)
				utils.RunCmd(cmd)
		}
        log.Printf(
            "Spark is configured to use %s as the metastore database and %s as the Hive warehouse.",
            metastore_db, warehouse,
        )
        // create databases and tables
		/*
        if args.schema_dir:
            create_dbs(spark_home, args.schema_dir)
        if not is_win():
            run_cmd(f"{args.prefix} chmod -R 777 {metastore_db}")
		*/
	}
	/*
    if args.uninstall:
        cmd = f"{args.prefix} rm -rf {spark_home}"
        run_cmd(cmd)
		*/
}

var sparkCmd = &cobra.Command{
    Use:   "spark",
    Aliases: []string{},
    Short:  "Install and configure Spark.",
    //Args:  cobra.ExactArgs(1),
    Run: spark,
}

func init() {
	/*
    subparser.add_argument(
        "--loc",
        "--location",
        dest="location",
        type=Path,
        default=Path(),
        help="The location (current work directory, by default) to install Spark to."
    )
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
	sparkCmd.Flags().String("spark-version", "3.3.0", "The version of Spark version to install.")
	sparkCmd.Flags().String("hadoop-version", "3.2", "The version of Spark version to install.")
	sparkCmd.Flags().StringP("directory", "d", "/opt", "The directory to install Spark.")
	sparkCmd.Flags().BoolP("install", "i", false, "If specified, install Spark.")
	sparkCmd.Flags().BoolP("config", "c", false, "If specified, configure Spark.")
    rootCmd.AddCommand(sparkCmd)
}
