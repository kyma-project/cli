package modulesv2

import (
	"context"
	"testing"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	modulesfake "github.com/kyma-project/cli.v3/internal/modulesv2/fake"
	"github.com/stretchr/testify/require"
)

func TestListService_Run_ReturnsEmptyWhenNoInstalledModules(t *testing.T) {
	installedModulesRepo := &modulesfake.InstalledModulesRepository{
		ListInstalledModulesResult: []kyma.ModuleStatus{},
	}
	svc := NewListService(installedModulesRepo)

	result, err := svc.Run(context.Background())

	require.NoError(t, err)
	require.Empty(t, result)
}

func TestListService_Run_ReturnsCoreModules(t *testing.T) {
	installedModulesRepo := &modulesfake.InstalledModulesRepository{
		ListInstalledModulesResult: []kyma.ModuleStatus{
			{Name: "api-gateway"},
			{Name: "istio"},
		},
	}
	svc := NewListService(installedModulesRepo)

	result, err := svc.Run(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, "api-gateway", result[0].Name)
	require.Equal(t, "istio", result[1].Name)
}
