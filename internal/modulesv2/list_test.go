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
		ListInstalledModulesResult: []kyma.KymaModuleInfo{},
	}
	svc := NewListService(installedModulesRepo)

	result, err := svc.Run(context.Background())

	require.NoError(t, err)
	require.Empty(t, result)
}

func TestListService_Run_ReturnsCoreModules(t *testing.T) {
	installedModulesRepo := &modulesfake.InstalledModulesRepository{
		ListInstalledModulesResult: []kyma.KymaModuleInfo{
			{Status: kyma.ModuleStatus{Name: "api-gateway"}},
			{Status: kyma.ModuleStatus{Name: "istio"}},
		},
	}
	svc := NewListService(installedModulesRepo)

	result, err := svc.Run(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, "api-gateway", result[0].Name)
	require.Equal(t, "istio", result[1].Name)
}

func TestListService_Run_ReturnsCoreModulesWithVersionAndChannel(t *testing.T) {
	installedModulesRepo := &modulesfake.InstalledModulesRepository{
		ListInstalledModulesResult: []kyma.KymaModuleInfo{
			{Status: kyma.ModuleStatus{Name: "api-gateway", Version: "3.5.1", Channel: "regular", State: "Ready"}},
		},
	}
	svc := NewListService(installedModulesRepo)

	result, err := svc.Run(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 1)
	module := result[0]
	require.Equal(t, "api-gateway", module.Name)
	require.Equal(t, "3.5.1", module.Version)
	require.Equal(t, "regular", module.Channel)
	require.Equal(t, "Ready", module.State)
}
