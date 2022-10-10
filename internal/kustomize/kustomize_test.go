package kustomize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseKustomization(t *testing.T) {
	// URL
	k, err := ParseKustomization("https://my-site.com/owner/repo/sub/path/to/kustomization")
	require.NoError(t, err)
	require.Equal(t, Definition{Name: "repo", Ref: "main", Location: "https://my-site.com/owner/repo/sub/path/to/kustomization"}, k)

	// Path
	k, err = ParseKustomization("/Path/to/my/repo/sub/path/kustomization")
	require.NoError(t, err)
	require.Equal(t, Definition{Name: "/Path/to/my/repo/sub/path/kustomization", Ref: "local", Location: "/Path/to/my/repo/sub/path/kustomization"}, k)

	// URL with ref
	k, err = ParseKustomization("https://my-site.com/owner/repo/sub/path/to/kustomization@branchX")
	require.NoError(t, err)
	require.Equal(t, Definition{Name: "repo", Ref: "branchX", Location: "https://my-site.com/owner/repo/sub/path/to/kustomization"}, k)

	// More than one ref
	_, err = ParseKustomization("https://my-site.com/owner/repo/sub/path/to/kustomization@branchX@branchY")
	require.Error(t, err)

	// URL not a valid repo
	_, err = ParseKustomization("https://my-site.com/owner@branchX")
	require.Error(t, err)
}
