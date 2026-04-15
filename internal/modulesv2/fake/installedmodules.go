package fake

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
)

type InstalledModulesRepository struct {
	ListInstalledModulesResult []kyma.ModuleStatus
	ListInstalledModulesError  error
}

func (f *InstalledModulesRepository) ListInstalledModules(_ context.Context) ([]kyma.ModuleStatus, error) {
	return f.ListInstalledModulesResult, f.ListInstalledModulesError
}
