package fake

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
)

type InstalledModulesRepository struct {
	ListInstalledModulesResult []kyma.KymaModuleInfo
	ListInstalledModulesError  error
}

func (f *InstalledModulesRepository) ListInstalledModules(_ context.Context) ([]kyma.KymaModuleInfo, error) {
	return f.ListInstalledModulesResult, f.ListInstalledModulesError
}
