package modules

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
)

type Module struct {
	Name     string
	Versions []ModuleVersion
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

	modulesList := ModulesList{}
	for _, moduleTemplate := range moduleTemplates.Items {
		version := ModuleVersion{
			Version:    moduleTemplate.Spec.Version,
			Repository: moduleTemplate.Spec.Info.Repository,
			Channel: getAssignedChannel(
				*modulereleasemetas,
				moduleTemplate.Spec.ModuleName,
				moduleTemplate.Spec.Version,
			),
		}

		if i := getModuleIndex(modulesList, moduleTemplate.Spec.ModuleName); i != -1 {
			// append version if module with same name is in the list
			modulesList[i].Versions = append(modulesList[i].Versions, version)
		} else {
			// otherwise create anew record in the list
			modulesList = append(modulesList, Module{
				Name: moduleTemplate.Spec.ModuleName,
				Versions: []ModuleVersion{
					version,
				},
			})
		}
	}

	return modulesList, nil
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
