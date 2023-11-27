//nolint:forcetypeassert
package mock

import (
	"context"

	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"github.com/stretchr/testify/mock"
)

type Interactor struct {
	mock.Mock
}

func (m *Interactor) Get(ctx context.Context) ([]v1beta2.Module, string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]v1beta2.Module), args.Get(1).(string), args.Error(2)
}

func (m *Interactor) Update(ctx context.Context, modules []v1beta2.Module) error {
	args := m.Called(ctx, modules)
	return args.Error(0)
}

func (m *Interactor) WaitUntilReady(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *Interactor) GetAllModuleTemplates(ctx context.Context) (v1beta2.ModuleTemplateList, error) {
	args := m.Called(ctx)
	return args.Get(0).(v1beta2.ModuleTemplateList), args.Error(1)
}
