package network

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/mcuadros/go-version"
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Get the release URL of a project on GitHub.
// @param repo: The repo name of the project on GitHub.
// return: The release URL of the project on GitHub.
func getReleaseURL(repo string) string {
	repo = strings.TrimSuffix(repo, ".git")
	if strings.HasPrefix(repo, "https://api.") {
		return repo
	}
	if strings.HasPrefix(repo, "https://") {
		lastIndex := strings.LastIndex(repo, "/")
		index := strings.LastIndex(repo[0:lastIndex], "/")
		repo = repo[index+1:]
	} else if strings.HasPrefix(repo, "git@") {
		index := strings.LastIndex(repo, ":")
		repo = repo[index+1:]
	}
	return "https://api.github.com/repos/" + repo + "/releases"
}

type assetInfo struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type releaseInfo struct {
	TagName string      `json:"tag_name"`
	Assets  []assetInfo `json:"assets"`
}

func assetNameContainKeywords(name string, keywords, keyworkdsExclude []string) bool {
	for _, keyword := range keywords {
		if !strings.Contains(name, keyword) {
			return false
		}
	}
	for _, keyword := range keyworkdsExclude {
		if strings.Contains(name, keyword) {
			return false
		}
	}
	return true
}

func filterReleases(url, constraint string) releaseInfo {
	log.Printf("Extracting release from %s with the constraint %s", url, constraint)
	const numRetry = 3
	const initialWaitingSeconds = 120
	bytes, err := utils.HTTPGetAsBytes(url, numRetry, initialWaitingSeconds)
	if err != nil {
		log.Fatal(err)
	}
	var releases []releaseInfo
	err = json.Unmarshal(bytes, &releases)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}
	c := version.NewConstrainGroupFromString(constraint)
	for _, release := range releases {
		if c.Match(release.TagName) {
			return release
		}
	}
	log.Fatal("No release matching the version constraint is found!")
	return releaseInfo{}
}

func getLatestRelease(releaseURL string) releaseInfo {
	url := releaseURL + "/latest"
	const numRetry = 3
	const initialWaitingSeconds = 120
	bytes, err := utils.HTTPGetAsBytes(url, numRetry, initialWaitingSeconds)
	if err != nil {
		log.Fatal(err)
	}
	var release releaseInfo
	err = json.Unmarshal(bytes, &release)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return release
}

// Download a release from GitHub.
// @param args: The arguments to parse.
// If None, the arguments from command-line are parsed.
func downloadGitHubReleaseArgs(cmd *cobra.Command, _ []string) {
	DownloadGitHubRelease(
		utils.GetStringFlag(cmd, "repo"),
		utils.GetStringFlag(cmd, "version"),
		map[string][]string{"common": utils.GetStringSliceFlag(cmd, "kwd")},
		utils.GetStringSliceFlag(cmd, "KWD"),
		utils.GetStringFlag(cmd, "output"),
	)
}

// Download a release from GitHub.
// @param args: The arguments to parse.
// If None, the arguments from command-line are parsed.
func DownloadGitHubRelease(repo, ver string, keywords map[string][]string, keywordsExclude []string, output string) {
	keywords_ := utils.BuildKernelOSKeywords(keywords)
	log.Printf(`Download release from the GitHub repository %s satisfying the following conditions:
	Version: %s
	Contains: %s
	Does not contain: %s
	Write to: %s
	`, repo, ver, strings.Join(keywords_, ", "), strings.Join(keywordsExclude, ", "), output)
	// form the release URL
	releaseURL := getReleaseURL(repo)
	log.Printf("Release URL: %s\n", releaseURL)
	var release releaseInfo
	if ver == "" {
		release = getLatestRelease(releaseURL)
	} else {
		release = filterReleases(releaseURL, ver)
	}
	// parse browser download url
	var browserDownloadURL string
	for _, asset := range release.Assets {
		if assetNameContainKeywords(asset.Name, keywords_, keywordsExclude) {
			log.Printf("Asset %s is matched.", asset.Name)
			browserDownloadURL = asset.BrowserDownloadURL
			break
		} else {
			log.Printf("Asset %s is not matched.", asset.Name)
		}
	}
	// download the asset
	_, err := utils.DownloadFile(browserDownloadURL, output, false)
	if err != nil {
		log.Fatal(err)
	}
}

var downloadGitHubReleaseCmd = &cobra.Command{
	Use:     "download_github_release",
	Aliases: []string{"download_github", "from_github", "github_release"},
	Short:   "Download file from GitHub.",
	//Args:  cobra.ExactArgs(1),
	Run: downloadGitHubReleaseArgs,
}

func ConfigDownloadGitHubReleaseCmd(rootCmd *cobra.Command) {
	downloadGitHubReleaseCmd.Flags().StringP("repo", "r", "", "A GitHub repo of the form 'user_name/repo_name'.")
	err := downloadGitHubReleaseCmd.MarkFlagRequired("repo")
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	downloadGitHubReleaseCmd.Flags().StringP("version", "v", "", "The version of the release.")
	downloadGitHubReleaseCmd.Flags().StringSliceP("kwd", "k", []string{}, "Keywords that the asset's name contains.")
	downloadGitHubReleaseCmd.Flags().StringSliceP("KWD", "K", []string{}, "Keywords that the asset's name must not contain.")
	err = downloadGitHubReleaseCmd.MarkFlagRequired("kwd")
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	downloadGitHubReleaseCmd.Flags().StringP("output", "o", "", "The output path for the downloaded asset.")
	err = downloadGitHubReleaseCmd.MarkFlagRequired("output")
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	rootCmd.AddCommand(downloadGitHubReleaseCmd)
}
