package communitymodules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRender(t *testing.T) {
	var testMap = moduleMap{
		"test": {
			Name:          "testName",
			Repository:    "testRepo",
			LatestVersion: "testLatest",
			Version:       "testVer",
			Channel:       "testMan",
		},
	}
	var moduleMapEmpty = moduleMap{}

	var testMapLong = moduleMap{
		"test1": {
			Name:          "testName1",
			Repository:    "testRepo1",
			LatestVersion: "testLatest1",
			Version:       "testVer1",
			Channel:       "testMan1",
		},
		"test2": {
			Name:          "testName2",
			Repository:    "testRepo2",
			LatestVersion: "testLatest2",
			Version:       "testVer2",
			Channel:       "testMan2",
		},
	}
	var testMapSort = moduleMap{
		"ghi": {
			Name:    "ghi",
			Version: "3",
		},
		"def": {
			Name:    "def",
			Version: "2",
		},
		"abc": {
			Name:    "abc",
			Version: "1",
		},
		"jkl": {
			Name:    "jkl",
			Version: "4",
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
		result := convertModuleMapToTable(testMapLong, func(r row) []string { return []string{r.Repository, r.LatestVersion, r.Channel} })
		require.ElementsMatch(t, [][]string{{"testRepo1", "testLatest1", "testMan1"}, {"testRepo2", "testLatest2", "testMan2"}}, result)
	})
	t.Run("sort names", func(t *testing.T) {
		result := convertModuleMapToTable(testMapSort, func(r row) []string { return []string{r.Name, r.Version} })
		require.Equal(t, [][]string{{"abc", "1"}, {"def", "2"}, {"ghi", "3"}, {"jkl", "4"}}, result)
	})
}
