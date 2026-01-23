package modulesv2

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
	modulesfake "github.com/kyma-project/cli.v3/internal/modulesv2/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCatalogService_Run(t *testing.T) {
	tests := []struct {
		name                        string
		catalogConfig               *dtos.CatalogConfig
		clusterManagedByKLM         bool
		listCoreResult              []*entities.CoreModuleTemplate
		listCoreError               error
		listLocalCommunityResult    []*entities.CommunityModuleTemplate
		listLocalCommunityError     error
		listExternalCommunityResult []*entities.ExternalModuleTemplate
		listExternalCommunityError  error
		expectedResults             []dtos.CatalogResult
		expectedError               bool
	}{
		{
			name: "no results in cluster not managed by KLM",
			catalogConfig: &dtos.CatalogConfig{
				ListKyma:     false,
				ListCluster:  false,
				ExternalUrls: []string{},
			},
			clusterManagedByKLM:         false,
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{},
			expectedResults:             []dtos.CatalogResult{},
			expectedError:               false,
		},
		{
			name: "no results in cluster managed by KLM",
			catalogConfig: &dtos.CatalogConfig{
				ListKyma:     false,
				ListCluster:  false,
				ExternalUrls: []string{},
			},
			clusterManagedByKLM:         true,
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{},
			expectedResults:             []dtos.CatalogResult{},
			expectedError:               false,
		},
		{
			name: "successful core modules response",
			catalogConfig: &dtos.CatalogConfig{
				ListKyma:     true,
				ListCluster:  false,
				ExternalUrls: []string{},
			},
			clusterManagedByKLM: true,
			listCoreResult: []*entities.CoreModuleTemplate{
				modulesfake.CoreModuleTemplate(&modulesfake.Params{
					ModuleName: "module1",
					Version:    "1.0.0",
					Channel:    "fast",
				}),
				modulesfake.CoreModuleTemplate(&modulesfake.Params{
					ModuleName: "module2",
					Version:    "2.0.0",
					Channel:    "regular",
				}),
			},
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{},
			expectedResults: []dtos.CatalogResult{
				{
					Name:              "module1",
					AvailableVersions: []string{"1.0.0(fast)"},
					Origin:            "kyma",
				},
				{
					Name:              "module2",
					AvailableVersions: []string{"2.0.0(regular)"},
					Origin:            "kyma",
				},
			},
			expectedError: false,
		},
		{
			name: "core modules not listed when cluster not managed by KLM",
			catalogConfig: &dtos.CatalogConfig{
				ListKyma:     true,
				ListCluster:  false,
				ExternalUrls: []string{},
			},
			clusterManagedByKLM: false,
			listCoreResult: []*entities.CoreModuleTemplate{
				modulesfake.CoreModuleTemplate(nil),
			},
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{},
			expectedResults:             []dtos.CatalogResult{},
			expectedError:               false,
		},
		{
			name: "successful local community modules response",
			catalogConfig: &dtos.CatalogConfig{
				ListKyma:     false,
				ListCluster:  true,
				ExternalUrls: []string{},
			},
			clusterManagedByKLM: false,
			listLocalCommunityResult: []*entities.CommunityModuleTemplate{
				modulesfake.CommunityModuleTemplate(&modulesfake.CommunityParams{
					ModuleName: "community-module1",
					Version:    "1.0.0",
					Namespace:  "kyma-system",
				}),
				modulesfake.CommunityModuleTemplate(&modulesfake.CommunityParams{
					ModuleName: "community-module2",
					Version:    "2.0.0",
					Namespace:  "default",
				}),
			},
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{},
			expectedResults: []dtos.CatalogResult{
				{
					Name:              "community-module1",
					AvailableVersions: []string{"1.0.0"},
					Origin:            "kyma-system/sample-community-template-1.0.0",
				},
				{
					Name:              "community-module2",
					AvailableVersions: []string{"2.0.0"},
					Origin:            "default/sample-community-template-1.0.0",
				},
			},
			expectedError: false,
		},
		{
			name: "successful external community modules response",
			catalogConfig: &dtos.CatalogConfig{
				ListKyma:     false,
				ListCluster:  false,
				ExternalUrls: []string{"https://example.com/modules.json"},
			},
			clusterManagedByKLM: false,
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{
					ModuleName: "external-module1",
					Version:    "1.0.0",
				}),
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{
					ModuleName: "external-module2",
					Version:    "2.0.0",
				}),
			},
			expectedResults: []dtos.CatalogResult{
				{
					Name:              "external-module1",
					AvailableVersions: []string{"1.0.0"},
					Origin:            "community",
				},
				{
					Name:              "external-module2",
					AvailableVersions: []string{"2.0.0"},
					Origin:            "community",
				},
			},
			expectedError: false,
		},
		{
			name: "combined kyma, cluster, and external modules",
			catalogConfig: &dtos.CatalogConfig{
				ListKyma:     true,
				ListCluster:  true,
				ExternalUrls: []string{"https://example.com/modules.json"},
			},
			clusterManagedByKLM: true,
			listCoreResult: []*entities.CoreModuleTemplate{
				modulesfake.CoreModuleTemplate(&modulesfake.Params{
					ModuleName: "core-module",
					Version:    "1.0.0",
				}),
			},
			listLocalCommunityResult: []*entities.CommunityModuleTemplate{
				modulesfake.CommunityModuleTemplate(&modulesfake.CommunityParams{
					ModuleName: "local-community",
					Version:    "1.0.0",
					Namespace:  "kyma-system",
				}),
			},
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{
					ModuleName: "external-community",
					Version:    "1.0.0",
				}),
			},
			expectedResults: []dtos.CatalogResult{
				{
					Name:              "core-module",
					AvailableVersions: []string{"1.0.0(fast)"},
					Origin:            "kyma",
				},
				{
					Name:              "local-community",
					AvailableVersions: []string{"1.0.0"},
					Origin:            "kyma-system/sample-community-template-1.0.0",
				},
				{
					Name:              "external-community",
					AvailableVersions: []string{"1.0.0"},
					Origin:            "community",
				},
			},
			expectedError: false,
		},
		{
			name: "error listing core modules",
			catalogConfig: &dtos.CatalogConfig{
				ListKyma:     true,
				ListCluster:  false,
				ExternalUrls: []string{},
			},
			clusterManagedByKLM: true,
			listCoreError:       errors.New("failed to connect to cluster"),
			expectedResults:     nil,
			expectedError:       true,
		},
		{
			name: "error listing local community modules",
			catalogConfig: &dtos.CatalogConfig{
				ListKyma:     false,
				ListCluster:  true,
				ExternalUrls: []string{},
			},
			clusterManagedByKLM:     false,
			listLocalCommunityError: errors.New("failed to list local modules"),
			expectedResults:         nil,
			expectedError:           true,
		},
		{
			name: "error listing external community modules",
			catalogConfig: &dtos.CatalogConfig{
				ListKyma:     false,
				ListCluster:  false,
				ExternalUrls: []string{"https://example.com/modules.json"},
			},
			clusterManagedByKLM:        false,
			listExternalCommunityError: errors.New("failed to fetch external modules"),
			expectedResults:            nil,
			expectedError:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockModuleRepo := &modulesfake.ModuleTemplatesRepository{
				ListCoreResult:              tt.listCoreResult,
				ListCoreError:               tt.listCoreError,
				ListLocalCommunityResult:    tt.listLocalCommunityResult,
				ListLocalCommunityError:     tt.listLocalCommunityError,
				ListExternalCommunityResult: tt.listExternalCommunityResult,
				ListExternalCommunityError:  tt.listExternalCommunityError,
			}

			mockMetadataRepo := &modulesfake.ClusterMetadataRepository{
				IsManagedByKLM: tt.clusterManagedByKLM,
			}

			service := NewCatalogService(mockModuleRepo, mockMetadataRepo)

			ctx := context.Background()
			results, err := service.Run(ctx, tt.catalogConfig)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResults, results)
			}
		})
	}
}
