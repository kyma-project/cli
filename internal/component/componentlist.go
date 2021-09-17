package component

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

// List collects component definitions
type List struct {
	Prerequisites []keb.Component
	Components    []keb.Component
}

// Definition defines a component in component list
type CompDefinition struct {
	Name      string
	Namespace string
}

// listData is the raw component list
type listData struct {
	DefaultNamespace string `yaml:"defaultNamespace" json:"defaultNamespace"`
	Prerequisites    []CompDefinition
	Components       []CompDefinition
}

func (cld *listData) createKebComp(compDef CompDefinition) keb.Component {
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
			dotIndex := strings.IndexAny(k, ".")
			//TODO: propagate the error if overrides are invalid
			if dotIndex <= 0 {
				continue
			}

			overrideComponent := k[:dotIndex]
			if overrideComponent == c.Component {
				c.Configuration = append(c.Configuration, keb.Configuration{Key: k[dotIndex+1:], Value: v})
			} else if overrideComponent == "global" {
				c.Configuration = append(c.Configuration, keb.Configuration{Key: k, Value: v})
			}
		}
		compList[i] = c
	}
	return compList
}

func (cld *listData) process(overrides map[string]interface{}) List {
	var compList List
	var preReqs []keb.Component
	var comps []keb.Component

	// read prerequisites
	for _, compDef := range cld.Prerequisites {
		preReqs = append(preReqs, cld.createKebComp(compDef))
	}

	// read component
	for _, compDef := range cld.Components {
		comps = append(comps, cld.createKebComp(compDef))
	}
	compList.Prerequisites = append(compList.Prerequisites, applyOverrides(preReqs, overrides)...)
	compList.Components = append(compList.Components, applyOverrides(comps, overrides)...)

	return compList
}

func FromStrings(list []string, overrides map[string]interface{}) List {
	var c List
	for _, item := range list {
		s := strings.Split(item, "@")

		component := keb.Component{Component: s[0], Namespace: s[1]}
		c.Components = append(c.Components, component)
	}
	c.Components = applyOverrides(c.Components, overrides)
	return c
}

// FromFile creates a new component list
func FromFile(componentsListPath string, overrides map[string]interface{}) (List, error) {
	if componentsListPath == "" {
		return List{}, fmt.Errorf("Path to component list file is required")
	}
	if _, err := os.Stat(componentsListPath); os.IsNotExist(err) {
		return List{}, fmt.Errorf("Components list file '%s' not found", componentsListPath)
	}

	data, err := ioutil.ReadFile(componentsListPath)
	if err != nil {
		return List{}, err
	}

	var compListData *listData = &listData{
		DefaultNamespace: defaultNamespace,
	}
	fileExt := filepath.Ext(componentsListPath)
	if fileExt == ".json" {
		if err := json.Unmarshal(data, &compListData); err != nil {
			return List{}, errors.Wrap(err, fmt.Sprintf("Failed to process component file '%s'", componentsListPath))
		}
	} else if fileExt == ".yaml" || fileExt == ".yml" {
		if err := yaml.Unmarshal(data, &compListData); err != nil {
			return List{}, errors.Wrap(err, fmt.Sprintf("Failed to process component file '%s'", componentsListPath))
		}
	} else {
		return List{}, fmt.Errorf("File extension '%s' is not supported for component list files", fileExt)
	}

	return compListData.process(overrides), nil
}
