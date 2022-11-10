package git

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/go-git/go-git/v5"
)

const (
	gitFolder = ".git"
)

var errNotGit = errors.New("not a git repository")

func Source(path, repo, version string) (*cdv2.Source, error) {
	u, err := url.Parse(repo)
	if err != nil {
		return nil, fmt.Errorf("could not parse repository URL %q: %w", repo, err)
	}

	pieces := strings.Split(u.Path, "/")
	repoName := pieces[len(pieces)-1]

	src := &cdv2.Source{
		IdentityObjectMeta: cdv2.IdentityObjectMeta{
			Name:    repoName,
			Version: version,
			Type:    "git",
		},
		Access: &cdv2.UnstructuredTypedObject{
			Object: map[string]interface{}{"repoUrl": repo, "type": domain(u)},
		},
	}

	// check for .git
	if gitPath, err := findGitInfo(path); err == nil {
		r, err := git.PlainOpen(gitPath)
		if err != nil {
			return src, fmt.Errorf("could not get git information from %q: %w", gitPath, err)
		}
		head, err := r.Head()
		if err != nil {
			return src, fmt.Errorf("could not get git information from %q: %w", gitPath, err)
		}

		src.Access.Object["ref"] = head.Name().String()
		src.Access.Object["commit"] = head.Hash().String()
	}

	return src, nil
}

// domain extracts the domain name (without extension) from a URL.
func domain(u *url.URL) string {
	// depending on the format the URL was passed on, the domain can be in different places.
	h := u.Hostname()
	if h == "" {
		h = u.Scheme
		if h == "" {
			h = u.Path
		}
	}
	parts := strings.Split(h, ".")

	return parts[len(parts)-2]
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
