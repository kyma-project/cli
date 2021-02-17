package git

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	repo       = "https://github.com/kyma-project/kyma"
	kyma117Rev = "2292dd21453af5f4f517c1c42a1cf5413d8c461c"
)

func TestCloneRevision(t *testing.T) {
	t.Parallel()

	os.RemoveAll("./clone") //ensure clone folder does not exist
	err := CloneRevision(repo, "./clone", kyma117Rev)
	defer os.RemoveAll("./clone")

	require.NoError(t, err, "Cloning Kyma 1.17 should not error")
	_, err = os.Stat("./clone")
	require.NoError(t, err, "cloned local kyma folder should not error")
	_, err = os.Stat("./clone/resources")
	require.NoError(t, err, "cloned local charts folder should not error")
}

// TestResolveRevision tests implicitly also the commit ID resolution functions for: Branch, PR and Tag
func TestResolveRevision(t *testing.T) {
	t.Parallel()

	// master branch head
	r, err := ResolveRevision(repo, "master")
	require.NoError(t, err, "Resolving Kyma's master revision should not error")
	require.True(t, isHex(r), "The resolved master revision should be a hex string")
	// version tag
	r, err = ResolveRevision(repo, "1.15.0")
	require.NoError(t, err, "Resolving Kyma's 1.15.0 version tag should not error")
	require.True(t, isHex(r), "The resolved 1.15.0 version tag revision should be a hex string")

	// Pull Request
	r, err = ResolveRevision(repo, "PR-9999")
	require.NoError(t, err, "Resolving Kyma's Pull request head should not error")
	require.True(t, isHex(r), "The resolved Pull request head should be a hex string")

	// Bad ref
	_, err = ResolveRevision(repo, "not-a-git-ref")
	require.Error(t, err)
}
