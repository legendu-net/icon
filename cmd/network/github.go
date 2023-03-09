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
func GetReleaseUrl(repo string) string {
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
	log.Printf("Extracting release from %s with the constraint %s", url, constraint)
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

func GetLatestRelease(releaseUrl string) ReleaseInfo {
	url := releaseUrl + "/latest"
	var releaseInfo ReleaseInfo
	err := json.Unmarshal(utils.HttpGetAsBytes(url), &releaseInfo)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return releaseInfo
}

// Download a release from GitHub.
// @param args: The arguments to parse.
// If None, the arguments from command-line are parsed.
func DownloadGitHubReleaseArgs(cmd *cobra.Command, args []string) {
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
func DownloadGitHubRelease(repo string, version string, keywords map[string][]string, keywordsExclude []string, output string) {
	keywords_ := utils.BuildKernelOSKeywords(keywords)
	log.Printf(`Download release from the GitHub repository %s satisfying the following conditions:
	Version: %s
	Contains: %s
	Does not contain: %s
	Write to: %s
	`, repo, version, strings.Join(keywords_, ", "), strings.Join(keywordsExclude, ", "), output)
	// form the release URL
	releaseUrl := GetReleaseUrl(repo)
	log.Printf("Release URL: %s\n", releaseUrl)
	var releaseInfo ReleaseInfo
	if version == "" {
		releaseInfo = GetLatestRelease(releaseUrl)
	} else {
		releaseInfo = filterReleases(releaseUrl, version)
	}
	// parse browser download url
	var browserDownloadUrl string
	for _, assert := range releaseInfo.Assets {
		if assetNameContainKeywords(assert.Name, keywords_, keywordsExclude) {
			log.Printf("Assert %s is matched.", assert.Name)
			browserDownloadUrl = assert.BrowserDownloadUrl
			break
		} else {
			log.Printf("Assert %s is not matched.", assert.Name)
		}
	}
	// download the asset
	utils.DownloadFile(browserDownloadUrl, output, false)
}

var DownloadGitHubReleaseCmd = &cobra.Command{
	Use:     "download_github_release",
	Aliases: []string{"download_github", "from_github", "github_release"},
	Short:   "Download file from GitHub.",
	//Args:  cobra.ExactArgs(1),
	Run: DownloadGitHubReleaseArgs,
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
