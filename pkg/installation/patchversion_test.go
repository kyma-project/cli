package installation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindLatestPatch(t *testing.T) {
	version := "1.7.0"
	versions := []string{"1.6.21", "1.7.0/config.yaml", "1.7.10", "1.7.11-rc1/kyma-installer-cr-cluster.yaml", "1.7.9"}
	latestPatch := FindLatestPatchVersion(version, versions)
	require.Equal(t, "1.7.11", latestPatch)

	version = "1.7.9"
	latestPatch = FindLatestPatchVersion(version, versions)
	require.Equal(t, "1.7.11", latestPatch)
}
