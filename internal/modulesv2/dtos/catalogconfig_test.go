package dtos_test

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/stretchr/testify/require"
)

func Test_NewCatalogConfigFromRemote(t *testing.T) {
	tests := []struct {
		remote         flags.BoolOrStrings
		expectedConfig dtos.CatalogConfig
	}{
		{
			remote:         flags.BoolOrStrings{Enabled: false, Values: nil},
			expectedConfig: dtos.CatalogConfig{ListKyma: true, ListCluster: true},
		},
		{
			remote:         flags.BoolOrStrings{Enabled: false, Values: []string{}},
			expectedConfig: dtos.CatalogConfig{ListKyma: true, ListCluster: true},
		},
		{
			remote:         flags.BoolOrStrings{Enabled: true, Values: nil},
			expectedConfig: dtos.CatalogConfig{ExternalUrls: []string{dtos.KYMA_COMMUNITY_MODULES_REPOSITORY_URL}},
		},
		{
			remote:         flags.BoolOrStrings{Enabled: true, Values: []string{}},
			expectedConfig: dtos.CatalogConfig{ExternalUrls: []string{dtos.KYMA_COMMUNITY_MODULES_REPOSITORY_URL}},
		},
		{
			remote:         flags.BoolOrStrings{Enabled: true, Values: []string{"https://external-repo.co.uk", "https://example.com"}},
			expectedConfig: dtos.CatalogConfig{ExternalUrls: []string{"https://external-repo.co.uk", "https://example.com"}},
		},
	}

	for _, test := range tests {
		result := dtos.NewCatalogConfigFromRemote(test.remote)

		require.Equal(t, test.expectedConfig.ListKyma, result.ListKyma)
		require.Equal(t, test.expectedConfig.ListCluster, result.ListCluster)
		require.Equal(t, test.expectedConfig.ExternalUrls, result.ExternalUrls)
	}
}
