package fake

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
)

type FakeEnabledModule struct {
	Name                 string
	Channel              string
	CustomResourcePolicy string
}

type KymaClient struct {
	// outputs
	ReturnErr                   error
	ReturnGetModuleInfoErr      error
	ReturnDisableModuleErr      error
	ReturnGetModuleTemplateErr  error
	ReturnWaitForModuleErr      error
	ReturnModuleReleaseMetaList kyma.ModuleReleaseMetaList
	ReturnModuleTemplateList    kyma.ModuleTemplateList
	ReturnModuleReleaseMeta     kyma.ModuleReleaseMeta
	ReturnModuleTemplate        kyma.ModuleTemplate
	ReturnDefaultKyma           kyma.Kyma
	ReturnModuleInfo            kyma.KymaModuleInfo

	// input arguments
	UpdateDefaultKymas []kyma.Kyma
	DisabledModules    []string
	EnabledModules     []FakeEnabledModule
}

func (c *KymaClient) ListModuleReleaseMeta(_ context.Context) (*kyma.ModuleReleaseMetaList, error) {
	return &c.ReturnModuleReleaseMetaList, c.ReturnErr
}

func (c *KymaClient) ListModuleTemplate(_ context.Context) (*kyma.ModuleTemplateList, error) {
	return &c.ReturnModuleTemplateList, c.ReturnErr
}

func (c *KymaClient) GetModuleReleaseMetaForModule(_ context.Context, _ string) (*kyma.ModuleReleaseMeta, error) {
	return &c.ReturnModuleReleaseMeta, c.ReturnErr
}

func (c *KymaClient) GetModuleTemplate(_ context.Context, _, _ string) (*kyma.ModuleTemplate, error) {
	return &c.ReturnModuleTemplate, c.ReturnGetModuleTemplateErr
}

func (c *KymaClient) GetModuleTemplateForModule(_ context.Context, _, _ string) (*kyma.ModuleTemplate, error) {
	return &c.ReturnModuleTemplate, c.ReturnGetModuleTemplateErr
}

func (c *KymaClient) GetDefaultKyma(_ context.Context) (*kyma.Kyma, error) {
	return &c.ReturnDefaultKyma, c.ReturnErr
}

func (c *KymaClient) UpdateDefaultKyma(_ context.Context, kyma *kyma.Kyma) error {
	c.UpdateDefaultKymas = append(c.UpdateDefaultKymas, *kyma)
	return c.ReturnErr
}

func (c *KymaClient) GetModuleInfo(_ context.Context, _ string) (*kyma.KymaModuleInfo, error) {
	return &c.ReturnModuleInfo, c.ReturnGetModuleInfoErr
}

func (c *KymaClient) WaitForModuleState(_ context.Context, _ string, _ ...string) error {
	return c.ReturnWaitForModuleErr
}

func (c *KymaClient) EnableModule(_ context.Context, module string, channel string, customResourcePolicy string) error {
	c.EnabledModules = append(c.EnabledModules, FakeEnabledModule{
		Name:                 module,
		Channel:              channel,
		CustomResourcePolicy: customResourcePolicy,
	})
	return c.ReturnErr
}

func (c *KymaClient) DisableModule(_ context.Context, module string) error {
	c.DisabledModules = append(c.DisabledModules, module)
	return c.ReturnDisableModuleErr
}

func (c *KymaClient) ManageModule(_ context.Context, _, _ string) error {
	return c.ReturnWaitForModuleErr
}

func (c *KymaClient) UnmanageModule(_ context.Context, _ string) error {
	return c.ReturnWaitForModuleErr
}
