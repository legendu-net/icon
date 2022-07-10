package cmd

import (
	"embed"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"legendu.net/icon/utils"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

//go:embed data/spark/spark-defaults.conf
var sparkDefaults embed.FS

// Get the latest version of Spark.
func getVersion() string {
	log.Printf("Parsing the latest version of Spark ...")
	resp, err := http.Get("https://spark.apache.org/downloads.html")
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode > 399 {
		log.Fatal("HTTP request got an error response with the status code ", resp.StatusCode)
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
func getSparkDownloadUrl(sparkVersion string, hadoopVersion string) string {
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
		log.Fatal("HTTP request got an error response with the status code ", resp.StatusCode)
	}
	html := string(body)
	html = html[strings.Index(html, "<strong>")+8:]
	return html[:strings.Index(html, "</strong>")]
}

// Install and configure Spark.
func spark(cmd *cobra.Command, args []string) {
	// get Spark version
	sparkVersion, err := cmd.Flags().GetString("spark-version")
	if err != nil {
		log.Fatal(err)
	}
	if sparkVersion == "" {
		sparkVersion = getVersion()
	}
	// get Hadoop version
	hadoopVersion := "3"
	/*
			hadoopVersion, err := cmd.Flags().GetString("hadoop-version")
			if err != nil {
				log.Fatal(err)
			}
		    if hadoopVersion == "" {
		        hadoopVersion = getVersion()
			}
	*/
	// installation location
	dir, err := cmd.Flags().GetString("directory")
	if err != nil {
		log.Fatal(err)
	}
	sparkHdp := fmt.Sprintf("spark-%s-bin-hadoop%s", sparkVersion, hadoopVersion)
	sparkHome := filepath.Join(dir, sparkHdp)
	// install Spark
	install, err := cmd.Flags().GetBool("install")
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	if install {
		sparkTgz := utils.DownloadFile("https://archive.apache.org/dist/spark/spark-3.3.0/spark-3.3.0-bin-hadoop3.tgz", "spark_*.tgz").Name()
		log.Printf("Installing Spark into the directory %s ...\n", sparkHome)
		switch runtime.GOOS {
		case "windows":
		default:
			cmd := utils.Format("mkdir -p {dir} && tar -zxf {sparkTgz} -C {dir} && rm {sparkTgz}", map[string]string{
				"dir":      dir,
				"sparkTgz": sparkTgz,
			})
			utils.RunCmd(cmd)
		}
	}
	config, err := cmd.Flags().GetBool("config")
	if err != nil {
		log.Fatal(err)
	}
	if config {
		metastoreDb := filepath.Join(sparkHome, "metastoreDb")
		warehouse := filepath.Join(sparkHome, "warehouse")
		switch runtime.GOOS {
		case "windows":
			os.MkdirAll(metastoreDb, 0750)
			os.MkdirAll(warehouse, 0750)
		default:
			cmd := utils.Format(
				`mkdir -p {metastoreDb} && chmod -R 777 {metastoreDb} &&
					mkdir -p {warehouse} && chmod -R 777 {warehouse}
					`,
				map[string]string{
					"metastoreDb": metastoreDb,
					"warehouse":   warehouse,
				})
			utils.RunCmd(cmd)
			// spark-defaults.conf
			bytes, err := sparkDefaults.ReadFile("data/spark/spark-defaults.conf")
			if err != nil {
				log.Fatal(err)
			}
			cmd = utils.Format("echo '{conf}' | tee {sparkDefaults} > /dev/null",
				map[string]string{
					"conf":          strings.ReplaceAll(string(bytes), "$SPARK_HOME", sparkHome),
					"sparkDefaults": filepath.Join(sparkHome, "conf/spark-defaults.conf"),
				},
			)
			utils.RunCmd(cmd)
		}
		log.Printf(
			"Spark is configured to use %s as the metastore database and %s as the Hive warehouse.",
			metastoreDb, warehouse,
		)
		// create databases and tables
		/*
		   if schemaDir:
		       createDbs(sparkHome, schemaDir)
		   if not isWin():
		       runCmd(f"chmod -R 777 {metastoreDb}")
		*/
	}
	/*
	   if args.uninstall:
	       cmd = f"rm -rf {sparkHome}"
	       runCmd(cmd)
	*/
}

var sparkCmd = &cobra.Command{
	Use:     "spark",
	Aliases: []string{},
	Short:   "Install and configure Spark.",
	//Args:  cobra.ExactArgs(1),
	Run: spark,
}

func init() {
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
	sparkCmd.Flags().String("spark-version", "3.3.0", "The version of Spark version to install.")
	sparkCmd.Flags().String("hadoop-version", "3.2", "The version of Spark version to install.")
	sparkCmd.Flags().StringP("directory", "d", "/opt", "The directory to install Spark.")
	sparkCmd.Flags().BoolP("install", "i", false, "If specified, install Spark.")
	sparkCmd.Flags().BoolP("config", "c", false, "If specified, configure Spark.")
	rootCmd.AddCommand(sparkCmd)
}
