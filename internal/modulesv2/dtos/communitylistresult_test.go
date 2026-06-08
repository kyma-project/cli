package dtos_test

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/stretchr/testify/require"
)

func TestCommunityListResult_HasRequiredFields(t *testing.T) {
	result := dtos.CommunityListResult{
		Name:               "default/docker-registry",
		Version:            "0.10.0",
		ModuleState:        "Ready",
		InstallationState:  "Ready",
	}

	require.Equal(t, "default/docker-registry", result.Name)
	require.Equal(t, "0.10.0", result.Version)
	require.Equal(t, "Ready", result.ModuleState)
	require.Equal(t, "Ready", result.InstallationState)
}
