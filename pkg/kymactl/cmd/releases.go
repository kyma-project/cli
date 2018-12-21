package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

//ReleasesOptions defines available options for the command
type ReleasesOptions struct {
	GithubToken string
}

//NewReleasesOptions creates options with default values
func NewReleasesOptions() *ReleasesOptions {
	return &ReleasesOptions{}
}

//NewReleasesCmd creates a new releases command
func NewReleasesCmd(o *ReleasesOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "releases",
		Short: "List available Kyma releases",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	cmd.Flags().StringVarP(&o.GithubToken, "tokens", "t", "", "Github token to use for querying for releases. In case of rate limiting problems")
	return cmd
}

//Run runs the command
func (o *ReleasesOptions) Run() error {
	releasesList, err := retrieveReleases(o)
	if err != nil {
		return err
	}

	fmt.Printf("Name   Date\n")
	for _, r := range releasesList {
		if *r.Draft {
			continue
		}
		if *r.Prerelease {
			continue
		}
		if *r.Name == "" {
			continue
		}

		fmt.Printf("%s  %d.%d.%d\n", *r.Name, (*r.PublishedAt).Year(), (*r.PublishedAt).Month(), (*r.PublishedAt).Day())
	}
	return nil
}

func retrieveReleases(o *ReleasesOptions) ([]*github.RepositoryRelease, error) {
	ctx := context.Background()
	var oAuthClient *http.Client
	if o.GithubToken != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: o.GithubToken},
		)
		oAuthClient = oauth2.NewClient(ctx, ts)
	}
	client := github.NewClient(oAuthClient)

	releases, _, err := client.Repositories.ListReleases(ctx, "kyma-project", "kyma", nil)
	if err != nil {
		return nil, err
	}

	return releases, nil
}
