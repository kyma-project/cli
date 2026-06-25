package dtos_test

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
	"github.com/stretchr/testify/require"
)

func TestListResultsFromModuleInstallations(t *testing.T) {
	managed := true
	modules := []entities.ModuleInstallation{
		{
			Name:                 "api-gateway",
			Version:              "3.5.1",
			Channel:              "regular",
			ModuleState:          "Ready",
			InstallationState:    "Ready",
			Managed:              &managed,
			CustomResourcePolicy: "CreateAndDelete",
		},
	}

	results := dtos.ListResultsFromModuleInstallations(modules)

	require.Len(t, results, 1)
	result := results[0]
	require.Equal(t, "api-gateway", result.Name)
	require.Equal(t, "3.5.1", result.Version)
	require.Equal(t, "regular", result.Channel)
	require.Equal(t, "Ready", result.ModuleState)
	require.Equal(t, "Ready", result.InstallationState)
	require.True(t, result.Managed)
	require.Equal(t, "CreateAndDelete", result.CustomResourcePolicy)
}
