package communitymodules

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRender(t *testing.T) {
	var moduleMap = moduleMap{
		"test": {
			Name:       "testName",
			Repository: "testRepo",
			Version:    "testVer",
			Managed:    "testMan",
		},
	}
	t.Run("convertRowToCatalog", func(t *testing.T) {
		result := convertRowToCatalog(moduleMap)
		require.Equal(t, [][]string{{"testName", "testRepo"}}, result)
	})
	t.Run("convertRowToManaged", func(t *testing.T) {
		result := convertRowToManaged(moduleMap)
		require.Equal(t, [][]string{{"testName"}}, result)
	})
	t.Run("convertRowToInstalled", func(t *testing.T) {
		result := convertRowToInstalled(moduleMap)
		require.Equal(t, [][]string{{"testName", "testVer"}}, result)
	})
	t.Run("convertRowToCollective", func(t *testing.T) {
		result := convertRowToCollective(moduleMap)
		require.Equal(t, [][]string{{"testName", "testRepo", "testVer", "testMan"}}, result)
	})
}
