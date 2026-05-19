package repository_test

import (
	"context"
	"testing"

	kubefake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
	"github.com/stretchr/testify/require"
)

func TestInstalledModulesRepository_ListInstalledModules_NormalCase(t *testing.T) {
	kymaClient := &kubefake.KymaClient{
		ReturnDefaultKyma: kyma.Kyma{
			Spec: kyma.KymaSpec{
				Modules: []kyma.Module{
					{Name: "api-gateway", CustomResourcePolicy: "CreateAndDelete"},
				},
			},
			Status: kyma.KymaStatus{
				Modules: []kyma.ModuleStatus{
					{Name: "api-gateway", State: "Ready"},
				},
			},
		},
	}
	repo := repository.NewInstalledModulesRepository(kymaClient)

	result, err := repo.ListInstalledModules(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 1)
	module := result[0]
	require.Equal(t, "api-gateway", module.Name)
	require.Equal(t, "Ready", module.ModuleState)
	require.Equal(t, "CreateAndDelete", module.CustomResourcePolicy)
}

func TestInstalledModulesRepository_ListInstalledModules_ModuleBeingAdded(t *testing.T) {
	kymaClient := &kubefake.KymaClient{
		ReturnDefaultKyma: kyma.Kyma{
			Spec: kyma.KymaSpec{
				Modules: []kyma.Module{
					{Name: "api-gateway", CustomResourcePolicy: "CreateAndDelete"},
				},
			},
			Status: kyma.KymaStatus{},
		},
	}
	repo := repository.NewInstalledModulesRepository(kymaClient)

	result, err := repo.ListInstalledModules(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 1)
	module := result[0]
	require.Equal(t, "api-gateway", module.Name)
	require.Equal(t, "", module.ModuleState)
	require.Equal(t, "CreateAndDelete", module.CustomResourcePolicy)
}

func TestInstalledModulesRepository_ListInstalledModules_ModuleBeingDeleted(t *testing.T) {
	kymaClient := &kubefake.KymaClient{
		ReturnDefaultKyma: kyma.Kyma{
			Spec: kyma.KymaSpec{},
			Status: kyma.KymaStatus{
				Modules: []kyma.ModuleStatus{
					{Name: "api-gateway", State: "Deleting"},
				},
			},
		},
	}
	repo := repository.NewInstalledModulesRepository(kymaClient)

	result, err := repo.ListInstalledModules(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 1)
	module := result[0]
	require.Equal(t, "api-gateway", module.Name)
	require.Equal(t, "Deleting", module.ModuleState)
}
