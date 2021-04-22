package installation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindLatestPatch(t *testing.T) {
	version := "1.7.0"
	versions := []tagStruct{{"1.6.21", false}, {"1.7.0", false}, {"1.7.10", false}, {"1.7.11-rc1/kyma-installer-cr-cluster.yaml", true}, {"1.7.9", false}}
	latestPatch := findLatestPatchVersion(version, versions)
	require.Equal(t, "1.7.10", latestPatch)

	version = "1.7.9"
	latestPatch = findLatestPatchVersion(version, versions)
	require.Equal(t, "1.7.10", latestPatch)

	version = "1.7.10"
	latestPatch = findLatestPatchVersion(version, versions)
	require.Equal(t, "1.7.10", latestPatch)
}
