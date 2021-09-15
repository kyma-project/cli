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
	Prerequisites []keb.Component
	Components    []keb.Component
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

func (cld *ComponentListData) createKebComp(compDef ComponentDefinition) keb.Component {
	var c keb.Component
	if compDef.Namespace == "" {
		c.Namespace = cld.DefaultNamespace
	} else {
		c.Namespace = compDef.Namespace
	}
	c.Component = compDef.Name
	return c
}

func applyOverrides(compList []keb.Component, overrides map[string]interface{}) []keb.Component {
	for i, c := range compList {
		for k, v := range overrides {
			overrideComponent := strings.Split(k, ".")[0]
			if overrideComponent == c.Component || overrideComponent == "global" {
				c.Configuration = append(c.Configuration, keb.Configuration{Key: k, Value: v})
			}
		}
		compList[i] = c
	}
	return compList
}

func (cld *ComponentListData) process(overrides map[string]interface{}) ComponentList {
	var compList ComponentList
	var preReqs []keb.Component
	var comps []keb.Component

	// read prerequisites
	for _, compDef := range cld.Prerequisites {
		preReqs = append(preReqs, cld.createKebComp(compDef))
	}

	// read components
	for _, compDef := range cld.Components {
		comps = append(comps, cld.createKebComp(compDef))
	}
	compList.Prerequisites = append(compList.Prerequisites, applyOverrides(preReqs, overrides)...)
	compList.Components = append(compList.Components, applyOverrides(comps, overrides)...)

	return compList
}

func FromStrings(list []string, overrides map[string]interface{}) ComponentList {
	var c ComponentList
	for _, item := range list {
		s := strings.Split(item, "@")

		component := keb.Component{Component: s[0], Namespace: s[1]}
		c.Components = append(c.Components, component)
	}
	c.Components = applyOverrides(c.Components, overrides)
	return c
}

// NewComponentList creates a new component list
func NewComponentList(componentsListPath string, overrides map[string]interface{}) (ComponentList, error) {
	if componentsListPath == "" {
		return ComponentList{}, fmt.Errorf("Path to components list file is required")
	}
	if _, err := os.Stat(componentsListPath); os.IsNotExist(err) {
		return ComponentList{}, fmt.Errorf("Components list file '%s' not found", componentsListPath)
	}

	data, err := ioutil.ReadFile(componentsListPath)
	if err != nil {
		return ComponentList{}, err
	}

	var compListData *ComponentListData = &ComponentListData{
		DefaultNamespace: defaultNamespace,
	}
	fileExt := filepath.Ext(componentsListPath)
	if fileExt == ".json" {
		if err := json.Unmarshal(data, &compListData); err != nil {
			return ComponentList{}, errors.Wrap(err, fmt.Sprintf("Failed to process components file '%s'", componentsListPath))
		}
	} else if fileExt == ".yaml" || fileExt == ".yml" {
		if err := yaml.Unmarshal(data, &compListData); err != nil {
			return ComponentList{}, errors.Wrap(err, fmt.Sprintf("Failed to process components file '%s'", componentsListPath))
		}
	} else {
		return ComponentList{}, fmt.Errorf("File extension '%s' is not supported for component list files", fileExt)
	}

	return compListData.process(overrides), nil
}

func BuildCompList(comps []keb.Component) []string {
	var compSlice []string
	for _, c := range comps {
		compSlice = append(compSlice, c.Component)
	}
	return compSlice
}
