package repository_test

import (
	"context"
	"testing"

	kubefake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
	"github.com/stretchr/testify/require"
)

type fixedCRStateRepo struct {
	state string
}

func (f *fixedCRStateRepo) GetModuleCRState(_ context.Context, _ entities.ModuleInstallation) (string, error) {
	return f.state, nil
}

type fixedInstallationStateRepo struct {
	state string
}

func (f *fixedInstallationStateRepo) GetInstallationState(_ context.Context, _ entities.ModuleInstallation) (string, error) {
	return f.state, nil
}

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
	repo := repository.NewInstalledModulesRepository(kymaClient, &fixedCRStateRepo{state: "Ready"}, &fixedInstallationStateRepo{})

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
	repo := repository.NewInstalledModulesRepository(kymaClient, &fixedCRStateRepo{state: ""}, &fixedInstallationStateRepo{})

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
	repo := repository.NewInstalledModulesRepository(kymaClient, &fixedCRStateRepo{state: "Deleting"}, &fixedInstallationStateRepo{})

	result, err := repo.ListInstalledModules(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 1)
	module := result[0]
	require.Equal(t, "api-gateway", module.Name)
	require.Equal(t, "Deleting", module.KymaModuleState)
}

func TestInstalledModulesRepository_ListInstalledModules_SetsInstallationStateForCreateAndDelete(t *testing.T) {
	kymaClient := &kubefake.KymaClient{
		ReturnDefaultKyma: kyma.Kyma{
			Spec: kyma.KymaSpec{
				Modules: []kyma.Module{
					{Name: "api-gateway", CustomResourcePolicy: "CreateAndDelete"},
				},
			},
			Status: kyma.KymaStatus{
				Modules: []kyma.ModuleStatus{
					{Name: "api-gateway", State: "Warning"},
				},
			},
		},
	}
	repo := repository.NewInstalledModulesRepository(kymaClient, &fixedCRStateRepo{state: "Warning"}, &fixedInstallationStateRepo{state: "Ready"})

	result, err := repo.ListInstalledModules(context.Background())

	require.NoError(t, err)
	module := result[0]
	require.Equal(t, "Warning", module.InstallationState)
}
