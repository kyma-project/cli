package version

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-project/cli/internal/nice"
	"io"
	"net/http"
	"runtime/debug"
)

type KymaCLIMetadata struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

// CheckForStableRelease Checks for new versions of the CLI
func CheckForStableRelease() {
	response, err := http.Get("https://api.github.com/repos/kyma-project/cli/tags")

	//	For any problems in fetching, reading or parsing the response from GitHub API, we simply ignore it
	//	and don't disrupt the usual CLI Flow
	if err != nil {
		return
	}
	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	var githubTags []KymaCLIMetadata
	err = json.Unmarshal(responseData, &githubTags)
	if err != nil {
		return
	}
	var stableCLI KymaCLIMetadata
	for _, tag := range githubTags {
		if tag.Name == "stable" {
			stableCLI = tag
			break
		}
	}

	var stableCLINumber string
	for _, tag := range githubTags {
		if tag.Name != stableCLI.Name && tag.Commit.SHA == stableCLI.Commit.SHA {
			stableCLINumber = tag.Name
		}
	}

	var currentCommitSHA string
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				currentCommitSHA = setting.Value
			}
		}
	}

	nicePrint := nice.Nice{}
	if Version != stableCLINumber || currentCommitSHA != stableCLI.Commit.SHA {
		nicePrint.PrintImportantf("CAUTION: You're using an outdated version of the Kyma CLI. The latest stable version is: %s", stableCLINumber)
		fmt.Println()
	}
}
