package fake

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
)

type ModuleTemplatesRepo struct {
	ReturnAll                                []kyma.ModuleTemplate
	ReturnCommunityByName                    []kyma.ModuleTemplate
	ReturnCommunityInstalledByName           []kyma.ModuleTemplate
	ReturnRunningAssociatedResourcesOfModule []unstructured.Unstructured
	ReturnResources                          []map[string]any
	ReturnDeleteResourceReturnWatcher        watch.Interface

	AllErr                                error
	CommunityByNameErr                    error
	CommunityInstalledByNameErr           error
	RunningAssociatedResourcesOfModuleErr error
	ResourcesErr                          error
	DeleteResourceReturnWatcherErr        error
}

func (r *ModuleTemplatesRepo) All(ctx context.Context) ([]kyma.ModuleTemplate, error) {
	return r.ReturnAll, r.AllErr
}

func (r *ModuleTemplatesRepo) CommunityByName(ctx context.Context, moduleName string) ([]kyma.ModuleTemplate, error) {
	return r.ReturnCommunityByName, r.CommunityByNameErr
}

func (r *ModuleTemplatesRepo) CommunityInstalledByName(ctx context.Context, moduleName string) ([]kyma.ModuleTemplate, error) {
	return r.ReturnCommunityInstalledByName, r.CommunityInstalledByNameErr
}

func (r *ModuleTemplatesRepo) RunningAssociatedResourcesOfModule(ctx context.Context, moduleTemplate kyma.ModuleTemplate) ([]unstructured.Unstructured, error) {
	return r.ReturnRunningAssociatedResourcesOfModule, r.RunningAssociatedResourcesOfModuleErr
}

func (r *ModuleTemplatesRepo) Resources(ctx context.Context, moduleTemplate kyma.ModuleTemplate) ([]map[string]any, error) {
	return r.ReturnResources, r.ResourcesErr
}

func (r *ModuleTemplatesRepo) DeleteResourceReturnWatcher(ctx context.Context, re map[string]any) (watch.Interface, error) {
	return r.ReturnDeleteResourceReturnWatcher, r.DeleteResourceReturnWatcherErr
}
