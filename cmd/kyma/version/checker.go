package version

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-project/cli/internal/nice"
	"io"
	"net/http"
	"regexp"
)

const (
	GitHubAPIEndpoint = "https://api.github.com/repos/kyma-project/cli/releases/latest"
)

type LatestGitTag struct {
	Name string `json:"tag_name"`
}

func CheckForStableRelease() {
	response, err := http.Get(GitHubAPIEndpoint)
	if err != nil {
		return
	}
	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	var latestGitTag LatestGitTag
	if err := json.Unmarshal(responseData, &latestGitTag); err != nil {
		return
	}

	matched, err := regexp.MatchString("[0-9]+[.][0-9]+[.][0-9]+", Version)
	if err != nil {
		return
	} else if matched && Version < latestGitTag.Name {
		nicePrint := nice.Nice{}
		nicePrint.PrintImportantf("CAUTION: You're using an outdated version of the Kyma CLI (%s)."+
			" The latest stable version is: %s", Version, latestGitTag.Name)
		fmt.Println()
	}
}
