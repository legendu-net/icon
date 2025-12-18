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
func GetReleaseURL(repo string) string {
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

type AssetInfo struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type ReleaseInfo struct {
	TagName string      `json:"tag_name"`
	Assets  []AssetInfo `json:"assets"`
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

func filterReleases(url, constraint string) ReleaseInfo {
	log.Printf("Extracting release from %s with the constraint %s", url, constraint)
	bytes, err := utils.HTTPGetAsBytes(url, 3, 120)
	if err != nil {
		log.Fatal(err)
	}
	var releases []ReleaseInfo
	err = json.Unmarshal(bytes, &releases)
	if err != nil {
		log.Fatalf("Failed to parse TOML: %v", err)
	}
	c := version.NewConstrainGroupFromString(constraint)
	for _, release := range releases {
		if c.Match(release.TagName) {
			return release
		}
	}
	log.Fatal("No release matching the version constraint is found!")
	return ReleaseInfo{}
}

func GetLatestRelease(releaseURL string) ReleaseInfo {
	url := releaseURL + "/latest"
	bytes, err := utils.HTTPGetAsBytes(url, 3, 120)
	if err != nil {
		log.Fatal(err)
	}
	var releaseInfo ReleaseInfo
	err = json.Unmarshal(bytes, &releaseInfo)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return releaseInfo
}

// Download a release from GitHub.
// @param args: The arguments to parse.
// If None, the arguments from command-line are parsed.
func DownloadGitHubReleaseArgs(cmd *cobra.Command, _ []string) {
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
	releaseURL := GetReleaseURL(repo)
	log.Printf("Release URL: %s\n", releaseURL)
	var releaseInfo ReleaseInfo
	if ver == "" {
		releaseInfo = GetLatestRelease(releaseURL)
	} else {
		releaseInfo = filterReleases(releaseURL, ver)
	}
	// parse browser download url
	var browserDownloadURL string
	for _, assert := range releaseInfo.Assets {
		if assetNameContainKeywords(assert.Name, keywords_, keywordsExclude) {
			log.Printf("Assert %s is matched.", assert.Name)
			browserDownloadURL = assert.BrowserDownloadURL
			break
		} else {
			log.Printf("Assert %s is not matched.", assert.Name)
		}
	}
	// download the asset
	_, err := utils.DownloadFile(browserDownloadURL, output, false)
	if err != nil {
		log.Fatal(err)
	}
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
}
