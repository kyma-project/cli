package repo

import (
	"encoding/json"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
)

const (
	ALL_COMMUNITY_MODULES_URL = "https://kyma-project.github.io/community-modules/all-modules.json"
)

type ModuleTemplatesRemoteRepository interface {
	Community() ([]kyma.ModuleTemplate, error)
}

type moduleTemplateRemoteRepo struct {
	url string
}

func (m *moduleTemplateRemoteRepo) Community() ([]kyma.ModuleTemplate, error) {
	moduleTemplatesDefinitions, err := getFileFromURL(m.url)
	if err != nil {
		return nil, fmt.Errorf("failed to get community modules definitions: %v", err)
	}

	var result []kyma.ModuleTemplate
	if err := json.Unmarshal(moduleTemplatesDefinitions, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal module template: %w", err)
	}

	return result, nil
}

func newModuleTemplatesRemoteRepo() *moduleTemplateRemoteRepo {
	return &moduleTemplateRemoteRepo{
		url: ALL_COMMUNITY_MODULES_URL,
	}
}

func newModuleTemplatesRemoteRepoWithURL(url string) *moduleTemplateRemoteRepo {
	return &moduleTemplateRemoteRepo{
		url: url,
	}
}
