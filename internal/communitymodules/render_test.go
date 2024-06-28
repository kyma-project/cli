package communitymodules

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRender(t *testing.T) {
	var testMap = moduleMap{
		"test": {
			Name:          "testName",
			Repository:    "testRepo",
			LatestVersion: "testLatest",
			Version:       "testVer",
			Managed:       "testMan",
		},
	}
	var moduleMapEmpty = moduleMap{}

	var testMapLong = moduleMap{
		"test1": {
			Name:          "testName1",
			Repository:    "testRepo1",
			LatestVersion: "testLatest1",
			Version:       "testVer1",
			Managed:       "testMan1",
		},
		"test2": {
			Name:          "testName2",
			Repository:    "testRepo2",
			LatestVersion: "testLatest2",
			Version:       "testVer2",
			Managed:       "testMan2",
		},
	}

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
		require.Equal(t, [][]string{{"testName", "testRepo", "testLatest"}}, result)
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
	t.Run("for map with mutliple entries", func(t *testing.T) {
		result := convertRowToCatalog(testMapLong)
		require.ElementsMatch(t, [][]string{{"testName1", "testRepo1", "testLatest1"}, {"testName2", "testRepo2", "testLatest2"}}, result)
	})
}
