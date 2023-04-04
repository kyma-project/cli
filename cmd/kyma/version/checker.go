package version

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-project/cli/internal/nice"
	"io"
	"net/http"
	"runtime/debug"
)

const (
	GitHubAPIEndpoint = "https://api.github.com/repos/kyma-project/cli/tags"
)

type LatestGitTag struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

// CheckForStableRelease Checks for new versions of the CLI
func CheckForStableRelease() {
	response, err := http.Get(GitHubAPIEndpoint)

	//	Any errors in API call shouldn't disrupt the usual CLI flow
	if err != nil {
		return
	}
	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	var githubTags []LatestGitTag
	err = json.Unmarshal(responseData, &githubTags)
	if err != nil {
		return
	}
	var stableCLI LatestGitTag
	for _, tag := range githubTags {
		if tag.Name == "stable" {
			stableCLI = tag
			break
		}
	}

	var stableVersion string
	for _, tag := range githubTags {
		if tag.Name != stableCLI.Name && tag.Commit.SHA == stableCLI.Commit.SHA {
			stableVersion = tag.Name
			break
		}
	}
	if stableVersion == "" {
		return
	}

	var currentCommitSHA string
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			currentCommitSHA = setting.Value
			break
		}
	}

	if currentCommitSHA == "" {
		return
	}

	nicePrint := nice.Nice{}
	if Version != stableVersion || currentCommitSHA != stableCLI.Commit.SHA {
		nicePrint.PrintImportantf("CAUTION: You're using an outdated version of the Kyma CLI (%s). The latest stable version is: %s", Version, stableVersion)
		fmt.Println()
	}
}
