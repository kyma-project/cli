package repo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
)

type ModuleTemplatesRepository interface {
	All(ctx context.Context) ([]kyma.ModuleTemplate, error)
	CommunityByName(ctx context.Context, moduleName string) ([]kyma.ModuleTemplate, error)
	CommunityInstalledByName(ctx context.Context, moduleName string) ([]kyma.ModuleTemplate, error)
	RunningAssociatedResourcesOfModule(ctx context.Context, moduleTemplate kyma.ModuleTemplate) ([]unstructured.Unstructured, error)
	Resources(ctx context.Context, moduleTemplate kyma.ModuleTemplate) ([]map[string]any, error)
	DeleteResourceReturnWatcher(ctx context.Context, resource map[string]any) (watch.Interface, error)
}

type moduleTemplatesRepo struct {
	client kube.Client
}

func NewModuleTemplatesRepo(client kube.Client) *moduleTemplatesRepo {
	return &moduleTemplatesRepo{
		client: client,
	}
}

func (r *moduleTemplatesRepo) All(ctx context.Context) ([]kyma.ModuleTemplate, error) {
	moduleTemplates, err := r.client.Kyma().ListModuleTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list module templates: %v", err)
	}

	return moduleTemplates.Items, nil
}

func (r *moduleTemplatesRepo) CommunityByName(ctx context.Context, moduleName string) ([]kyma.ModuleTemplate, error) {
	allModuleTemplates, err := r.All(ctx)
	if err != nil {
		return nil, err
	}

	communityModulesWithName := []kyma.ModuleTemplate{}

	for _, moduleTemplate := range allModuleTemplates {
		if isCommunityModule(&moduleTemplate) && moduleTemplate.Spec.ModuleName == moduleName {
			communityModulesWithName = append(communityModulesWithName, moduleTemplate)
		}
	}

	return communityModulesWithName, nil
}

func (r *moduleTemplatesRepo) CommunityInstalledByName(ctx context.Context, moduleName string) ([]kyma.ModuleTemplate, error) {
	communityModulesWithName, err := r.CommunityByName(ctx, moduleName)
	if err != nil {
		return nil, err
	}

	installedModules := r.selectInstalled(ctx, communityModulesWithName)

	return installedModules, nil
}

func (r *moduleTemplatesRepo) RunningAssociatedResourcesOfModule(ctx context.Context, moduleTemplate kyma.ModuleTemplate) ([]unstructured.Unstructured, error) {
	associatedResources := moduleTemplate.Spec.AssociatedResources
	operator := moduleTemplate.Spec.Data

	var runningResources []unstructured.Unstructured

	for _, associatedResource := range associatedResources {
		associatedResourceApiVersion := associatedResource.Group + "/" + associatedResource.Version
		if associatedResource.Kind == operator.GetKind() && associatedResourceApiVersion == operator.GetAPIVersion() {
			continue
		}

		list, err := r.client.RootlessDynamic().List(ctx, &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": associatedResource.Group + "/" + associatedResource.Version,
				"kind":       associatedResource.Kind,
			},
		}, &rootlessdynamic.ListOptions{
			AllNamespaces: true,
		})
		if err != nil && !errors.IsNotFound(err) {
			fmt.Printf("failed to list resources %v: %v", associatedResource, err)
			continue
		}
		if err != nil && errors.IsNotFound(err) {
			continue
		}

		runningResources = append(runningResources, list.Items...)
	}

	return runningResources, nil
}

func (r *moduleTemplatesRepo) Resources(ctx context.Context, moduleTemplate kyma.ModuleTemplate) ([]map[string]any, error) {
	var parsedResources []map[string]any

	for _, resource := range moduleTemplate.Spec.Resources {
		resourceYamls, err := getResourceYamlStringsFromURL(resource.Link)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch resource YAMLs from %s: %w", resource.Link, err)
		}

		for _, yamlStr := range resourceYamls {
			var res map[string]any
			if err := yaml.Unmarshal([]byte(yamlStr), &res); err != nil {
				return nil, fmt.Errorf("failed to parse module resource YAML for %s:%s - %w", moduleTemplate.Spec.ModuleName, moduleTemplate.Spec.Version, err)
			}
			parsedResources = append(parsedResources, res)
		}
	}

	return parsedResources, nil
}

func (r *moduleTemplatesRepo) DeleteResourceReturnWatcher(ctx context.Context, resource map[string]any) (watch.Interface, error) {
	u := &unstructured.Unstructured{Object: resource}
	watcher, err := r.client.RootlessDynamic().WatchSingleResource(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to watch resource %s: %v", u.GetName(), err)
	}

	err = r.client.RootlessDynamic().Remove(ctx, u, false)
	if err != nil {
		return nil, fmt.Errorf("failed to remove resource %s: %v", u.GetName(), err)
	}

	return watcher, nil
}

func (r *moduleTemplatesRepo) selectInstalled(ctx context.Context, moduleTemplates []kyma.ModuleTemplate) []kyma.ModuleTemplate {
	installedModules := []kyma.ModuleTemplate{}

	for _, moduleTemplate := range moduleTemplates {
		moduleResources, err := r.Resources(ctx, moduleTemplate)
		if err != nil {
			fmt.Printf("failed to get resources for module %v: %v\n", moduleTemplate.Spec.ModuleName, err)
			continue
		}

		managerFromResources, err := getManagerFromResources(moduleTemplate, moduleResources)
		if err != nil {
			fmt.Printf("failed to retrieve manager info from %s: %v\n", moduleTemplate.Spec.ModuleName, err)
			continue
		}

		installedManager, err := r.getInstalledManager(ctx, managerFromResources)
		if err != nil {
			fmt.Printf("failed to retrieve installed manager from the cluster %v\n", err)
			continue
		}
		if installedManager == nil {
			// skip modules which moduletemplates exist but are not installed
			continue
		}

		installedModules = append(installedModules, moduleTemplate)
	}

	return installedModules
}

func (r *moduleTemplatesRepo) getInstalledManager(ctx context.Context, managerFromResources map[string]any) (*unstructured.Unstructured, error) {
	metadata, ok := managerFromResources["metadata"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("metadata not found in unstructured object")
	}

	unstructManager := generateUnstruct(
		managerFromResources["apiVersion"].(string),
		managerFromResources["kind"].(string),
		metadata["name"].(string),
		metadata["namespace"].(string),
	)

	unstructRes, err := r.client.RootlessDynamic().Get(ctx, &unstructManager)
	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("failed to get resource: %v", err)
	}

	return unstructRes, nil
}

func isCommunityModule(moduleTemplate *kyma.ModuleTemplate) bool {
	managedBy, exist := moduleTemplate.ObjectMeta.Labels["operator.kyma-project.io/managed-by"]
	return !exist || managedBy != "kyma"
}

func getManagerFromResources(moduleTemplate kyma.ModuleTemplate, moduleResources []map[string]any) (map[string]any, error) {
	managerFromSpec := moduleTemplate.Spec.Manager

	for _, moduleResource := range moduleResources {
		metadata, ok := moduleResource["metadata"].(map[string]any)
		if ok && managerFromSpec.GroupVersionKind.Kind == moduleResource["kind"] && managerFromSpec.Name == metadata["name"] {
			return moduleResource, nil
		}
	}

	return nil, fmt.Errorf("manager not found in resources")
}

func getResourceYamlStringsFromURL(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download resource from %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource body: %w", err)
	}

	return strings.Split(string(body), "---"), nil
}

func generateUnstruct(apiVersion, kind, name, namespace string) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]any{
				"name":      name,
				"namespace": namespace,
			},
		},
	}
}
