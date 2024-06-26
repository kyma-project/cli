package communitymodules

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMergeRowMaps(t *testing.T) {
	moduleMapCatalog := moduleMap{
		"serverless": {
			Name:       "serverless",
			Repository: "github.com/kyma-project/serverless",
		},
	}
	moduleMapManaged := moduleMap{
		"serverless": {
			Name:    "serverless",
			Managed: "Managed",
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
			Managed:    "Managed",
			Version:    "1.0.0",
		},
	}

	t.Run("Create collective view", func(t *testing.T) {
		result := MergeRowMaps(moduleMapCatalog, moduleMapManaged, moduleMapInstalled)
		require.Equal(t, moduleMapCollective, result)
	})
}

func TestMergeTwoRows(t *testing.T) {
	var rowA = row{
		Name:       "serverless",
		Repository: "github.com/kyma-project/serverless",
	}
	var rowB = row{
		Name:    "serverless",
		Managed: "Managed",
	}
	var rowResult = row{
		Name:       "serverless",
		Repository: "github.com/kyma-project/serverless",
		Managed:    "Managed",
	}
	t.Run("Merge two rows", func(t *testing.T) {
		result := mergeTwoRows(rowA, rowB)
		require.Equal(t, rowResult, result)
	})
}
