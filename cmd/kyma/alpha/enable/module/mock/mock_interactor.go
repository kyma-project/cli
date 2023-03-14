package mock

import (
	"context"
	"github.com/kyma-project/lifecycle-manager/api/v1beta1"
	"github.com/stretchr/testify/mock"
)

type MockInteractor struct {
	mock.Mock
}

func (m *MockInteractor) Get(ctx context.Context) ([]v1beta1.Module, string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]v1beta1.Module), args.Get(1).(string), args.Error(2)
}

func (m *MockInteractor) Update(ctx context.Context, modules []v1beta1.Module) error {
	args := m.Called(ctx, modules)
	return args.Error(0)
}

func (m *MockInteractor) WaitUntilReady(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockInteractor) GetAllModuleTemplates(ctx context.Context) (v1beta1.ModuleTemplateList, error) {
	args := m.Called(ctx)
	return args.Get(0).(v1beta1.ModuleTemplateList), args.Error(1)
}
