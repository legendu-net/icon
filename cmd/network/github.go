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
func getReleaseUrl(repo string) string {
	if strings.HasSuffix(repo, ".git") {
		repo = repo[:len(repo)-4]
	}
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

type AssetInfo struct {
	Name               string `json:"name"`
	BrowserDownloadUrl string `json:"browser_download_url"`
}

type ReleaseInfo struct {
	TagName string      `json:"tag_name"`
	Assets  []AssetInfo `json:"assets"`
}

func assetNameContainKeywords(name string, keywords []string, keyworkdsExclude []string) bool {
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

func filterReleases(url string, constraint string) ReleaseInfo {
	var releases []ReleaseInfo
	json.Unmarshal(utils.HttpGetAsBytes(url), &releases)
	c := version.NewConstrainGroupFromString(constraint)
	for _, release := range releases {
		if c.Match(release.TagName) {
			return release
		}
	}
	log.Fatal("No release matching the version constraint is found!")
	return ReleaseInfo{}
}

func getLatestRelease(releaseUrl string) ReleaseInfo {
	url := releaseUrl + "/latest"
	var releaseInfo ReleaseInfo
	json.Unmarshal(utils.HttpGetAsBytes(url), &releaseInfo)
	return releaseInfo
}

// Download a release from GitHub.
// @param args: The arguments to parse.
// If None, the arguments from command-line are parsed.
func DownloadGitHubRelease(cmd *cobra.Command, args []string) {
	repo := utils.GetStringFlag(cmd, "repo")
	// get the version to download
	version := utils.GetStringFlag(cmd, "version")
	// form the release URL
	releaseUrl := getReleaseUrl(repo)
	var releaseInfo ReleaseInfo
	if version == "" {
		releaseInfo = getLatestRelease(releaseUrl)
	} else {
		releaseInfo = filterReleases(releaseUrl, version)
	}
	// parse browser download url
	keywords := utils.GetStringSliceFlag(cmd, "kwd")
	keywordsExclude := utils.GetStringSliceFlag(cmd, "KWD")
	var browserDownloadUrl string
	for _, assert := range releaseInfo.Assets {
		if assetNameContainKeywords(assert.Name, keywords, keywordsExclude) {
			browserDownloadUrl = assert.BrowserDownloadUrl
		}
	}
	// download the asset
	utils.DownloadFile(browserDownloadUrl, utils.GetStringFlag(cmd, "output"), false)
}

var DownloadGitHubReleaseCmd = &cobra.Command{
	Use:     "download_github_release",
	Aliases: []string{"download_github", "from_github", "github_release"},
	Short:   "Download file from GitHub.",
	//Args:  cobra.ExactArgs(1),
	Run: DownloadGitHubRelease,
}

func init() {
	DownloadGitHubReleaseCmd.Flags().StringP("repo", "r", "", "A GitHub repo of the form 'user_name/repo_name'.")
	err := DownloadGitHubReleaseCmd.MarkFlagRequired("repo")
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	DownloadGitHubReleaseCmd.Flags().StringP("version", "v", "", "The version of the release.")
	DownloadGitHubReleaseCmd.Flags().StringSliceP("kwd", "k", []string{}, "Keywords that the assert's name contains.")
	DownloadGitHubReleaseCmd.Flags().StringSliceP("KWD", "K", []string{}, "Keywords that the assert's name contains.")
	err = DownloadGitHubReleaseCmd.MarkFlagRequired("kwd")
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	DownloadGitHubReleaseCmd.Flags().StringP("output", "o", "", "The output path for the downloaded asset.")
	err = DownloadGitHubReleaseCmd.MarkFlagRequired("output")
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	// rootCmd.AddCommand(downloadgitHubReleaseCmd)
}
