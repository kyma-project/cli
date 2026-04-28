package repository_test

import (
	"context"
	"testing"

	kubefake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
	"github.com/stretchr/testify/require"
)

func TestInstalledModulesRepository_ListInstalledModules(t *testing.T) {
	kymaClient := &kubefake.KymaClient{
		ReturnDefaultKyma: kyma.Kyma{
			Status: kyma.KymaStatus{
				Modules: []kyma.ModuleStatus{
					{Name: "api-gateway"},
					{Name: "istio"},
				},
			},
		},
	}
	repo := repository.NewInstalledModulesRepository(kymaClient)

	result, err := repo.ListInstalledModules(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, "api-gateway", result[0].Name)
	require.Equal(t, "istio", result[1].Name)
}
