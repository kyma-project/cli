package releases

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

type release struct {
	TagName   string `json:"tag_name"`
	CreatedAt string `json:"created_at"`
}

const kymaReleaseURL = "https://api.github.com/repos/kyma-project/kyma/releases"

//NewCmd creates a new releases command
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "releases",
		Short: "List Kyma releases",
		RunE:  listReleases,
	}
	return cmd
}

func listReleases(cmd *cobra.Command, args []string) error {
	var myClient = &http.Client{Timeout: 10 * time.Second}
	var releasesList []release
	var err error
	resp, err := myClient.Get(kymaReleaseURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&releasesList)
	if err != nil {
		return err
	}
	for _, r := range releasesList {
		parsedTime, err := time.Parse(time.RFC3339, r.CreatedAt)
		if err != nil {
			return err
		}
		timeString := fmt.Sprintf("%d-%d-%d", parsedTime.Day(), parsedTime.Month(), parsedTime.Year())
		fmt.Printf("%s -> %s\n", r.TagName, timeString)
	}
	return nil
}
