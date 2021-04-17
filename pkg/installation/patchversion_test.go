package installation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindLatestPatch(t *testing.T) {
	version := "1.7.0"
	versions := []string{"1.6.3", "1.7.0/config.yaml", "1.7.0", "1.7.1", "1.7.2", "1.7.3"}
	latestPatch := setToLatestPatchVersion(version, versions)
	require.Equal(t, "1.7.3", latestPatch)
}
