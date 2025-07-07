package modules

import (
	"strings"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"gopkg.in/yaml.v3"
)

// Compatibility, remove when old moduleTemplate definitions are removed

// listOldModulesCatalog returns ModulesList with expected modules details based on given ModuleTemplateList
func listOldModulesCatalog(moduleTemplates *kyma.ModuleTemplateList) ModulesList {
	modulesList := ModulesList{}
	for _, moduleTemplate := range moduleTemplates.Items {
		componentInfo := getModuleTemplateComponentInfo(&moduleTemplate)
		version := ModuleVersion{
			Version:    componentInfo.Version,
			Repository: moduleTemplate.Spec.Info.Repository,
			Channel:    moduleTemplate.Spec.Channel,
		}

		if version.Version == "" || version.Channel == "" {
			// ignore corrupted ModuleTemplates (without version or channel)
			continue
		}

		if i := getModuleIndex(modulesList, componentInfo.Name, false); i != -1 {
			// append version if module with same name is in the list
			modulesList[i].Versions = append(modulesList[i].Versions, version)
		} else {
			// otherwise create a new record in the list
			modulesList = append(modulesList, Module{
				Name: componentInfo.Name,
				Versions: []ModuleVersion{
					version,
				},
			})
		}
	}

	return modulesList
}

type descriptor struct {
	Component component `yaml:"component,omitempty"`
}

type component struct {
	Version string `yaml:"version,omitempty"`
	Name    string `yaml:"name,omitempty"`
}

func getModuleTemplateComponentInfo(moduleTemplate *kyma.ModuleTemplate) component {
	d := descriptor{}
	err := yaml.Unmarshal(moduleTemplate.Spec.Descriptor.Raw, &d)
	if err != nil {
		// unexpected error
		return component{}
	}

	// set name to last elem
	// on cluster it looks like this:
	// kyma-project.io/module/keda
	nameElems := strings.Split(d.Component.Name, "/")
	componentName := nameElems[len(nameElems)-1]

	return component{
		Version: d.Component.Version,
		Name:    componentName,
	}
}
