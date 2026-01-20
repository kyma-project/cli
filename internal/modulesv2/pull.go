package modulesv2

import (
	"context"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
	"github.com/kyma-project/cli.v3/internal/out"
)

type PullService struct {
	moduleTemplatesRepository repository.ModuleTemplatesRepository
}

func NewPullService(
	moduleTemplateRepository repository.ModuleTemplatesRepository,
) *PullService {
	return &PullService{
		moduleTemplatesRepository: moduleTemplateRepository,
	}
}

func (s *PullService) Run(ctx context.Context) error {
	url := "https://kyma-project.github.io/community-modules/all-modules.json"
	communityModules, err := s.moduleTemplatesRepository.ListExternalCommunity(ctx, []string{url})
	if err != nil {
		return fmt.Errorf("failed to list external modules: %v", err)
	}

	// 1. validate that incoming namespace != 'kyma-system'
	// 2. ensure that there is only one version of specific module
	// 	  OR that specific version is provided
	// 3.

	s.moduleTemplatesRepository.CreateFromJSON(jsonModule, namespace, additionalAnnotations)

	out.Msgfln("%v", communityModules)

	return nil
}
