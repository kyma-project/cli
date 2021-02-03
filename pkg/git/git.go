package git

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

const prPrefix = "PR-"

// CloneRevision clones the repository in the given URL to the given path and checks out the given revision.
// The clone downloads the bare minimum to only get the given revision.
// If the revision is empty, HEAD will be used.
func CloneRevision(url, path, rev string) error {
	r, err := git.PlainCloneContext(context.Background(), path, false, &git.CloneOptions{
		Depth:      0,
		URL:        url,
		NoCheckout: rev != "", // only checkout HEAD if the revision is empty
	})
	if err != nil {
		return errors.Wrap(err, "Error downloading Kyma")

	}

	if rev != "" {
		w, err := r.Worktree()
		if err != nil {
			return errors.Wrap(err, "Error getting the worktree")

		}

		err = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(rev),
		})

		if err != nil {
			return errors.Wrap(err, "Error checking out revision")

		}
	}
	return nil
}

// ResolveRevision tries to convert a pseudo-revision reference (e.g. semVer, tag, PR, master, etc...) into a revision that can be checked out.
// If source is "local", no conversion is attempted to signal that no sources need to be checked out.
func ResolveRevision(repo, rev string) (string, error) {
	switch {
	//Install the master version
	case strings.EqualFold(rev, "master"):
		return BranchHead(repo, rev)

	//Install the specific commit hash (e.g. 34edf09a)
	case isHex(rev):
		// no need for conversion
		return rev, nil

	//Install the specific version from release (ex: 1.15.1)
	case isSemVer(rev):
		// get tag commit ID
		return Tag(repo, rev)

	//Install the specific pull request (e.g. PR-9486)
	case strings.HasPrefix(rev, "PR-"):
		// get PR HEAD commit ID
		return PRHead(repo, rev)

	default:
		return "", fmt.Errorf("failed to parse the source flag. It can take one of the following: 'local', 'master', release version (e.g. 1.4.1), commit hash (e.g. 34edf09a) or installer image")
	}
}

// BranchHead finds the HEAD commit hash of the given branch in the given repository.
func BranchHead(repo, branch string) (string, error) {
	// Create the remote with repository URL
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repo},
	})

	// We can then use every Remote functions to retrieve wanted information
	refs, err := rem.List(&git.ListOptions{})
	if err != nil {
		return "", errors.Wrap(err, "could not list commits")
	}
	// Find branch and its HEAD
	for _, ref := range refs {
		if ref.Name().IsBranch() && ref.Name().Short() == branch {
			return ref.Hash().String(), nil
		}
	}
	return "", errors.Errorf("could not find HEAD of branch %s in %s", branch, repo)
}

// Tag finds the commit hash of the given tag in the given repository.
func Tag(repo, tag string) (string, error) {
	// Create the remote with repository URL
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repo},
	})

	// We can then use every Remote functions to retrieve wanted information
	refs, err := rem.List(&git.ListOptions{})
	if err != nil {
		return "", errors.Wrap(err, "could not list commits")
	}
	// Find branch and its HEAD
	for _, ref := range refs {
		if ref.Name().IsTag() && ref.Name().Short() == tag {
			return ref.Hash().String(), nil
		}
	}
	return "", errors.Errorf("could not find tag %s in %s", tag, repo)
}

// PR finds the commit hash of the HEAD of the given PR in the given repository.
func PRHead(repo, pr string) (string, error) {
	// Create the remote with repository URL
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repo},
	})

	// We can then use every Remote functions to retrieve wanted information
	refs, err := rem.List(&git.ListOptions{})
	if err != nil {
		return "", errors.Wrap(err, "could not list commits")
	}

	if strings.HasPrefix(pr, prPrefix) {
		pr = strings.TrimLeft(pr, prPrefix)
	}

	// Find branch and its HEAD
	for _, ref := range refs {
		if strings.HasPrefix(ref.Name().String(), "refs/pull") && strings.HasSuffix(ref.Name().String(), "head") && strings.Contains(ref.Name().String(), pr) {
			return ref.Hash().String(), nil
		}
	}
	return "", errors.Errorf("could not find HEAD of pull request %s in %s", pr, repo)
}

func isSemVer(s string) bool {
	_, err := semver.Parse(s)
	return err == nil
}

func isHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil && len(s) > 7
}
