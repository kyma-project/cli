package modules

import (
	"context"
	"fmt"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strconv"
)

type Module struct {
	Name           string
	Versions       []ModuleVersion
	InstallDetails ModuleInstallDetails
	Healthy        Healthy // #TODO Failsafe - remove when all modules are updated
}

type Managed string
type Healthy string

const (
	ManagedTrue    Managed = "true"
	ManagedFalse   Managed = "false"
	HealthyTrue    Healthy = "true"
	HealthyFalse   Healthy = "false"
	HealthyUnknown Healthy = "unknown" // #TODO Change to empty message
)

type ModuleInstallDetails struct {
	Version string
	Channel string
	Managed Managed
}

type ModuleVersion struct {
	Repository string
	Version    string
	Channel    string
}

type ModulesList []Module

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

		health, _ := getStateOrStatus(ctx, client, moduleName, moduleTemplate, modulereleasemetas, defaultKyma)

		if i := getModuleIndex(modulesList, moduleName); i != -1 {
			// append version if module with same name is in the list
			modulesList[i].Versions = append(modulesList[i].Versions, version)
		} else {
			// otherwise create a new record in the list
			modulesList = append(modulesList, Module{
				Name:           moduleName,
				InstallDetails: getInstallDetails(defaultKyma, *modulereleasemetas, moduleName),
				Versions: []ModuleVersion{
					version,
				},
				Healthy: health,
			})
		}
	}

	return modulesList, nil
}

func getStateOrStatus(ctx context.Context, client kube.Client, moduleName string, moduleTemplate kyma.ModuleTemplate, modulereleasemetas *kyma.ModuleReleaseMetaList, kymaCR *kyma.Kyma) (Healthy, error) {
	for _, module := range kymaCR.Status.Modules {
		if module.Name == moduleName {
			if module.State == "Ready" {
				return HealthyTrue, nil
			}
			if module.State == "" {
				break
			}
			return HealthyFalse, nil
		}
	}
	if moduleTemplate.Spec.Data != nil {
		//namespace := "kyma-system"
		//if moduleTemplate.Spec.Data.Metadata.Namespace != "" {
		//	namespace = moduleTemplate.Spec.Data.Metadata.Namespace
		//}

		unstruct := unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": moduleTemplate.Spec.Data.ApiVersion,
				"kind":       moduleTemplate.Spec.Data.Kind,
				"metadata": map[string]interface{}{
					"name":      moduleName,
					"namespace": "kyma-system",
				},
			},
		}

		result, err := client.RootlessDynamic().Get(ctx, &unstruct)
		if err != nil {
			fmt.Println(err)
			return HealthyUnknown, err
		}

		//groupVersion := strings.Split(moduleTemplate.Spec.Data.ApiVersion, "/")
		//unstructuredModule, err := client.Dynamic().Resource(schema.GroupVersionResource{
		//	Group:    groupVersion[0],
		//	Version:  groupVersion[1],
		//	Resource: "samples",
		//}).Namespace(namespace).Get(ctx, moduleTemplate.Spec.Data.Metadata.Name, metav1.GetOptions{})
		//if err != nil {
		//	fmt.Println(err)
		//	return HealthyUnknown, err
		//}
		state := result.Object["status"].(map[string]interface{})["state"].(string)
		fmt.Printf("Module %s state: %s", moduleName, state)
	}
	return HealthyUnknown, nil
}

func getInstallDetails(kyma *kyma.Kyma, releaseMetas kyma.ModuleReleaseMetaList, moduleName string) ModuleInstallDetails {
	if kyma != nil {
		for _, module := range kyma.Status.Modules {
			if module.Name == moduleName {
				moduleVersion := module.Version
				return ModuleInstallDetails{
					Channel: getAssignedChannel(releaseMetas, module.Name, moduleVersion),
					Managed: getManaged(kyma.Spec.Modules, moduleName),
					Version: moduleVersion,
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
