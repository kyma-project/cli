package fake

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
)

type ModuleTemplatesRepository struct {
	ListCoreResult              []*entities.CoreModuleTemplate
	ListCoreError               error
	ListLocalCommunityResult    []*entities.CommunityModuleTemplate
	ListLocalCommunityError     error
	ListExternalCommunityResult []*entities.CommunityModuleTemplate
	ListExternalCommunityError  error
}

func (m *ModuleTemplatesRepository) ListCore(ctx context.Context) ([]*entities.CoreModuleTemplate, error) {
	return m.ListCoreResult, m.ListCoreError
}

func (m *ModuleTemplatesRepository) ListLocalCommunity(ctx context.Context) ([]*entities.CommunityModuleTemplate, error) {
	return m.ListLocalCommunityResult, m.ListLocalCommunityError
}

func (m *ModuleTemplatesRepository) ListExternalCommunity(ctx context.Context, urls []string) ([]*entities.CommunityModuleTemplate, error) {
	return m.ListExternalCommunityResult, m.ListExternalCommunityError
}
