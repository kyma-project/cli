package fake

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
)

type InstalledModulesRepository struct {
	ListInstalledModulesResult []entities.ModuleInstallation
	ListInstalledModulesError  error
}

func (f *InstalledModulesRepository) ListInstalledModules(_ context.Context) ([]entities.ModuleInstallation, error) {
	return f.ListInstalledModulesResult, f.ListInstalledModulesError
}
