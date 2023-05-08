package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/kyma-project/cli/internal/nice"
)

const (
	gitHubAPIEndpoint = "https://api.github.com/repos/kyma-project/cli/releases/latest"
)

type latestGitTag struct {
	Name string `json:"tag_name"`
}

func CheckForLatestRelease(currentVersion string) {
	response, err := http.Get(gitHubAPIEndpoint)
	if err != nil {
		return
	}
	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	var latestGitTag latestGitTag
	if err := json.Unmarshal(responseData, &latestGitTag); err != nil {
		return
	}

	matched, err := regexp.MatchString("[0-9]+[.][0-9]+[.][0-9]+", currentVersion)
	if err != nil || !matched {
		return
	}
	if currentVersion < latestGitTag.Name {
		nicePrint := nice.Nice{}
		nicePrint.PrintImportantf("CAUTION: You're using an outdated version of the Kyma CLI (%s)."+
			" The latest released version is: %s", currentVersion, latestGitTag.Name)
		fmt.Println()
	}
}
