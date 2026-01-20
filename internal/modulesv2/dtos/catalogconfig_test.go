package dtos_test

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/stretchr/testify/require"
)

func Test_NewCatalogConfigFromRemote(t *testing.T) {
	tests := []struct {
		remote         bool
		remoteUrl      []string
		expectedConfig dtos.CatalogConfig
	}{
		{
			remote:         false,
			remoteUrl:      nil,
			expectedConfig: dtos.CatalogConfig{ListKyma: true, ListCluster: true},
		},
		{
			remote:         false,
			remoteUrl:      []string{},
			expectedConfig: dtos.CatalogConfig{ListKyma: true, ListCluster: true},
		},
		{
			remote:         true,
			remoteUrl:      nil,
			expectedConfig: dtos.CatalogConfig{ExternalUrls: []string{dtos.KYMA_COMMUNITY_MODULES_REPOSITORY_URL}},
		},
		{
			remote:         true,
			remoteUrl:      []string{},
			expectedConfig: dtos.CatalogConfig{ExternalUrls: []string{dtos.KYMA_COMMUNITY_MODULES_REPOSITORY_URL}},
		},
		{
			remote:         false,
			remoteUrl:      []string{"https://external-repo.co.uk", "https://example.com"},
			expectedConfig: dtos.CatalogConfig{ExternalUrls: []string{"https://external-repo.co.uk", "https://example.com"}},
		},
		{
			remote:         true,
			remoteUrl:      []string{"https://external-repo.co.uk", "https://example.com"},
			expectedConfig: dtos.CatalogConfig{ExternalUrls: []string{dtos.KYMA_COMMUNITY_MODULES_REPOSITORY_URL, "https://external-repo.co.uk", "https://example.com"}},
		},
		{
			remote:         false,
			remoteUrl:      []string{"https://external-repo.co.uk", "https://example.com", "https://external-repo.co.uk", "https://external-repo.co.uk"},
			expectedConfig: dtos.CatalogConfig{ExternalUrls: []string{"https://external-repo.co.uk", "https://example.com"}},
		},
	}

	for _, test := range tests {
		result := dtos.NewCatalogConfigFromRemote(test.remote, test.remoteUrl)

		require.Equal(t, test.expectedConfig.ListKyma, result.ListKyma)
		require.Equal(t, test.expectedConfig.ListCluster, result.ListCluster)
		require.Equal(t, test.expectedConfig.ExternalUrls, result.ExternalUrls)
	}
}
