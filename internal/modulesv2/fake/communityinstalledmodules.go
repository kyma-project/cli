package fake

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
)

type CommunityModuleInstallationsRepository struct {
	ListInstalledCommunityModulesResult []entities.CommunityModuleInstallation
	ListInstalledCommunityModulesError  error
}

func (f *CommunityModuleInstallationsRepository) ListInstalledCommunityModules(_ context.Context) ([]entities.CommunityModuleInstallation, error) {
	return f.ListInstalledCommunityModulesResult, f.ListInstalledCommunityModulesError
}
