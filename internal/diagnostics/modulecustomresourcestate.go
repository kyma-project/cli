package diagnostics

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	"github.com/kyma-project/cli.v3/internal/modules/repo"
	"github.com/kyma-project/cli.v3/internal/out"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ModuleCustomResourceState struct {
	ApiVersion string             `json:"apiVersion" yaml:"apiVersion"`
	Kind       string             `json:"kind" yaml:"kind"`
	State      string             `json:"state" yaml:"state"`
	Conditions []metav1.Condition `json:"conditions" yaml:"conditions"`
}

type ModuleCustomResourceStateCollector struct {
	client              kube.Client
	moduleTemplatesRepo repo.ModuleTemplatesRepository
	*out.Printer
}

type moduleTemplatesDataResource struct {
	ApiVersion string
	Kind       string
	Name       string
	Namespace  string
}

func NewModuleCustomResourceStateCollector(client kube.Client) *ModuleCustomResourceStateCollector {
	return NewModuleCustomResourceStateCollectorWithRepo(client, nil)
}

func NewModuleCustomResourceStateCollectorWithRepo(client kube.Client, repo repo.ModuleTemplatesRepository) *ModuleCustomResourceStateCollector {
	return &ModuleCustomResourceStateCollector{
		client:              client,
		moduleTemplatesRepo: repo,
		Printer:             out.Default,
	}
}

func (c *ModuleCustomResourceStateCollector) Run(ctx context.Context) []ModuleCustomResourceState {
	moduleStates := make([]ModuleCustomResourceState, 0)

	coreModules, err := c.moduleTemplatesRepo.Core(ctx)
	if err != nil {
		c.Verbosefln("Failed to list core modules: %v", err)
		coreModules = []kyma.ModuleTemplate{}
	}

	communityModules, err := c.moduleTemplatesRepo.Community(ctx)
	if err != nil {
		c.Verbosefln("Failed to list community modules: %v", err)
		communityModules = []kyma.ModuleTemplate{}
	}

	allModules := append(coreModules, communityModules...)

	resourcesAlreadyCheckedCache := make(map[string]bool, 0)

	for _, module := range allModules {
		moduleData, err := c.toModuleTemplatesData(module.Spec.Data)
		if err != nil {
			c.Verbosefln("Failed to get data resource for %s module: %v", module.Spec.ModuleName, err)
			continue
		}

		if resourcesAlreadyCheckedCache[c.dataResourceCacheKey(moduleData)] {
			continue
		}

		resourcesList, err := c.queryModulesDataResource(ctx, *moduleData)
		if err != nil {
			c.Verbosefln("Failed to get data resource for %s module: %v", module.Spec.ModuleName, err)
			continue
		}

		if len(resourcesList.Items) > 0 {
			resourcesAlreadyCheckedCache[c.dataResourceCacheKey(moduleData)] = true
		}

		nonReadyModuleStates := collectNonReadyModulesStates(resourcesList)

		if len(nonReadyModuleStates) != 0 {
			moduleStates = append(moduleStates, nonReadyModuleStates...)
		}
	}

	return moduleStates
}

func (c *ModuleCustomResourceStateCollector) dataResourceCacheKey(moduleData *moduleTemplatesDataResource) string {
	return moduleData.ApiVersion + "/" + moduleData.Kind
}

func (c *ModuleCustomResourceStateCollector) toModuleTemplatesData(data unstructured.Unstructured) (*moduleTemplatesDataResource, error) {
	apiVersion, ok := data.Object["apiVersion"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to get apiVersion from data: %v", data)
	}

	kind, ok := data.Object["kind"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to get kind from data: %v", data)
	}

	metadata, ok := data.Object["metadata"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("failed to extract metadata: %v", data)
	}

	name, ok := metadata["name"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to get name from metadata: %v", metadata)
	}

	// namespaces are often blank
	namespace, ok := metadata["namespace"].(string)
	if !ok {
		namespace = ""
	}

	return &moduleTemplatesDataResource{
		ApiVersion: apiVersion,
		Kind:       kind,
		Name:       name,
		Namespace:  namespace,
	}, nil
}

func collectNonReadyModulesStates(resourcesList *unstructured.UnstructuredList) []ModuleCustomResourceState {
	nonReadyModuleStates := make([]ModuleCustomResourceState, 0)

	for _, resource := range resourcesList.Items {
		statusMap, ok := resource.Object["status"].(map[string]any)
		if !ok {
			out.Errfln("failed to read status from resource %s", resource)
		}

		state, ok := statusMap["state"].(string)
		if !ok {
			out.Errfln("failed to read state from statusMap %s", statusMap)
		}

		if state == "Ready" {
			continue
		}

		nonReadyModuleState := ModuleCustomResourceState{
			ApiVersion: resource.GetAPIVersion(),
			Kind:       resource.GetKind(),
			State:      state,
			Conditions: extractConditionsFromStatusMap(statusMap),
		}

		nonReadyModuleStates = append(nonReadyModuleStates, nonReadyModuleState)
	}

	return nonReadyModuleStates
}

func extractConditionsFromStatusMap(statusMap map[string]any) []metav1.Condition {
	conditions := make([]metav1.Condition, 0)

	conditionsRaw, ok := statusMap["conditions"]
	if !ok {
		return conditions
	}

	conditionsSlice, ok := conditionsRaw.([]any)
	if !ok {
		return conditions
	}

	for _, condRaw := range conditionsSlice {
		condMap, ok := condRaw.(map[string]any)
		if !ok {
			continue
		}

		condition := metav1.Condition{}

		if typeStr, ok := condMap["type"].(string); ok {
			condition.Type = typeStr
		}

		if statusStr, ok := condMap["status"].(string); ok {
			condition.Status = metav1.ConditionStatus(statusStr)
		}

		if reasonStr, ok := condMap["reason"].(string); ok {
			condition.Reason = reasonStr
		}

		if messageStr, ok := condMap["message"].(string); ok {
			condition.Message = messageStr
		}

		if lastTransitionTimeRaw, ok := condMap["lastTransitionTime"].(string); ok {
			if t, err := time.Parse(time.RFC3339, lastTransitionTimeRaw); err == nil {
				metaTime := metav1.NewTime(t)
				condition.LastTransitionTime = metaTime
			}
		}

		conditions = append(conditions, condition)
	}

	return conditions
}

func (c *ModuleCustomResourceStateCollector) isResourceRegistered(apiVersion, kind string) bool {
	discoveryClient := c.client.Static().Discovery()

	resourceList, err := discoveryClient.ServerResourcesForGroupVersion(apiVersion)
	if err != nil {
		return false
	}

	for _, resource := range resourceList.APIResources {
		if resource.Kind == kind {
			return true
		}
	}

	return false
}

func (c *ModuleCustomResourceStateCollector) queryModulesDataResource(ctx context.Context, data moduleTemplatesDataResource) (*unstructured.UnstructuredList, error) {
	if !c.isResourceRegistered(data.ApiVersion, data.Kind) {
		return nil, fmt.Errorf("resource '%s' in API version '%s' is not registered on cluster", data.Kind, data.ApiVersion)
	}

	resourceList, err := c.client.RootlessDynamic().List(ctx, &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": data.ApiVersion,
			"kind":       data.Kind,
			"metadata": map[string]any{
				"name":      data.Name,
				"namespace": data.Namespace,
			},
		},
	}, &rootlessdynamic.ListOptions{
		AllNamespaces: data.Namespace == "",
	})

	if err != nil {
		return nil, err
	}

	return resourceList, nil
}
