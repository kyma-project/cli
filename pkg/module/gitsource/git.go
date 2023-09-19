package gitsource

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/github"
	ocm "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	ocmv1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
)

func (g GitSource) FetchSource(path, repo, version string) (*ocm.Source, error) {
	ref, commit, err := g.getGitInfo(path)
	if err != nil {
		return nil, err
	}

	sourceType := github.CONSUMER_TYPE

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

func (g GitSource) DetermineRepositoryURL(gitRemote, repo, repoPath string) (string, error) {
	if repo != "" {
		return repo, nil
	}

	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", fmt.Errorf("could not open git repository: %w", err)
	}

	remotes, err := r.Remotes()
	if err != nil || len(remotes) == 0 {
		return "", fmt.Errorf("could not get git remotes: %w", err)
	}

	// get URL from git info if not provided in the project
	repo, err = fetchRepoURLFromRemotes(remotes, gitRemote)
	if err != nil {
		return "", err
	}
	return repo, nil
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

func fetchRepoURLFromRemotes(gitRemotes []*git.Remote, remoteName string) (string, error) {
	remote := &git.Remote{}
	for _, r := range gitRemotes {
		if r.Config().Name == remoteName {
			remote = r
			break
		}
	}

	if remote.Config() != nil {
		// get remote URL and convert to HTTPS in case it is an SSH URL
		httpURL := remote.Config().URLs[0]
		if strings.HasPrefix(httpURL, "git") {
			httpURL = strings.Replace(httpURL, ":", "/", 1)
			httpURL = strings.Replace(httpURL, "git@", "https://", 1)
			httpURL = strings.TrimSuffix(httpURL, gitFolder)
		}
		repoURL, err := url.Parse(httpURL)
		if err != nil {
			return "", fmt.Errorf("could not parse repository URL %q: %w", httpURL, err)
		}
		return repoURL.String(), nil
	}
	return "", fmt.Errorf("could not find git remote in: %s", remoteName)
}
