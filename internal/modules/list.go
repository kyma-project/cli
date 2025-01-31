package modules

import (
	"context"
	"fmt"
	"strconv"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Module struct {
	Name           string
	Versions       []ModuleVersion
	InstallDetails ModuleInstallDetails
}

type Managed string

const (
	ManagedTrue  Managed = "true"
	ManagedFalse Managed = "false"
)

type ModuleInstallDetails struct {
	Version string
	Channel string
	Managed Managed
	// Possible states: https://github.com/kyma-project/lifecycle-manager/blob/main/api/shared/state.go
	State string
}

type ModuleVersion struct {
	Repository string
	Version    string
	Channel    string
}

type ModulesList []Module

// List returns list of available module on a cluster
// collects info about modules based on ModuleTemplates, ModuleReleaseMetas and the KymaCR
func List(ctx context.Context, client kube.Client) (ModulesList, error) {
	moduleTemplates, err := client.Kyma().ListModuleTemplate(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list all ModuleTemplate CRs from the cluster")
	}

	modulereleasemetas, err := client.Kyma().ListModuleReleaseMeta(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list all ModuleReleaseMeta CRs from the cluster")
	}

	defaultKyma, err := client.Kyma().GetDefaultKyma(ctx)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, errors.Wrap(err, "failed to get default Kyma CR from the cluster")
	}

	modulesList := ModulesList{}
	for _, moduleTemplate := range moduleTemplates.Items {
		moduleName := moduleTemplate.Spec.ModuleName
		if moduleName == "" {
			// ignore incompatible/corrupted ModuleTemplates
			continue
		}
		version := ModuleVersion{
			Version:    moduleTemplate.Spec.Version,
			Repository: moduleTemplate.Spec.Info.Repository,
			Channel: getAssignedChannel(
				*modulereleasemetas,
				moduleName,
				moduleTemplate.Spec.Version,
			),
		}

		moduleInstalled := isModuleInstalled(defaultKyma, moduleName)

		state := ""
		if moduleInstalled {
			// only get state of installed modules
			state, err = getModuleState(ctx, client, moduleTemplate, defaultKyma)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get module state from the %s ModuleTemplate", moduleTemplate.GetName())
			}
		}
		if i := getModuleIndex(modulesList, moduleName); i != -1 {
			// append version if module with same name is in the list
			modulesList[i].Versions = append(modulesList[i].Versions, version)
		} else {
			// otherwise create a new record in the list
			modulesList = append(modulesList, Module{
				Name:           moduleName,
				InstallDetails: getInstallDetails(defaultKyma, *modulereleasemetas, moduleName, state),
				Versions: []ModuleVersion{
					version,
				},
			})
		}
	}

	return modulesList, nil
}

func getModuleState(ctx context.Context, client kube.Client, moduleTemplate kyma.ModuleTemplate, kymaCR *kyma.Kyma) (string, error) {
	// get state from Kyma CR if it exists
	if state := getStateFromKymaCR(moduleTemplate, kymaCR); state != "" {
		return state, nil
	}

	// get state from moduleTemplate.Spec.Data if it exists
	state, err := getStateFromData(ctx, client, moduleTemplate.Spec.Data)
	if err != nil || state != "" {
		return state, err
	}

	// get state from resource described in moduleTemplate.Spec.Manager if it exists
	state, err = getResourceState(ctx, client, moduleTemplate.Spec.Manager)
	return state, err
}

func getStateFromKymaCR(moduleTemplate kyma.ModuleTemplate, kymaCR *kyma.Kyma) string {
	if kymaCR != nil {
		for _, module := range kymaCR.Status.Modules {
			if module.Name == moduleTemplate.Spec.ModuleName && module.State != "" {
				return module.State
			}
		}
	}
	return ""
}

func getStateFromData(ctx context.Context, client kube.Client, data unstructured.Unstructured) (string, error) {
	if len(data.Object) == 0 {
		return "", nil
	}
	namespace := "kyma-system"
	metadata := data.Object["metadata"].(map[string]interface{})
	if ns, ok := metadata["namespace"]; ok && ns.(string) != "" {
		namespace = metadata["namespace"].(string)
	}

	apiVersion := data.Object["apiVersion"].(string)
	kind := data.Object["kind"].(string)
	name := metadata["name"].(string)

	unstruct := generateUnstruct(apiVersion, kind, name, namespace)
	result, err := client.RootlessDynamic().Get(ctx, &unstruct)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}
	status := result.Object["status"].(map[string]interface{})
	if state, ok := status["state"]; ok {
		return state.(string), nil
	}
	return "", nil
}

func getResourceState(ctx context.Context, client kube.Client, manager *kyma.Manager) (string, error) {
	if manager == nil {
		return "", nil
	}
	namespace := "kyma-system"
	if manager.Namespace != "" {
		namespace = manager.Namespace
	}

	apiVersion := fmt.Sprintf("%s/%s", manager.Group, manager.Version)

	unstruct := generateUnstruct(apiVersion, manager.Kind, manager.Name, namespace)
	result, err := client.RootlessDynamic().Get(ctx, &unstruct)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}

	status := result.Object["status"].(map[string]interface{})
	if state, ok := status["state"]; ok {
		return state.(string), nil
	}

	if conditions, ok := status["conditions"]; ok {
		state := getStateFromConditions(conditions.([]interface{}))
		if state != "" {
			return state, nil
		}
	}
	//check if readyreplicas and wantedreplicas exist
	if readyReplicas, ok := status["readyReplicas"]; ok {
		spec := result.Object["spec"].(map[string]interface{})
		if wantedReplicas, ok := spec["replicas"]; ok {
			state := resolveStateFromReplicas(readyReplicas.(int64), wantedReplicas.(int64))
			if state != "" {
				return state, nil
			}
		}
	}

	return "", nil
}

func generateUnstruct(apiVersion, kind, name, namespace string) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
		},
	}
}

func resolveStateFromReplicas(ready, wanted int64) string {
	if ready == wanted {
		return "Ready"
	}
	if ready < wanted {
		return "Processing"
	}
	// ready > wanted
	return "Deleting"
}

func getStateFromConditions(conditions []interface{}) string {
	for _, condition := range conditions {
		conditionUnwrapped := condition.(map[string]interface{})
		if conditionUnwrapped["status"] != "True" {
			continue
		}

		conditionType := conditionUnwrapped["type"].(string)

		switch conditionType {
		case "Available":
			return "Ready"
		case "Processing", "Error", "Warning":
			return conditionType
		}
	}
	return ""
}

func isModuleInstalled(kyma *kyma.Kyma, moduleName string) bool {
	if kyma != nil {
		for _, module := range kyma.Status.Modules {
			if module.Name == moduleName {
				return true
			}
		}
	}

	// module is not installed
	return false
}

func getInstallDetails(kyma *kyma.Kyma, releaseMetas kyma.ModuleReleaseMetaList, moduleName, state string) ModuleInstallDetails {
	if kyma != nil {
		for _, module := range kyma.Status.Modules {
			if module.Name == moduleName {
				moduleVersion := module.Version
				return ModuleInstallDetails{
					Channel: getAssignedChannel(releaseMetas, module.Name, moduleVersion),
					Managed: getManaged(kyma.Spec.Modules, moduleName),
					Version: moduleVersion,
					State:   state,
				}
			}
		}
	}

	// TODO: support community modules

	// return empty struct because module is not installed
	return ModuleInstallDetails{}
}

// look for value of managed for specific moduleName
func getManaged(specModules []kyma.Module, moduleName string) Managed {
	for _, module := range specModules {
		if module.Name == moduleName {
			if module.Managed != nil {
				return Managed(strconv.FormatBool(*module.Managed))
			}
		}
	}

	return ""
}

// look for channel assigned to version with specified moduleName
func getAssignedChannel(releaseMetas kyma.ModuleReleaseMetaList, moduleName, version string) string {
	for _, releaseMeta := range releaseMetas.Items {
		if releaseMeta.Spec.ModuleName == moduleName {
			return getChannelFromAssignments(releaseMeta.Spec.Channels, version)
		}
	}
	return ""
}

func getChannelFromAssignments(assignments []kyma.ChannelVersionAssignment, version string) string {
	for _, assignment := range assignments {
		if assignment.Version == version {
			return assignment.Channel
		}
	}

	return ""
}

// return index of module with given name. if not exists return -1
func getModuleIndex(list ModulesList, name string) int {
	for i := range list {
		if list[i].Name == name {
			return i
		}
	}

	return -1
}
