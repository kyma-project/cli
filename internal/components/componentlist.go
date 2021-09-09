package components

import (
"encoding/json"
"fmt"
	"github.com/kyma-incubator/reconciler/pkg/keb"
	"io/ioutil"
"os"
"path/filepath"
	"strings"

	"github.com/pkg/errors"
"gopkg.in/yaml.v3"
)

const defaultNamespace = "kyma-system"

// ComponentList collects component definitions
type ComponentList struct {
	Prerequisites []keb.Components
	Components    []keb.Components
}

// ComponentDefinition defines a component in components list
type ComponentDefinition struct {
	Name      string
	Namespace string
}

// ComponentListData is the raw component list
type ComponentListData struct {
	DefaultNamespace string `yaml:"defaultNamespace" json:"defaultNamespace"`
	Prerequisites    []ComponentDefinition
	Components       []ComponentDefinition
}

func (cld *ComponentListData) createKebComp(compDef ComponentDefinition) keb.Components{
	var c keb.Components
	if compDef.Namespace == "" {
		c.Namespace = cld.DefaultNamespace
	} else {
		c.Namespace = compDef.Namespace
	}
	c.Component = compDef.Name
	return  c
}

func applyOverrides(compList []keb.Components, overrides map[string]string) []keb.Components {
	for i, c := range compList {
		for k,v := range overrides {
			overrideComponent := strings.Split(k, ".")[0]
			if overrideComponent == c.Component || overrideComponent == "global" {
				c.Configuration = append(c.Configuration, keb.Configuration{Key: k, Value: v})
			}
		}
		compList[i] = c
	}
	return  compList
}

func (cld *ComponentListData) process(overrides map[string]string) []keb.Components {
	var compList []keb.Components
	var preReqs []keb.Components
	var comps []keb.Components

	// read prerequisites
	for _, compDef := range cld.Prerequisites {
		preReqs = append(preReqs, cld.createKebComp(compDef))
	}

	// read components
	for _, compDef := range cld.Components {
		comps = append(comps, cld.createKebComp(compDef))
	}
	compList = append(compList, preReqs...)
	compList = append(compList, comps...)

	return applyOverrides(compList, overrides)
}

func ComponentsFromStrings(list []string, overrides map[string]string) []keb.Components {
	var components []keb.Components
	for _, item := range list {
		s := strings.Split(item, "@")

		component := keb.Components{Component: s[0], Namespace: s[1]}
		components = append(components, component)
	}
	return applyOverrides(components, overrides)
}

// NewComponentList creates a new component list
func NewComponentList(componentsListPath string, overrides map[string]string) ([]keb.Components, error) {
	if componentsListPath == "" {
		return nil, fmt.Errorf("Path to components list file is required")
	}
	if _, err := os.Stat(componentsListPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Components list file '%s' not found", componentsListPath)
	}

	data, err := ioutil.ReadFile(componentsListPath)
	if err != nil {
		return nil, err
	}

	var compListData *ComponentListData = &ComponentListData{
		DefaultNamespace: defaultNamespace,
	}
	fileExt := filepath.Ext(componentsListPath)
	if fileExt == ".json" {
		if err := json.Unmarshal(data, &compListData); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Failed to process components file '%s'", componentsListPath))
		}
	} else if fileExt == ".yaml" || fileExt == ".yml" {
		if err := yaml.Unmarshal(data, &compListData); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Failed to process components file '%s'", componentsListPath))
		}
	} else {
		return nil, fmt.Errorf("File extension '%s' is not supported for component list files", fileExt)
	}

	return compListData.process(overrides), nil
}

