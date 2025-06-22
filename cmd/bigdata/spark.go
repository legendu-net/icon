package bigdata

import (
	//"embed"
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"io"
	"legendu.net/icon/utils"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// Get the latest version of Spark.
func getSparkVersion() string {
	log.Printf("Parsing the latest version of Spark ...")
	html := utils.HttpGetAsString("https://spark.apache.org/downloads.html", 3, 120)
	re := regexp.MustCompile(`Latest Release \(Spark (\d.\d.\d)\)`)
	for _, line := range strings.Split(html, "\n") {
		match := re.FindString(line)
		if match != "" {
			return match
		}
	}
	return "4.0.0"
}

// Get the recommended downloading URL for Spark.
func getSparkDownloadUrl(sparkVersion string, hadoopVersion string, connect bool) string {
	url := "https://www.apache.org/dyn/closer.lua/spark/spark-%s/spark-%s-bin-hadoop%s%s.tgz"
	suffix := ""
	if sparkVersion >= "4.0.0" {
		if connect {
			suffix = "-connect"
		}
	}
	return fmt.Sprintf(url, sparkVersion, sparkVersion, hadoopVersion, suffix)
}

// Install and configure Spark.
func spark(cmd *cobra.Command, args []string) {
	// get Spark version
	sparkVersion, err := cmd.Flags().GetString("spark-version")
	if err != nil {
		log.Fatal(err)
	}
	//if sparkVersion == "" {
	//	sparkVersion = getSparkVersion()
	//}
	// get Hadoop version
	hadoopVersion, err := cmd.Flags().GetString("hadoop-version")
	if err != nil {
		log.Fatal(err)
	}
	// use Spark connect or not
	connect, err := cmd.Flags().GetBool("connect")
	if err != nil {
		log.Fatal(err)
	}
	// installation location
	dir, err := cmd.Flags().GetString("directory")
	if err != nil {
		log.Fatal(err)
	}
	sparkHdp := fmt.Sprintf("spark-%s-bin-hadoop%s", sparkVersion, hadoopVersion)
	sparkHome := filepath.Join(dir, sparkHdp)
	// install Spark
	prefix := utils.GetCommandPrefix(false, map[string]uint32{
		dir: unix.W_OK | unix.R_OK,
	})
	if utils.GetBoolFlag(cmd, "install") {
		sparkTgz := utils.DownloadFile(getSparkDownloadUrl(sparkVersion, hadoopVersion, connect), "spark_*.tgz", true)
		log.Printf("Installing Spark into the directory %s ...\n", sparkHome)
		switch runtime.GOOS {
		case "windows":
		default:
			cmd := utils.Format("{prefix} mkdir -p {dir} && {prefix} tar -zxf {sparkTgz} -C {dir} && rm {sparkTgz}", map[string]string{
				"prefix":   prefix,
				"dir":      dir,
				"sparkTgz": sparkTgz,
			})
			utils.RunCmd(cmd)
		}
	}
	if utils.GetBoolFlag(cmd, "config") {
		metastoreDb := filepath.Join(sparkHome, "metastoreDb")
		warehouse := filepath.Join(sparkHome, "warehouse")
		switch runtime.GOOS {
		case "windows":
			utils.MkdirAll(metastoreDb, 0777)
			utils.MkdirAll(warehouse, 0777)
		default:
			cmd := utils.Format(
				`{prefix} mkdir -p {metastoreDb} && 
					{prefix} chmod -R 777 {metastoreDb} &&
					{prefix} mkdir -p {warehouse} &&
					{prefix} chmod -R 777 {warehouse}
					`,
				map[string]string{
					"prefix":      prefix,
					"metastoreDb": metastoreDb,
					"warehouse":   warehouse,
				})
			utils.RunCmd(cmd)
			// spark-defaults.conf
			text := utils.ReadEmbeddedFileAsString("data/spark/spark-defaults.conf")
			cmd = utils.Format("echo '{conf}' | {prefix} tee {sparkDefaults} > /dev/null",
				map[string]string{
					"prefix":        prefix,
					"conf":          strings.ReplaceAll(text, "$SPARK_HOME", sparkHome),
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

var SparkCmd = &cobra.Command{
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
	SparkCmd.Flags().String("spark-version", "4.0.0", "The version of Spark version to install.")
	SparkCmd.Flags().String("hadoop-version", "3", "The version of Spark version to install.")
	SparkCmd.Flags().Bool("connect", true, "Install Spark Connect if supported.")
	SparkCmd.Flags().StringP("directory", "d", "/opt", "The directory to install Spark.")
	SparkCmd.Flags().BoolP("install", "i", false, "Install Spark.")
	SparkCmd.Flags().BoolP("uninstall", "u", false, "Uninstall Spark.")
	SparkCmd.Flags().BoolP("config", "c", false, "Configure Spark.")
	// rootCmd.AddCommand(sparkCmd)
}
