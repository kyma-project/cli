package dtos_test

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/stretchr/testify/require"
)

func Test_NewCatalogConfigFromOriginsList(t *testing.T) {
	tests := []struct {
		origin         []string
		expectedConfig dtos.CatalogConfig
	}{
		{
			origin:         []string{"kyma"},
			expectedConfig: dtos.CatalogConfig{ListKyma: true},
		},
		{
			origin:         []string{"cluster"},
			expectedConfig: dtos.CatalogConfig{ListCluster: true},
		},
		{
			origin:         []string{"community"},
			expectedConfig: dtos.CatalogConfig{ExternalUrls: []string{dtos.KYMA_COMMUNITY_MODULES_REPOSITORY_URL}},
		},
		{
			origin:         []string{"https://external-repo.co.uk", "https://example.com"},
			expectedConfig: dtos.CatalogConfig{ExternalUrls: []string{"https://external-repo.co.uk", "https://example.com"}},
		},
		{
			origin: []string{"kyma", "cluster", "community", "https://external-repo.co.uk", "https://example.com"},
			expectedConfig: dtos.CatalogConfig{
				ListKyma:    true,
				ListCluster: true,
				ExternalUrls: []string{
					dtos.KYMA_COMMUNITY_MODULES_REPOSITORY_URL,
					"https://external-repo.co.uk",
					"https://example.com",
				},
			},
		},
	}

	for _, test := range tests {
		result := dtos.NewCatalogConfigFromOriginsList(test.origin)

		require.Equal(t, test.expectedConfig.ListKyma, result.ListKyma)
		require.Equal(t, test.expectedConfig.ListCluster, result.ListCluster)
		require.Equal(t, test.expectedConfig.ExternalUrls, result.ExternalUrls)
	}
}
