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
	ListExternalCommunityResult []*entities.ExternalModuleTemplate
	ListExternalCommunityError  error
	GetLocalCommunityResult     *entities.CommunityModuleTemplate
	GetLocalCommunityError      error
	SaveCommunityModuleError    error
}

func (m *ModuleTemplatesRepository) ListCore(_ context.Context) ([]*entities.CoreModuleTemplate, error) {
	return m.ListCoreResult, m.ListCoreError
}

func (m *ModuleTemplatesRepository) ListLocalCommunity(_ context.Context) ([]*entities.CommunityModuleTemplate, error) {
	return m.ListLocalCommunityResult, m.ListLocalCommunityError
}

func (m *ModuleTemplatesRepository) ListExternalCommunity(_ context.Context, _ []string, _ func(*entities.ExternalModuleTemplate) bool) ([]*entities.ExternalModuleTemplate, error) {
	return m.ListExternalCommunityResult, m.ListExternalCommunityError
}

func (m *ModuleTemplatesRepository) GetLocalCommunity(_ context.Context, _, _ string) (*entities.CommunityModuleTemplate, error) {
	return m.GetLocalCommunityResult, m.GetLocalCommunityError
}

func (m *ModuleTemplatesRepository) SaveCommunityModule(_ context.Context, _ *entities.ExternalModuleTemplate) error {
	return m.SaveCommunityModuleError
}
