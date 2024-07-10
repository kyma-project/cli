package communitymodules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergeRowMaps(t *testing.T) {
	moduleMapCatalog := moduleMap{
		"serverless": {
			Name:       "serverless",
			Repository: "github.com/kyma-project/serverless",
		},
	}
	moduleMapCatalogWith := moduleMap{
		"serverless": {
			Name:       "serverless",
			Repository: "github.com/kyma-project/serverless",
		},
		"istio": {
			Name:       "istio",
			Repository: "github.com/kyma-project/istio",
		},
	}
	moduleMapManaged := moduleMap{
		"serverless": {
			Name:    "serverless",
			Channel: "Managed",
		},
	}
	moduleMapInstalled := moduleMap{
		"serverless": {
			Name:    "serverless",
			Version: "1.0.0",
		},
	}
	moduleMapCollective := moduleMap{
		"serverless": {
			Name:       "serverless",
			Repository: "github.com/kyma-project/serverless",
			Channel:    "Managed",
			Version:    "1.0.0",
		},
	}

	onlyIstio := moduleMap{
		"istio": {
			Name:       "istio",
			Repository: "github.com/kyma-project/istio",
		},
	}

	moduleMapCollectiveWith := moduleMap{
		"serverless": {
			Name:       "serverless",
			Repository: "github.com/kyma-project/serverless",
			Channel:    "Managed",
			Version:    "1.0.0",
		},
		"istio": {
			Name:       "istio",
			Repository: "github.com/kyma-project/istio",
			Channel:    "",
			Version:    "",
		},
	}

	t.Run("Create collective view", func(t *testing.T) {
		result := MergeRowMaps(moduleMapCatalog, moduleMapManaged, moduleMapInstalled)
		require.Equal(t, moduleMapCollective, result)
	})
	t.Run("Create collective view with additional catalog entry", func(t *testing.T) {
		result := MergeRowMaps(moduleMapCatalogWith, moduleMapManaged, moduleMapInstalled)
		require.Equal(t, moduleMapCollectiveWith, result)
	})
	t.Run("Create collective view with a map that has a single entry with diffrent key", func(t *testing.T) {
		result := MergeRowMaps(moduleMapCatalog, moduleMapManaged, moduleMapInstalled, onlyIstio)
		require.Equal(t, moduleMapCollectiveWith, result)
	})
}

func TestMergeTwoRows(t *testing.T) {
	var rowA = row{
		Name:       "serverless",
		Repository: "github.com/kyma-project/serverless",
	}
	var rowB = row{
		Name:    "serverless",
		Channel: "Managed",
	}
	var rowResult = row{
		Name:       "serverless",
		Repository: "github.com/kyma-project/serverless",
		Channel:    "Managed",
	}
	var rowC = row{
		Name:       "serverless",
		Repository: "github.com/kyma-project/test",
	}
	t.Run("Merge two rows", func(t *testing.T) {
		result := mergeTwoRows(rowA, rowB)
		require.Equal(t, rowResult, result)
	})
	t.Run("Merge two rows with different repository", func(t *testing.T) {
		result := mergeTwoRows(rowA, rowC)
		require.Equal(t, rowA, result)
	})
}
