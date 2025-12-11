package dtos_test

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
	"github.com/stretchr/testify/require"
)

func Test_CatalogResultFromCoreModuleTemplates(t *testing.T) {
	entities := []*entities.CoreModuleTemplate{
		entities.NewCoreModuleTemplateFromParams("sample-template-0.0.3", "sample-template", "0.0.3", "experimental", "kyma-system"),
		entities.NewCoreModuleTemplateFromParams("sample-template-0.0.2", "sample-template", "0.0.2", "fast", "kyma-system"),
		entities.NewCoreModuleTemplateFromParams("sample-template-0.0.1", "sample-template", "0.0.1", "regular", "kyma-system"),
	}

	expectedCatalogResult := []dtos.CatalogResult{
		{
			Name:              "sample-template",
			AvailableVersions: []string{"0.0.3(experimental)", "0.0.2(fast)", "0.0.1(regular)"},
			Origin:            "kyma",
		},
	}

	require.Equal(t, expectedCatalogResult, dtos.CatalogResultFromCoreModuleTemplates(entities))
}

func Test_CatalogResultFromCommunityModuleTemplates(t *testing.T) {
	entities := []*entities.CommunityModuleTemplate{
		entities.NewCommunityModuleTemplate(
			entities.MapBaseModuleTemplateFromParams("local-template-0.0.1", "local-template", "0.0.1", "kyma-system"),
			"https://source.url",
			map[string]string{},
		),
		entities.NewCommunityModuleTemplate(
			entities.MapBaseModuleTemplateFromParams("local-template-0.0.2", "local-template", "0.0.2", "kyma-system"),
			"https://source.url",
			map[string]string{},
		),
		entities.NewCommunityModuleTemplate(
			entities.MapBaseModuleTemplateFromParams("community-template-0.0.1", "community-template", "0.0.1", ""),
			"https://source.url",
			map[string]string{},
		),
	}

	expectedCatalogResult := []dtos.CatalogResult{
		{
			Name:              "local-template",
			AvailableVersions: []string{"0.0.1"},
			Origin:            "kyma-system/local-template-0.0.1",
		},
		{
			Name:              "local-template",
			AvailableVersions: []string{"0.0.2"},
			Origin:            "kyma-system/local-template-0.0.2",
		},
		{
			Name:              "community-template",
			AvailableVersions: []string{"0.0.1"},
			Origin:            "community",
		},
	}

	require.Equal(t, expectedCatalogResult, dtos.CatalogResultFromCommunityModuleTemplates(entities))
}
