package git

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/github"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compatattr"
	ocm "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	ocmv1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

const (
	gitFolder = ".git"
	Identity  = "module-sources"
)

var errNotGit = errors.New("not a git repository")

func Source(ctx cpi.Context, path, repo, version, gitRemote string) (*ocm.Source, error) {

	var ref, commit string
	// check for .git
	if gitPath, err := findGitInfo(path); err == nil {
		gitRepository, err := git.PlainOpen(gitPath)
		if err != nil {
			return nil, fmt.Errorf("could not get git information from %q: %w", gitPath, err)
		}

		// get URL from git info if not provided in the project
		if repo == "" {
			remotes, err := gitRepository.Remotes()
			if err != nil {
				return nil, fmt.Errorf("could not get git remotes for repository: %w", err)
			}

			repo, err = FetchRepoURLFromRemotes(remotes, gitRemote)
			if err != nil {
				return nil, err
			}
		}

		head, err := gitRepository.Head()
		if err != nil {
			return nil, fmt.Errorf("could not get git information from %q: %w", gitPath, err)
		}

		ref = head.Name().String()
		commit = head.Hash().String()
	}

	// without a repo URL we can't create a valid source => skipping source
	if repo == "" {
		return nil, nil
	}

	refLabel, err := ocmv1.NewLabel("git.kyma-project.io/ref", ref, ocmv1.WithVersion("v1"))
	if err != nil {
		return nil, err
	}

	var sourceType string
	if compatattr.Get(ctx) {
		sourceType = "git"
	} else {
		sourceType = github.CONSUMER_TYPE
	}

	return &ocm.Source{
		SourceMeta: ocm.SourceMeta{Type: sourceType, ElementMeta: ocm.ElementMeta{
			Name:    Identity,
			Version: version,
			Labels:  ocmv1.Labels{*refLabel},
		}},
		Access: github.New(repo, "", commit),
	}, nil
}

// findGitInfo recursively crawls a path up until a .git folder is found and returns its path.
// If no .git folder is found in the path or its parents, notGitErr is returned.
func findGitInfo(path string) (string, error) {
	if path == string(filepath.Separator) {
		return "", errNotGit
	}

	gitPath := filepath.Join(path, gitFolder)
	_, err := os.Stat(gitPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return findGitInfo(filepath.Dir(path))
		}
		return "", err
	}

	return gitPath, nil
}

func FetchRepoURLFromRemotes(gitRemotes []*git.Remote, remoteName string) (string, error) {
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
