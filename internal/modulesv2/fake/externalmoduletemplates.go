package fake

import "github.com/kyma-project/cli.v3/internal/kube/kyma"

type ExternalModuleTemplatesRepository struct {
	Modules []kyma.ModuleTemplate
	Err     error
}

func (r *ExternalModuleTemplatesRepository) Get(_ []string) ([]kyma.ModuleTemplate, error) {
	return r.Modules, r.Err
}
