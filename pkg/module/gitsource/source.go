package gitsource

import (
	"github.com/pkg/errors"
)

const (
	ocmIdentityName = "module-sources"
	ocmVersion      = "v1"
	refLabel        = "git.kyma-project.io/ref"
)

var (
	errNotGit = errors.New("not a git repository")
)

type GitSource struct{}

func NewGitSource() *GitSource {
	return &GitSource{}
}
