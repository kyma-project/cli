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
type Healthy string

const (
	ManagedTrue    Managed = "true"
	ManagedFalse   Managed = "false"
	HealthyTrue    Healthy = "true"
	HealthyFalse   Healthy = "false"
	HealthyUnknown Healthy = ""
)

type ModuleInstallDetails struct {
	Version string
	Channel string
	Managed Managed
	Healthy Healthy // #TODO Failsafe - remove when all modules are updated
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

		health, err := getModuleDeploymentHealth(ctx, client, moduleTemplate, defaultKyma)
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
				InstallDetails: getInstallDetails(defaultKyma, *modulereleasemetas, moduleName, health),
				Versions: []ModuleVersion{
					version,
				},
			})
		}
	}

	return modulesList, nil
}

func getModuleDeploymentHealth(ctx context.Context, client kube.Client, moduleTemplate kyma.ModuleTemplate, kymaCR *kyma.Kyma) (Healthy, error) {
	if kymaCR != nil {
		for _, module := range kymaCR.Status.Modules {
			if module.Name == moduleTemplate.Name {
				if module.State == "Ready" {
					return HealthyTrue, nil
				} else if module.State == "" {
					break
				}
				return HealthyFalse, nil
			}
		}
	}

	if moduleTemplate.Spec.Data != nil {
		data := moduleTemplate.Spec.Data
		namespace := "kyma-system"
		if data.Metadata.Namespace != "" {
			namespace = data.Metadata.Namespace
		}

		state, err := getResourceState(ctx, client, data.ApiVersion, data.Kind, namespace, data.Metadata.Name)
		if err == nil {
			fmt.Printf("Module %s state: %s\n", moduleTemplate.Name, state)
			if state == "Ready" {
				return HealthyTrue, nil
			} else {
				return HealthyFalse, nil
			}
		}
		if !errors.IsNotFound(err) {
			return HealthyUnknown, err
		}
	}

	if moduleTemplate.Spec.Manager != nil {
		manager := moduleTemplate.Spec.Manager
		namespace := "kyma-system"
		if manager.Namespace != "" {
			namespace = manager.Namespace
		}

		apiVersion := fmt.Sprintf("%s/%s", manager.Group, manager.Version)

		state, err := getResourceState(ctx, client, apiVersion, manager.Kind, namespace, manager.Name)
		if err == nil {
			if state == "Ready" {
				return HealthyTrue, nil
			} else if state == "Unknown" {
				return HealthyUnknown, nil
			} else {
				return HealthyFalse, nil
			}
		}
		if !errors.IsNotFound(err) {
			return HealthyFalse, err
		}
	}

	return HealthyUnknown, nil
}

func getResourceState(ctx context.Context, client kube.Client, apiVersion, kind, namespace, name string) (string, error) {
	unstruct := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
		},
	}

	result, err := client.RootlessDynamic().Get(ctx, &unstruct)
	if err != nil {
		return "", err
	}

	status := result.Object["status"].(map[string]interface{})
	if state, ok := status["state"]; ok {
		if state.(string) == "Ready" {
			return "Ready", nil
		}
	}
	if conditions, ok := status["conditions"]; ok {
		for _, condition := range conditions.([]interface{}) {
			conditionUnwrapped := condition.(map[string]interface{})
			if conditionUnwrapped["type"] == "Available" {
				if conditionUnwrapped["status"] == "True" {
					return "Ready", nil
				} else {
					return "NotReady", nil
				}
			}
		}
	}
	if readyReplicas, ok := status["readyReplicas"]; ok {
		if readyReplicas.(int64) > 0 {
			return "Ready", nil
		} else {
			return "NotReady", nil
		}
	}
	return "Unknown", nil
}

func getInstallDetails(kyma *kyma.Kyma, releaseMetas kyma.ModuleReleaseMetaList, moduleName string, health Healthy) ModuleInstallDetails {
	if kyma != nil {
		for _, module := range kyma.Status.Modules {
			if module.Name == moduleName {
				moduleVersion := module.Version
				return ModuleInstallDetails{
					Channel: getAssignedChannel(releaseMetas, module.Name, moduleVersion),
					Managed: getManaged(kyma.Spec.Modules, moduleName),
					Version: moduleVersion,
					Healthy: health,
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
