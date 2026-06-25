package fake

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
)

type ModuleInstallationsRepository struct {
	ListInstalledModulesResult          []entities.ModuleInstallation
	ListInstalledModulesError           error
	ListInstalledCommunityModulesResult []entities.CommunityModuleInstallation
	ListInstalledCommunityModulesError  error
}

func (f *ModuleInstallationsRepository) ListInstalledModules(_ context.Context) ([]entities.ModuleInstallation, error) {
	return f.ListInstalledModulesResult, f.ListInstalledModulesError
}

func (f *ModuleInstallationsRepository) ListInstalledCommunityModules(_ context.Context) ([]entities.CommunityModuleInstallation, error) {
	return f.ListInstalledCommunityModulesResult, f.ListInstalledCommunityModulesError
}
