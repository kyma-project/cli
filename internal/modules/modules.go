package modules

import (
	"context"

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

	// assign spec to another value to operate on nil-free value
	defaultKymaSpec := kyma.KymaSpec{}
	if defaultKyma != nil {
		defaultKymaSpec = defaultKyma.Spec
	}

	modulesList := ModulesList{}
	for _, moduleTemplate := range moduleTemplates.Items {
		moduleName := moduleTemplate.Spec.ModuleName
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
				InstallDetails: getInstallDetails(defaultKymaSpec, *modulereleasemetas, moduleName),
				Versions: []ModuleVersion{
					version,
				},
			})
		}
	}

	return modulesList, nil
}

func getInstallDetails(kymaSpec kyma.KymaSpec, releaseMetas kyma.ModuleReleaseMetaList, moduleName string) ModuleInstallDetails {
	for _, module := range kymaSpec.Modules {
		if module.Name == moduleName {
			moduleChannel := kymaSpec.Channel
			if module.Channel != "" {
				moduleChannel = module.Channel
			}

			return ModuleInstallDetails{
				Channel: moduleChannel,
				Version: getAssignedVersion(releaseMetas, module.Name, moduleChannel),
				Managed: ManagedTrue,
			}
		}
	}

	// TODO: support community modules

	// return empty struct because module is not installed
	return ModuleInstallDetails{}
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

// look for version assigned to channel with specified moduleName
func getAssignedVersion(releaseMetas kyma.ModuleReleaseMetaList, moduleName, channel string) string {
	for _, releaseMeta := range releaseMetas.Items {
		if releaseMeta.Spec.ModuleName == moduleName {
			return getVersionFromAssignments(releaseMeta.Spec.Channels, channel)
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

func getVersionFromAssignments(assignments []kyma.ChannelVersionAssignment, channel string) string {
	for _, assignment := range assignments {
		if assignment.Channel == channel {
			return assignment.Version
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
