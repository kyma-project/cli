package modules

import (
	"context"
	"strconv"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
}

type ModuleVersion struct {
	Repository string
	Version    string
	Channel    string
}

type ModulesList []Module

// List returns list of available module on a cluster
// collects info about modules based on ModuleTemplates, ModuleReleaseMetas and the KymaCR
func List(ctx context.Context, client kyma.Interface) (ModulesList, error) {
	moduleTemplates, err := client.ListModuleTemplate(ctx)
	if err != nil {
		return nil, err
	}

	modulereleasemetas, err := client.ListModuleReleaseMeta(ctx)
	if err != nil {
		return nil, err
	}

	defaultKyma, err := client.GetDefaultKyma(ctx)
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

		if i := getModuleIndex(modulesList, moduleName); i != -1 {
			// append version if module with same name is in the list
			modulesList[i].Versions = append(modulesList[i].Versions, version)
		} else {
			// otherwise create anew record in the list
			modulesList = append(modulesList, Module{
				Name:           moduleName,
				InstallDetails: getInstallDetails(defaultKyma, *modulereleasemetas, moduleName),
				Versions: []ModuleVersion{
					version,
				},
			})
		}
	}

	return modulesList, nil
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
