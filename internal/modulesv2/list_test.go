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
