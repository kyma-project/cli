package modulesv2

import (
	"context"
	"fmt"

	semver "github.com/Masterminds/semver/v3"
	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
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

func (s *PullService) Run(ctx context.Context, pullConfig *dtos.PullConfig) error {
	externalModule, err := s.getExternalCommunityModule(ctx, pullConfig.ModuleName, pullConfig.Version, pullConfig.RemoteRepositoryUrl)
	if err != nil {
		return fmt.Errorf("failed to get community module from remote: %v", err)
	}

	err = externalModule.SetNamespace(pullConfig.Namespace)
	if err != nil {
		return fmt.Errorf("failed to store community module in the provided namespace: %v", err)
	}

	existingModule, err := s.moduleTemplatesRepository.GetLocalCommunity(ctx, externalModule.TemplateName, pullConfig.Namespace)
	if existingModule != nil && !pullConfig.Force {
		return fmt.Errorf("failed to apply module template, '%s' template already exists in the '%s' namespace. Use `--force` flag to override it", externalModule.TemplateName, pullConfig.Namespace)
	}

	return s.moduleTemplatesRepository.SaveCommunityModule(ctx, externalModule)
}

func (s *PullService) getExternalCommunityModule(ctx context.Context, moduleName, version, remoteRepositoryUrl string) (*entities.ExternalModuleTemplate, error) {
	filterModulesByName := func(name string) func(cmt *entities.ExternalModuleTemplate) bool {
		return func(cmt *entities.ExternalModuleTemplate) bool {
			return cmt.ModuleName == name
		}
	}

	externalModules, err := s.moduleTemplatesRepository.ListExternalCommunity(
		ctx,
		[]string{remoteRepositoryUrl},
		filterModulesByName(moduleName),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to list external modules: %v", err)
	}

	if len(externalModules) == 0 {
		return nil, fmt.Errorf("community module %s does not exist in the %s repository", moduleName, remoteRepositoryUrl)
	}

	if version == "" {
		return s.findLatestExternalCommunityModule(externalModules)
	}

	return findExternalCommunityModuleWithVersion(externalModules, version)
}

func (s *PullService) findLatestExternalCommunityModule(modules []*entities.ExternalModuleTemplate) (*entities.ExternalModuleTemplate, error) {
	if len(modules) == 0 {
		return nil, fmt.Errorf("no community modules provided")
	}

	latest := modules[0]
	for _, module := range modules[1:] {
		if s.isNewerVersion(module.Version, latest.Version) {
			latest = module
		}
	}
	return latest, nil
}

func (s *PullService) isNewerVersion(newVersion, oldVersion string) bool {
	newV, err := semver.NewVersion(newVersion)
	if err != nil {
		return false
	}

	oldV, err := semver.NewVersion(oldVersion)
	if err != nil {
		return true
	}

	if newV.Prerelease() != "" {
		return false
	}

	// If old version is pre-release but version1 is stable, version1 is newer
	if oldV.Prerelease() != "" && newV.Prerelease() == "" {
		return true
	}

	return newV.GreaterThan(oldV)
}

func findExternalCommunityModuleWithVersion(communityModules []*entities.ExternalModuleTemplate, version string) (*entities.ExternalModuleTemplate, error) {
	for _, communityModule := range communityModules {
		if communityModule.Version == version {
			return communityModule, nil
		}
	}

	return nil, fmt.Errorf("community module %s:%s does not exist in remote repository", communityModules[0].ModuleName, version)
}
