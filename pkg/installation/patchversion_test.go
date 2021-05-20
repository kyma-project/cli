package installation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindKymaPatch(t *testing.T) {
	cliVersion := "1.7.9"
	// KymaVer <= CliVer
	versions := []tagStruct{{"1.6.21", false}, {"1.7.0", false}, {"1.7.10", false}, {"1.7.11-rc1/kyma-installer-cr-cluster.yaml", true}, {"1.7.8", false}}
	latestPatch := findKymaPatchVersion(cliVersion, versions)
	require.Equal(t, "1.7.8", latestPatch)

	// find highest available KymaVersion for patched CLI
	cliVersion = "1.7.20"
	latestPatch = findKymaPatchVersion(cliVersion, versions)
	require.Equal(t, "1.7.10", latestPatch)

	cliVersion = "1.7.0"
	latestPatch = findKymaPatchVersion(cliVersion, versions)
	require.Equal(t, "1.7.0", latestPatch)
}
