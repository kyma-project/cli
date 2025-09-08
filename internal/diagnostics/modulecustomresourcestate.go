package diagnostics

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	"github.com/kyma-project/cli.v3/internal/modules/repo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ModuleCustomResourceState struct {
	ApiVersion string
	Kind       string
	State      string
	Conditions []metav1.Condition
}

type ModuleCustomResourceStateCollector struct {
	client              kube.Client
	moduleTemplatesRepo repo.ModuleTemplatesRepository
	writer              io.Writer
	verbose             bool
}

func NewModuleCustomResourceStateCollector(client kube.Client, writer io.Writer, verbose bool) *ModuleCustomResourceStateCollector {
	return &ModuleCustomResourceStateCollector{
		client:              client,
		moduleTemplatesRepo: repo.NewModuleTemplatesRepo(client),
		writer:              writer,
		verbose:             verbose,
	}
}

func NewModuleCustomResourceStateCollectorWithRepo(client kube.Client, repo repo.ModuleTemplatesRepository, writer io.Writer, verbose bool) *ModuleCustomResourceStateCollector {
	return &ModuleCustomResourceStateCollector{
		client:              client,
		moduleTemplatesRepo: repo,
		writer:              writer,
		verbose:             verbose,
	}
}

func (c *ModuleCustomResourceStateCollector) Run(ctx context.Context) []ModuleCustomResourceState {
	moduleStates := make([]ModuleCustomResourceState, 0)

	installedCoreModules, err := c.moduleTemplatesRepo.CoreInstalled(ctx)
	if err != nil {
		c.WriteVerboseError(err, "Failed to list core modules")
		installedCoreModules = []kyma.ModuleTemplate{}
	}

	installedCommunityModules, err := c.moduleTemplatesRepo.CommunityInstalled(ctx)
	if err != nil {
		c.WriteVerboseError(err, "Failed to list community modules")
		installedCommunityModules = []kyma.ModuleTemplate{}
	}

	allInstalled := append(installedCoreModules, installedCommunityModules...)

	for _, installedModule := range allInstalled {
		resourcesList, err := c.queryModulesDataResource(ctx, installedModule.Spec.Data)
		if err != nil {
			c.WriteVerboseError(err, fmt.Sprintf("Failed to get data resource for %s module", installedModule.Spec.ModuleName))
			continue
		}

		nonReadyModuleStates := collectNonReadyModulesStates(resourcesList)

		if len(nonReadyModuleStates) != 0 {
			moduleStates = append(moduleStates, nonReadyModuleStates...)
		}
	}

	return moduleStates
}

func collectNonReadyModulesStates(resourcesList *unstructured.UnstructuredList) []ModuleCustomResourceState {
	nonReadyModuleStates := make([]ModuleCustomResourceState, 0)

	for _, resource := range resourcesList.Items {
		statusMap, ok := resource.Object["status"].(map[string]any)
		if !ok {
			fmt.Println("failed to read status from resource", resource)
		}

		state, ok := statusMap["state"].(string)
		if !ok {
			fmt.Println("failed to read state from statusMap", statusMap)
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

	// Convert the raw conditions to a slice of maps
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

		// Extract fields from the condition map
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
			// Parse the time string
			if t, err := time.Parse(time.RFC3339, lastTransitionTimeRaw); err == nil {
				metaTime := metav1.NewTime(t)
				condition.LastTransitionTime = metaTime
			}
		}

		conditions = append(conditions, condition)
	}

	return conditions
}

func (c *ModuleCustomResourceStateCollector) queryModulesDataResource(ctx context.Context, data unstructured.Unstructured) (*unstructured.UnstructuredList, error) {
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

	namespace, ok := metadata["namespace"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to get namespace from metadata: %v", metadata)
	}

	resourceList, err := c.client.RootlessDynamic().List(ctx, &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]any{
				name:      name,
				namespace: namespace,
			},
		},
	}, &rootlessdynamic.ListOptions{
		AllNamespaces: namespace == "",
	})

	if err != nil {
		return nil, err
	}

	return resourceList, nil
}

func (c *ModuleCustomResourceStateCollector) WriteVerboseError(err error, message string) {
	if !c.verbose || err == nil {
		return
	}

	fmt.Fprintf(c.writer, "%s: %s\n", message, err.Error())
}
