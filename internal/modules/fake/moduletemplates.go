package fake

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
)

type ModuleTemplatesRepo struct {
	ReturnCore                               []kyma.ModuleTemplate
	ReturnCoreInstalled                      []kyma.ModuleTemplate
	ReturnCommunity                          []kyma.ModuleTemplate
	ReturnCommunityInstalled                 []kyma.ModuleTemplate
	ReturnCommunityByName                    []kyma.ModuleTemplate
	ReturnCommunityInstalledByName           []kyma.ModuleTemplate
	ReturnRunningAssociatedResourcesOfModule []unstructured.Unstructured
	ReturnResources                          []map[string]any
	ReturnInstalledManager                   *unstructured.Unstructured
	ReturnDeleteResourceReturnWatcher        watch.Interface

	CoreErr                               error
	CoreInstalledErr                      error
	CommunityErr                          error
	CommunityInstalledErr                 error
	CommunityByNameErr                    error
	CommunityInstalledByNameErr           error
	RunningAssociatedResourcesOfModuleErr error
	ResourcesErr                          error
	DeleteResourceReturnWatcherErr        error
	InstalledManagerErr                   error
}

func (r *ModuleTemplatesRepo) Core(_ context.Context) ([]kyma.ModuleTemplate, error) {
	return r.ReturnCore, r.CoreErr
}

func (r *ModuleTemplatesRepo) CoreInstalled(_ context.Context) ([]kyma.ModuleTemplate, error) {
	return r.ReturnCoreInstalled, r.CoreInstalledErr
}

func (r *ModuleTemplatesRepo) Community(_ context.Context) ([]kyma.ModuleTemplate, error) {
	return r.ReturnCommunity, r.CommunityErr
}

func (r *ModuleTemplatesRepo) CommunityInstalled(_ context.Context) ([]kyma.ModuleTemplate, error) {
	return r.ReturnCommunityInstalled, r.CommunityInstalledErr
}

func (r *ModuleTemplatesRepo) CommunityByName(_ context.Context, _ string) ([]kyma.ModuleTemplate, error) {
	return r.ReturnCommunityByName, r.CommunityByNameErr
}

func (r *ModuleTemplatesRepo) CommunityInstalledByName(_ context.Context, _ string) ([]kyma.ModuleTemplate, error) {
	return r.ReturnCommunityInstalledByName, r.CommunityInstalledByNameErr
}

func (r *ModuleTemplatesRepo) RunningAssociatedResourcesOfModule(_ context.Context, _ kyma.ModuleTemplate) ([]unstructured.Unstructured, error) {
	return r.ReturnRunningAssociatedResourcesOfModule, r.RunningAssociatedResourcesOfModuleErr
}

func (r *ModuleTemplatesRepo) Resources(_ context.Context, _ kyma.ModuleTemplate) ([]map[string]any, error) {
	return r.ReturnResources, r.ResourcesErr
}

func (r *ModuleTemplatesRepo) DeleteResourceReturnWatcher(_ context.Context, _ map[string]any) (watch.Interface, error) {
	return r.ReturnDeleteResourceReturnWatcher, r.DeleteResourceReturnWatcherErr
}

func (r *ModuleTemplatesRepo) InstalledManager(_ context.Context, _ kyma.ModuleTemplate) (*unstructured.Unstructured, error) {
	return r.ReturnInstalledManager, r.InstalledManagerErr
}
