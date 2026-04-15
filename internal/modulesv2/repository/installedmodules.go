package repository

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
)

type InstalledModulesRepository interface {
	ListInstalledModules(ctx context.Context) ([]kyma.ModuleStatus, error)
}
