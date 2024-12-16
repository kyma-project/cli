package modules

import (
	"context"
	"fmt"
	"strconv"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"k8s.io/apimachinery/pkg/api/errors"
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
	State   string
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
		return nil, err
	}

	modulereleasemetas, err := client.Kyma().ListModuleReleaseMeta(ctx)
	if err != nil {
		return nil, err
	}

	defaultKyma, err := client.Kyma().GetDefaultKyma(ctx)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
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

		state, err := getModuleState(ctx, client, moduleTemplate, defaultKyma)
		if err != nil {
			return nil, err
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
	if kymaCR != nil {
		for _, module := range kymaCR.Status.Modules {
			if module.Name == moduleTemplate.Spec.ModuleName {
				if module.State != "" {
					return module.State, nil
				}
			}
		}
	}

	// get state from moduleTemplate.Spec.Data if it exists
	if moduleTemplate.Spec.Data != nil {
		state, err := getStateFromData(ctx, client, *moduleTemplate.Spec.Data)
		if err == nil {
			return state, nil
		}
		if !errors.IsNotFound(err) {
			return "", err
		}
	}

	// get state from resource described in moduleTemplate.Spec.Manager if it exists
	if moduleTemplate.Spec.Manager != nil {
		state, err := getResourceState(ctx, client, *moduleTemplate.Spec.Manager)
		if err == nil {
			return state, nil
		}
		if !errors.IsNotFound(err) {
			return "", err
		}
	}

	return "", nil
}

func getStateFromData(ctx context.Context, client kube.Client, data kyma.ModuleData) (string, error) {
	namespace := "kyma-system"
	if data.Metadata.Namespace != "" {
		namespace = data.Metadata.Namespace
	}

	unstruct := giveUnstruct(data.ApiVersion, data.Kind, data.Metadata.Name, namespace)
	result, err := client.RootlessDynamic().Get(ctx, &unstruct)
	if err != nil {
		return "", err
	}
	status := result.Object["status"].(map[string]interface{})
	if state, ok := status["state"]; ok {
		return state.(string), nil
	}
	return "", nil
}

func giveUnstruct(apiVersion, kind, name, namespace string) unstructured.Unstructured {
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

func getResourceState(ctx context.Context, client kube.Client, manager kyma.Manager) (string, error) {
	namespace := "kyma-system"
	if manager.Namespace != "" {
		namespace = manager.Namespace
	}

	apiVersion := fmt.Sprintf("%s/%s", manager.Group, manager.Version)

	unstruct := giveUnstruct(apiVersion, manager.Kind, manager.Name, namespace)

	result, err := client.RootlessDynamic().Get(ctx, &unstruct)
	if err != nil {
		return "", err
	}

	spec := result.Object["spec"].(map[string]interface{})
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
		if wantedReplicas, ok := spec["replicas"]; ok {
			state := resolveStateFromReplicas(readyReplicas.(int), wantedReplicas.(int))
			if state != "" {
				return state, nil
			}
		}
	}

	return "", nil
}

func resolveStateFromReplicas(ready, wanted int) string {
	if ready == wanted {
		return "Ready"
	}
	if ready < wanted {
		return "Processing"
	}
	if ready > wanted {
		return "Deleting"
	}
	return ""
}

func getStateFromConditions(conditions []interface{}) string {
	for _, condition := range conditions {
		conditionUnwrapped := condition.(map[string]interface{})
		if conditionUnwrapped["type"] == "Available" {
			if conditionUnwrapped["status"] == "True" {
				return "Ready"
			}
		}
		if conditionUnwrapped["type"] == "Progressing" {
			if conditionUnwrapped["status"] == "True" {
				return "Processing"
			}
		}
		if conditionUnwrapped["Type"] == "Error" {
			if conditionUnwrapped["status"] == "True" {
				return "Error"
			}
		}
		if conditionUnwrapped["Type"] == "Warning" {
			if conditionUnwrapped["status"] == "True" {
				return "Warning"
			}
		}
	}
	return ""
}

// Possible states
//Processing
//Deleting
//Ready
//Error
//""
//Warning
//Unmanaged

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
			return Managed(strconv.FormatBool(module.Managed))
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
