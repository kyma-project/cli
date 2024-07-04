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

	t.Run("RenderModules doesn't panic for empty Map", func(t *testing.T) {
		require.NotPanics(t, func() {
			RenderModules(false, moduleMapEmpty, CatalogTableInfo)
		})
	})
	t.Run("convertRowToCatalog", func(t *testing.T) {
		result := convertModuleMapToTable(testMap, func(r row) []string { return []string{r.Name, r.LatestVersion, r.Version} })
		require.Equal(t, [][]string{{"testName", "testLatest", "testVer"}}, result)
	})
	t.Run("convertRowToCatalog for map with mutliple entries", func(t *testing.T) {
		result := convertModuleMapToTable(testMapLong, func(r row) []string { return []string{r.Repository, r.LatestVersion, r.Managed} })
		require.ElementsMatch(t, [][]string{{"testRepo1", "testLatest1", "testMan1"}, {"testRepo2", "testLatest2", "testMan2"}}, result)
	})
}
