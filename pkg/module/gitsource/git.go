package gitsource

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/github"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compatattr"
	ocm "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	ocmv1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

func (g GitSource) FetchSource(ctx cpi.Context, path, repo, version string) (*ocm.Source, error) {
	ref, commit, err := g.getGitInfo(path)
	if err != nil {
		return nil, err
	}

	if repo == "" {
		if repo, err = g.determineRepositoryURL(); err != nil {
			return nil, err
		}
	}

	sourceType := "git"
	if !compatattr.Get(ctx) {
		sourceType = github.CONSUMER_TYPE
	}

	label, err := ocmv1.NewLabel(refLabel, ref, ocmv1.WithVersion(ocmVersion))
	if err != nil {
		return nil, err
	}

	access := github.New(repo, "", commit)

	sourceMeta := ocm.SourceMeta{
		Type: sourceType,
		ElementMeta: ocm.ElementMeta{
			Name:    ocmIdentityName,
			Version: version,
			Labels:  ocmv1.Labels{*label},
		},
	}

	return &ocm.Source{
		SourceMeta: sourceMeta,
		Access:     access,
	}, nil
}

func (g GitSource) determineRepositoryURL() (string, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return "", fmt.Errorf("could not open git repository: %w", err)
	}

	remotes, err := r.Remotes()
	if err != nil || len(remotes) == 0 {
		return "", fmt.Errorf("could not get git remotes: %w", err)
	}

	httpURL := remotes[0].Config().URLs[0]
	if strings.HasPrefix(httpURL, "git") {
		httpURL = strings.Replace(httpURL, ":", "/", 1)
		httpURL = strings.Replace(httpURL, "git@", "https://", 1)
		httpURL = strings.TrimSuffix(httpURL, gitFolder)
	}
	repoURL, err := url.Parse(httpURL)
	if err != nil {
		return "", fmt.Errorf("could not parse repository URL: %w", err)
	}
	return repoURL.String(), nil
}

func (g GitSource) getGitInfo(gitPath string) (string, string, error) {
	if gitPath == "" {
		return "", "", fmt.Errorf("could not get git information, the path is empty")

	}

	if gitPath == string(filepath.Separator) {
		return "", "", errNotGit
	}

	r, err := git.PlainOpen(gitPath)

	if err != nil {
		return "", "", fmt.Errorf("could not open git repository: %w", err)
	}

	head, err := r.Head()
	if err != nil {
		return "", "", fmt.Errorf("could not get git head information: %w", err)
	}

	return head.Name().String(), head.Hash().String(), nil
}
