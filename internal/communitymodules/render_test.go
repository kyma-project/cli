package communitymodules

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRender(t *testing.T) {
	var testMap = moduleMap{
		"test": {
			Name:       "testName",
			Repository: "testRepo",
			Version:    "testVer",
			Managed:    "testMan",
		},
	}
	var moduleMapEmpty = moduleMap{}

	t.Run("RenderTable... doesn't panic for empty Map", func(t *testing.T) {
		require.NotPanics(t, func() {
			RenderTableForCatalog(false, moduleMapEmpty)
			RenderTableForManaged(false, moduleMapEmpty)
			RenderTableForInstalled(false, moduleMapEmpty)
			RenderTableForCollective(false, moduleMapEmpty)
		})
	})
	t.Run("convertRow... doesn't panic for empty Map", func(t *testing.T) {
		require.NotPanics(t, func() {
			convertRowToCollective(moduleMapEmpty)
			convertRowToCatalog(moduleMapEmpty)
			convertRowToManaged(moduleMapEmpty)
			convertRowToInstalled(moduleMapEmpty)
		})
	})
	t.Run("convertRowToCatalog", func(t *testing.T) {
		result := convertRowToCatalog(testMap)
		require.Equal(t, [][]string{{"testName", "testRepo"}}, result)
	})
	t.Run("convertRowToManaged", func(t *testing.T) {
		result := convertRowToManaged(testMap)
		require.Equal(t, [][]string{{"testName"}}, result)
	})
	t.Run("convertRowToInstalled", func(t *testing.T) {
		result := convertRowToInstalled(testMap)
		require.Equal(t, [][]string{{"testName", "testVer"}}, result)
	})
	t.Run("convertRowToCollective", func(t *testing.T) {
		result := convertRowToCollective(testMap)
		require.Equal(t, [][]string{{"testName", "testRepo", "testVer", "testMan"}}, result)
	})
}
